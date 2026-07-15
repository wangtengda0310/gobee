# -*- coding: utf-8 -*-
"""校验 SkillUI_技能表现表.xlsx：结构化 + 导出语义行。"""
from __future__ import annotations

import argparse
import json
import re
import sys
from collections import Counter
from dataclasses import asdict, dataclass
from pathlib import Path

import pandas as pd

TYPE_ROW = 1
HEADER_ROW = 2
DATA_START_ROW = 4  # 本表仅 4 行元数据
COLOR_PREFIX = "<color=#FFFFFF00>占位</color>"
ID_RE = re.compile(r"^[A-Za-z][A-Za-z0-9_]*$")
INT_ARRAY_RE = re.compile(r"^\d+(,\d+)*$")
PET_DES_KEY_RE = re.compile(r"^(\{\d+;\d+;\d+;\d+;\d+\})+$")
AUDIO_ID_RE = re.compile(r"^[A-Za-z0-9_]+(,[A-Za-z0-9_]+)*$")
BOOL_TRUE = {"1", "true", "yes", "y"}
BOOL_FALSE = {"0", "false", "no", "n"}
REQUIRED_FIELDS = ("SkillName", "SkillText", "ShortSkillText")
INT_ARRAY_FIELDS = ("Audio", "KeyWords", "SkillTag", "RelatedSkill")


@dataclass
class Issue:
    row_id: object
    skill_name: str
    field: str
    message: str
    source: str = "structural"

    def format_line(self) -> str:
        name = self.skill_name or "-"
        rid = self.row_id if self.row_id != "" else "-"
        return f"Id={rid} | SkillName={name} | {self.field} | {self.message}"


@dataclass
class SemanticRow:
    id: str
    skill_name: str
    skill_text: str
    short_skill_text: str
    settlement_des: str
    allusion: str
    design_thought: str

    def to_dict(self) -> dict:
        return asdict(self)


def is_empty(value: object) -> bool:
    if pd.isna(value):
        return True
    text = str(value).strip()
    return text in ("", "nan", "NaN")


def is_blank_id(value: object) -> bool:
    return is_empty(value)


def parse_int(value: object) -> int | None:
    if is_empty(value):
        return None
    try:
        return int(float(str(value).strip()))
    except (TypeError, ValueError):
        return None


def parse_bool(value: object) -> bool | None:
    if is_empty(value):
        return None
    text = str(value).strip().lower()
    if text in BOOL_TRUE:
        return True
    if text in BOOL_FALSE:
        return False
    return None


def add(
    issues: list[Issue],
    row_id: object,
    skill_name: str,
    field: str,
    message: str,
) -> None:
    issues.append(Issue(row_id=row_id, skill_name=skill_name, field=field, message=message))


def load_rows(path: Path) -> pd.DataFrame:
    """读 SkillUI：跳过空/# Id；连续空 3 行后截断（公共约定）。"""
    raw = pd.read_excel(path, sheet_name="技能表现配置表|SkillUI", header=None, dtype=object)
    types = raw.iloc[TYPE_ROW].tolist()
    headers = raw.iloc[HEADER_ROW].tolist()
    width = max(len(types), len(headers), len(raw.columns))

    keep_cols: list[int] = []
    col_names: list[str] = []
    for i in range(width):
        type_val = types[i] if i < len(types) else None
        header_val = headers[i] if i < len(headers) else None
        if is_empty(type_val) and is_empty(header_val):
            continue
        if is_empty(header_val):
            # 保留无字段名但有类型的列（本表偶发）；用类型名占位
            name = str(type_val).strip() if not is_empty(type_val) else f"_col{i}"
        else:
            name = str(header_val).strip()
        keep_cols.append(i)
        col_names.append(name)

    id_idx = None
    for idx, name in enumerate(col_names):
        if name == "Id":
            id_idx = keep_cols[idx]
            break
    if id_idx is None:
        id_idx = keep_cols[0] if keep_cols else 0

    records: list[dict] = []
    blank_run = 0
    for r in range(DATA_START_ROW, len(raw)):
        raw_id = raw.iloc[r, id_idx]
        if is_blank_id(raw_id):
            blank_run += 1
            if blank_run >= 3:
                break
            continue
        blank_run = 0
        if str(raw_id).strip().startswith("#"):
            continue
        row: dict[str, object] = {}
        for col_i, name in zip(keep_cols, col_names):
            row[name] = raw.iloc[r, col_i]
        records.append(row)
    return pd.DataFrame.from_records(records)


def resolve_skill_table_path(ui_path: Path, skill_path: Path | None) -> Path:
    if skill_path is not None:
        return skill_path
    return ui_path.parent / "Skill.xlsx"


def load_skill_xref(skill_path: Path) -> tuple[set[int], set[str]]:
    """返回 (数值 Id 集合, E#ESkillId 字符串集合)。Skill 表用公共截断规则。"""
    raw = pd.read_excel(skill_path, sheet_name="技能表|Skill", header=None, dtype=object)
    headers = raw.iloc[HEADER_ROW].tolist()
    types = raw.iloc[TYPE_ROW].tolist()
    id_col = 0
    e_col: int | None = None
    width = max(len(headers), len(types), len(raw.columns))
    for i in range(width):
        header_val = headers[i] if i < len(headers) else None
        type_val = types[i] if i < len(types) else None
        if not is_empty(header_val) and str(header_val).strip() == "Id":
            id_col = i
        if not is_empty(type_val) and str(type_val).strip() == "E#ESkillId":
            e_col = i
        if is_empty(header_val) and not is_empty(type_val) and str(type_val).strip() == "E#ESkillId":
            e_col = i

    int_ids: set[int] = set()
    str_ids: set[str] = set()
    blank_run = 0
    skill_data_start = 5
    for r in range(skill_data_start, len(raw)):
        raw_id = raw.iloc[r, id_col]
        if is_blank_id(raw_id):
            blank_run += 1
            if blank_run >= 3:
                break
            continue
        blank_run = 0
        if str(raw_id).strip().startswith("#"):
            continue
        parsed = parse_int(raw_id)
        if parsed is not None:
            int_ids.add(parsed)
        if e_col is not None:
            e_val = raw.iloc[r, e_col]
            if not is_empty(e_val):
                str_ids.add(str(e_val).strip())
    return int_ids, str_ids


def validate_structural(
    path: Path,
    skill_path: Path | None = None,
) -> list[Issue]:
    data = load_rows(path)
    issues: list[Issue] = []
    seen_ids: list[str] = []

    resolved = resolve_skill_table_path(path, skill_path)
    skill_int_ids: set[int] | None = None
    skill_str_ids: set[str] | None = None
    if not resolved.is_file():
        add(
            issues,
            "",
            "",
            "Skill",
            f"缺少 Skill.xlsx，无法校验技能外联: {resolved}",
        )
    else:
        try:
            skill_int_ids, skill_str_ids = load_skill_xref(resolved)
        except Exception as exc:  # noqa: BLE001
            add(
                issues,
                "",
                "",
                "Skill",
                f"读取 Skill.xlsx 失败 ({resolved}): {exc}",
            )

    for _, row in data.iterrows():
        raw_id = row.get("Id")
        row_id = str(raw_id).strip()
        skill_name = "" if is_empty(row.get("SkillName")) else str(row.get("SkillName")).strip()
        seen_ids.append(row_id)

        if not ID_RE.match(row_id):
            add(
                issues,
                row_id,
                skill_name,
                "Id",
                f"Id 须为拼音/标识符 ^[A-Za-z][A-Za-z0-9_]*$，实际: {raw_id}",
            )

        if skill_str_ids is not None and row_id not in skill_str_ids:
            add(
                issues,
                row_id,
                skill_name,
                "Id",
                f"Id={row_id} 不在 Skill.xlsx 的 E#ESkillId 中",
            )

        for field in REQUIRED_FIELDS:
            if field not in data.columns:
                continue
            if is_empty(row.get(field)):
                add(issues, row_id, skill_name, field, f"{field} 不能为空")

        has_rel = parse_bool(row.get("HasRelation")) if "HasRelation" in data.columns else None
        if "HasRelation" in data.columns and not is_empty(row.get("HasRelation")) and has_rel is None:
            add(
                issues,
                row_id,
                skill_name,
                "HasRelation",
                f"HasRelation 须为 bool 语义，实际: {row.get('HasRelation')}",
            )

        related = row.get("RelatedSkill") if "RelatedSkill" in data.columns else None
        related_filled = not is_empty(related)

        if has_rel is True and not related_filled:
            add(
                issues,
                row_id,
                skill_name,
                "RelatedSkill",
                "HasRelation 为真时 RelatedSkill 必填",
            )
        if related_filled and has_rel is not True:
            add(
                issues,
                row_id,
                skill_name,
                "HasRelation",
                f"RelatedSkill 有值时 HasRelation 须为真，实际: {row.get('HasRelation')}",
            )

        for field in INT_ARRAY_FIELDS:
            if field not in data.columns:
                continue
            value = row.get(field)
            if is_empty(value):
                continue
            text = str(value).strip().replace(" ", "")
            if not INT_ARRAY_RE.match(text):
                add(
                    issues,
                    row_id,
                    skill_name,
                    field,
                    f"{field} 应为逗号分隔的 int 列表，实际: {value}",
                )
            elif field == "RelatedSkill" and skill_int_ids is not None:
                for part in text.split(","):
                    sid = int(part)
                    if sid not in skill_int_ids:
                        add(
                            issues,
                            row_id,
                            skill_name,
                            "RelatedSkill",
                            f"RelatedSkill Id={sid} 不在 Skill.xlsx 的 Id 中",
                        )

        for field in ("Allusion", "DesignThought"):
            if field not in data.columns:
                continue
            value = row.get(field)
            if is_empty(value):
                continue
            text = str(value).strip()
            if not text.startswith(COLOR_PREFIX):
                add(
                    issues,
                    row_id,
                    skill_name,
                    field,
                    f"{field} 有值时须以 {COLOR_PREFIX} 开头",
                )

        if "IdentityLine" in data.columns and not is_empty(row.get("IdentityLine")):
            text = str(row.get("IdentityLine")).strip().replace(" ", "")
            if not AUDIO_ID_RE.match(text):
                add(
                    issues,
                    row_id,
                    skill_name,
                    "IdentityLine",
                    f"IdentityLine 应为 EAudioId 或逗号分隔列表，实际: {row.get('IdentityLine')}",
                )

        for field in ("PlayCardAudio", "SpecialAudio", "SettlementDes", "AuraButtonDes", "BattleSkillStep"):
            if field not in data.columns:
                continue
            value = row.get(field)
            if is_empty(value):
                continue
            if not str(value).strip():
                add(issues, row_id, skill_name, field, f"{field} 不能为空白")

        if "PetDesKey" in data.columns and not is_empty(row.get("PetDesKey")):
            text = str(row.get("PetDesKey")).strip().replace(" ", "")
            if not PET_DES_KEY_RE.match(text):
                add(
                    issues,
                    row_id,
                    skill_name,
                    "PetDesKey",
                    f"PetDesKey 应为一个或多个 {{int;int;int;int;int}}，实际: {row.get('PetDesKey')}",
                )

    id_counts = Counter(seen_ids)
    for dup_id, count in id_counts.items():
        if count > 1:
            add(
                issues,
                dup_id,
                "",
                "Id",
                f"Id 重复出现 {count} 次",
            )

    return issues


def collect_semantic_rows(data: pd.DataFrame) -> list[SemanticRow]:
    rows: list[SemanticRow] = []
    for _, row in data.iterrows():
        if is_blank_id(row.get("Id")):
            continue
        rows.append(
            SemanticRow(
                id=str(row.get("Id")).strip(),
                skill_name="" if is_empty(row.get("SkillName")) else str(row.get("SkillName")).strip(),
                skill_text="" if is_empty(row.get("SkillText")) else str(row.get("SkillText")).strip(),
                short_skill_text=(
                    ""
                    if is_empty(row.get("ShortSkillText"))
                    else str(row.get("ShortSkillText")).strip()
                ),
                settlement_des=(
                    ""
                    if is_empty(row.get("SettlementDes"))
                    else str(row.get("SettlementDes")).strip()
                ),
                allusion=(
                    "" if is_empty(row.get("Allusion")) else str(row.get("Allusion")).strip()
                ),
                design_thought=(
                    ""
                    if is_empty(row.get("DesignThought"))
                    else str(row.get("DesignThought")).strip()
                ),
            )
        )
    return rows


def build_report(path: Path, skill_path: Path | None = None) -> dict:
    structural_issues = validate_structural(path, skill_path=skill_path)
    data = load_rows(path)
    semantic_rows = collect_semantic_rows(data)
    return {
        "file": str(path),
        "skill_file": str(resolve_skill_table_path(path, skill_path)),
        "structural_issue_count": len(structural_issues),
        "semantic_row_count": len(semantic_rows),
        "structural_issues": [asdict(issue) for issue in structural_issues],
        "semantic_rows": [row.to_dict() for row in semantic_rows],
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="校验 SkillUI_技能表现表.xlsx")
    parser.add_argument("path", help="SkillUI_技能表现表.xlsx 路径")
    parser.add_argument(
        "--skill",
        dest="skill_path",
        default=None,
        help="Skill.xlsx 路径（默认：与表现表同目录）",
    )
    parser.add_argument("--json", action="store_true", help="输出完整 JSON")
    parser.add_argument(
        "--semantic-json",
        action="store_true",
        help="仅输出待 LLM 语义的行",
    )
    args = parser.parse_args()

    path = Path(args.path)
    if not path.is_file():
        print(f"文件不存在: {path}", file=sys.stderr)
        return 2

    skill_path = Path(args.skill_path) if args.skill_path else None

    if args.json:
        print(
            json.dumps(
                build_report(path, skill_path=skill_path),
                ensure_ascii=False,
                indent=2,
            )
        )
        return 0

    data = load_rows(path)
    if args.semantic_json:
        rows = [row.to_dict() for row in collect_semantic_rows(data)]
        print(json.dumps(rows, ensure_ascii=False, indent=2, default=str))
        return 0

    issues = validate_structural(path, skill_path=skill_path)
    print(f"检查文件: {path}")
    print(f"技能表: {resolve_skill_table_path(path, skill_path)}")
    print(f"结构化问题数量: {len(issues)}")
    for issue in issues:
        print(issue.format_line())
    semantic_count = len(collect_semantic_rows(data))
    print(
        f"待语义分析行数: {semantic_count}"
        f"（见 skill：SkillName↔Id、ShortSkillText、SkillText、SettlementDes、Allusion、DesignThought）"
    )
    return 1 if issues else 0


if __name__ == "__main__":
    raise SystemExit(main())
