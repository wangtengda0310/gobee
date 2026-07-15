# -*- coding: utf-8 -*-
"""校验 Hero.xlsx：结构化 + 导出语义行（Name↔拼音、Name→Country/MeltName/BelongExpansionPack/Gender）。"""
from __future__ import annotations

import argparse
import json
import re
import sys
from collections import Counter
from dataclasses import asdict, dataclass
from datetime import datetime
from pathlib import Path

import pandas as pd

TYPE_ROW = 1
HEADER_ROW = 2
DATA_START_ROW = 5
DATE_FMT = "%Y-%m-%d %H:%M:%S"
DATE_RE = re.compile(r"^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$")
INT_ARRAY_RE = re.compile(r"^\d+(,\d+)*$")
E_HERO_ID_RE = re.compile(r"^[A-Za-z][A-Za-z0-9_]*$")

HERO_TYPES = {1, 2, 3, 4, 5, 6}
COUNTRIES = {
    "CaoWei",
    "Chu",
    "DongHan",
    "Fei",
    "Han",
    "Huang",
    "NianShou",
    "Qi",
    "Qin",
    "Shu",
    "SunWu",
    "Wei",
    "XiChu",
    "XiHan",
    "XiJin",
    "XiZhou",
    "Yan",
    "ZhangChu",
    "Zhao",
}
EXPANSION_PACKS = {
    "HeroExpansionPack_ChuHanZhiZheng",
    "HeroExpansionPack_HeZongLianHeng",
    "HeroExpansionPack_JiangXinTianGong",
    "HeroExpansionPack_JinLuoXingTi",
    "HeroExpansionPack_QinSaoLiuHe",
    "HeroExpansionPack_SanGuoYunQi",
    "HeroExpansionPack_WuDiShengShi",
}
BOOL_FIELDS = (
    "IsOpen",
    "IsAlwaysZhuGong",
    "CanMelt",
    "IsNewHero",
    "IsGacha",
)
INT_ARRAY_FIELDS = (
    "Skill",
    "ExcludeIdentity",
    "NotUseModeType",
    "AssociationHeroId",
)
TYPE1_REQUIRED = (
    "IsOpen",
    "Gender",
    "Point",
    "HpLimit",
    "HandLimit",
    "EquipLimit",
    "Country",
    "IsAlwaysZhuGong",
    "CanMelt",
    "BelongExpansionPack",
)
TYPE1_OPEN_REQUIRED = (
    "Skill",
    "MeltName",
    "NotUseModeType",
)
TYPE3_EMPTY = (
    "IsOpen",
    "Gender",
    "Point",
    "HpLimit",
    "HandLimit",
    "EquipLimit",
    "Country",
    "Skill",
    "CanMelt",
    "MeltName",
    "BelongExpansionPack",
    "OpenDate",
)
SPECIAL_REQUIRED = (
    "IsOpen",
    "Gender",
    "Point",
    "HpLimit",
    "HandLimit",
    "EquipLimit",
    "Country",
)
SPECIAL_EMPTY = ("CanMelt", "MeltName", "BelongExpansionPack")


@dataclass
class Issue:
    row_id: object
    name: str
    field: str
    message: str
    source: str = "structural"

    def format_line(self) -> str:
        name = self.name or "-"
        rid = self.row_id if self.row_id != "" else "-"
        return f"Id={rid} | Name={name} | {self.field} | {self.message}"


def is_empty(value: object) -> bool:
    if pd.isna(value):
        return True
    text = str(value).strip()
    return text in ("", "nan", "NaN", "None")


def is_blank_id(raw_id: object) -> bool:
    if pd.isna(raw_id):
        return True
    text = str(raw_id).strip()
    return text in ("", "nan", "NaN", "None")


def is_comment_or_blank_id(raw_id: object) -> bool:
    if is_blank_id(raw_id):
        return True
    return str(raw_id).strip().startswith("#")


def truncate_after_three_blank_ids(data: pd.DataFrame) -> pd.DataFrame:
    if data.empty or "Id" not in data.columns:
        return data
    blank_run = 0
    cut_at: int | None = None
    for idx, raw_id in enumerate(data["Id"].tolist()):
        if is_blank_id(raw_id):
            blank_run += 1
            if blank_run >= 3:
                cut_at = idx - 2
                break
        else:
            blank_run = 0
    if cut_at is None:
        return data
    return data.iloc[: max(cut_at, 0)].copy()


def parse_int(value: object) -> int | None:
    if is_empty(value):
        return None
    if isinstance(value, bool):
        return None
    if isinstance(value, int):
        return value
    if isinstance(value, float) and value.is_integer():
        return int(value)
    text = str(value).strip()
    if text.isdigit() or (text.startswith("-") and text[1:].isdigit()):
        return int(text)
    return None


def parse_bool(value: object) -> bool | None:
    if is_empty(value):
        return None
    if isinstance(value, bool):
        return value
    if isinstance(value, (int, float)) and not isinstance(value, bool):
        if value == 1:
            return True
        if value == 0:
            return False
        return None
    text = str(value).strip().lower()
    if text in ("true", "1"):
        return True
    if text in ("false", "0"):
        return False
    return None


def load_rows(path: Path) -> pd.DataFrame:
    raw = pd.read_excel(path, header=None, dtype=object)
    types = raw.iloc[TYPE_ROW].tolist()
    headers = raw.iloc[HEADER_ROW].tolist()
    keep_idx: list[int] = []
    keep_names: list[object] = []
    width = max(len(types), len(headers), len(raw.columns))
    for i in range(width):
        type_val = types[i] if i < len(types) else None
        header_val = headers[i] if i < len(headers) else None
        if is_empty(type_val) and is_empty(header_val):
            continue
        type_text = "" if is_empty(type_val) else str(type_val).strip()
        # 无字段名的列默认丢弃；例外保留 E#EHeroId / E#EHeroType
        if is_empty(header_val):
            if type_text in ("E#EHeroId", "E#EHeroType"):
                keep_idx.append(i)
                keep_names.append(type_text)
            continue
        keep_idx.append(i)
        keep_names.append(header_val)
    data = raw.iloc[DATA_START_ROW:, keep_idx].copy()
    data.columns = keep_names
    data = truncate_after_three_blank_ids(data)
    mask = ~data["Id"].map(is_comment_or_blank_id)
    return data.loc[mask].reset_index(drop=True)


def add(
    issues: list[Issue],
    row_id: object,
    name: str,
    field: str,
    message: str,
) -> None:
    issues.append(Issue(row_id, name, field, message))


def require_present(
    issues: list[Issue],
    row_id: object,
    name: str,
    field: str,
    value: object,
) -> bool:
    if is_empty(value):
        add(issues, row_id, name, field, f"{field} 不能为空")
        return False
    return True


def require_empty(
    issues: list[Issue],
    row_id: object,
    name: str,
    field: str,
    value: object,
) -> None:
    if not is_empty(value):
        add(issues, row_id, name, field, f"{field} 应为空，实际: {value}")


def resolve_skill_table_path(hero_path: Path, skill_path: Path | None) -> Path:
    if skill_path is not None:
        return skill_path
    return hero_path.parent / "Skill.xlsx"


def load_skill_ids(skill_path: Path) -> set[int]:
    """从 Skill.xlsx 收集技能 Id（连续空 3 行截断；跳过 # 注释行）。"""
    raw = pd.read_excel(skill_path, header=None, dtype=object)
    types = raw.iloc[TYPE_ROW].tolist()
    headers = raw.iloc[HEADER_ROW].tolist()
    id_col: int | None = None
    width = max(len(types), len(headers), len(raw.columns))
    for i in range(width):
        header_val = headers[i] if i < len(headers) else None
        if not is_empty(header_val) and str(header_val).strip() == "Id":
            id_col = i
            break
    if id_col is None:
        id_col = 0
    ids: set[int] = set()
    blank_run = 0
    for raw_id in raw.iloc[DATA_START_ROW:, id_col].tolist():
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
            ids.add(parsed)
    return ids


def validate_structural(
    path: Path,
    skill_path: Path | None = None,
) -> list[Issue]:
    data = load_rows(path)
    issues: list[Issue] = []
    ids: list[int] = []
    all_ids: set[int] = set()
    e_hero_ids: list[str] = []

    resolved_skill = resolve_skill_table_path(path, skill_path)
    skill_ids: set[int] | None = None
    if not resolved_skill.is_file():
        add(
            issues,
            "",
            "",
            "Skill",
            f"缺少 Skill.xlsx，无法校验技能 Id 外联: {resolved_skill}",
        )
    else:
        try:
            skill_ids = load_skill_ids(resolved_skill)
        except Exception as exc:  # noqa: BLE001
            add(
                issues,
                "",
                "",
                "Skill",
                f"读取 Skill.xlsx 失败 ({resolved_skill}): {exc}",
            )

    for _, row in data.iterrows():
        raw_id = row.get("Id")
        name = "" if is_empty(row.get("Name")) else str(row.get("Name")).strip()
        row_id = parse_int(raw_id)
        if row_id is None:
            add(issues, raw_id, name, "Id", f"Id 必须是 int，实际: {raw_id}")
            continue
        ids.append(row_id)
        all_ids.add(row_id)

        if not name:
            add(issues, row_id, name, "Name", "Name 必须是 string 且非空")

        # Name（中文）可重复；E#EHeroId（拼音）唯一
        e_hero_id_raw = row.get("E#EHeroId") if "E#EHeroId" in data.columns else None
        if is_empty(e_hero_id_raw):
            add(issues, row_id, name, "E#EHeroId", "E#EHeroId 不能为空")
        else:
            e_hero_id = str(e_hero_id_raw).strip()
            if not E_HERO_ID_RE.match(e_hero_id):
                add(
                    issues,
                    row_id,
                    name,
                    "E#EHeroId",
                    f"E#EHeroId 须为拼音/标识符 ^[A-Za-z][A-Za-z0-9_]*$，实际: {e_hero_id}",
                )
            e_hero_ids.append(e_hero_id)

        hero_type_raw = row.get("HeroType")
        e_hero_type = row.get("E#EHeroType") if "E#EHeroType" in data.columns else None
        ht_empty = is_empty(hero_type_raw)
        e_empty = is_empty(e_hero_type)
        hero_type = parse_int(hero_type_raw)

        if ht_empty and e_empty:
            # 与 E#EHeroType 同空：允许，跳过类型分叉
            hero_type = None
        elif ht_empty and not e_empty:
            add(
                issues,
                row_id,
                name,
                "HeroType",
                f"E#EHeroType 有值（{e_hero_type}）时 HeroType 不能为空",
            )
            hero_type = None
        elif not ht_empty and e_empty:
            add(
                issues,
                row_id,
                name,
                "E#EHeroType",
                f"HeroType 有值（{hero_type_raw}）时 E#EHeroType 不能为空",
            )
            # 仍用 HeroType 做后续分叉
            if hero_type is None:
                add(
                    issues,
                    row_id,
                    name,
                    "HeroType",
                    f"HeroType 必须是 int，实际: {hero_type_raw}",
                )
            elif hero_type not in HERO_TYPES:
                add(
                    issues,
                    row_id,
                    name,
                    "HeroType",
                    f"HeroType 不在枚举 {sorted(HERO_TYPES)}，实际: {hero_type}",
                )
        else:
            if hero_type is None:
                add(
                    issues,
                    row_id,
                    name,
                    "HeroType",
                    f"HeroType 必须是 int，实际: {hero_type_raw}",
                )
            elif hero_type not in HERO_TYPES:
                add(
                    issues,
                    row_id,
                    name,
                    "HeroType",
                    f"HeroType 不在枚举 {sorted(HERO_TYPES)}，实际: {hero_type}",
                )

        for field in BOOL_FIELDS:
            if field not in data.columns:
                continue
            value = row.get(field)
            if is_empty(value):
                continue
            if parse_bool(value) is None:
                add(
                    issues,
                    row_id,
                    name,
                    field,
                    f"{field} 必须是 bool（true/false/0/1），实际: {value}",
                )

        open_date = row.get("OpenDate")
        if not is_empty(open_date):
            text = str(open_date).strip()
            if isinstance(open_date, datetime):
                text = open_date.strftime(DATE_FMT)
            elif hasattr(open_date, "to_pydatetime"):
                try:
                    text = open_date.to_pydatetime().strftime(DATE_FMT)
                except Exception:
                    text = str(open_date).strip()
            if not DATE_RE.match(text):
                # 容忍带毫秒
                ok = False
                for fmt in (DATE_FMT, "%Y-%m-%d %H:%M:%S.%f"):
                    try:
                        datetime.strptime(text, fmt)
                        ok = True
                        break
                    except ValueError:
                        continue
                if not ok:
                    add(
                        issues,
                        row_id,
                        name,
                        "OpenDate",
                        f"OpenDate 须为 YYYY-MM-DD HH:MM:SS，实际: {open_date}",
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
                    name,
                    field,
                    f"{field} 应为逗号分隔的 int 列表，实际: {value}",
                )
            elif field == "Skill" and skill_ids is not None:
                for part in text.split(","):
                    sid = int(part)
                    if sid not in skill_ids:
                        add(
                            issues,
                            row_id,
                            name,
                            "Skill",
                            f"Skill Id={sid} 不在 Skill.xlsx 的 Id 中",
                        )

        melt_name = row.get("MeltName") if "MeltName" in data.columns else None
        if not is_empty(melt_name):
            parts = [p.strip() for p in str(melt_name).split(",")]
            if any(not p for p in parts) or len(parts) < 1:
                add(
                    issues,
                    row_id,
                    name,
                    "MeltName",
                    f"MeltName 应为逗号分隔非空片段，实际: {melt_name}",
                )

        buff = row.get("Buff") if "Buff" in data.columns else None
        if not is_empty(buff) and not str(buff).strip():
            add(issues, row_id, name, "Buff", "Buff 有值时不能为空白")

        # 独有：HeroType 分叉
        if hero_type == 1:
            for field in TYPE1_REQUIRED:
                if field in data.columns:
                    require_present(issues, row_id, name, field, row.get(field))
            gender = parse_int(row.get("Gender"))
            if gender is not None and gender not in (1, 2):
                add(
                    issues,
                    row_id,
                    name,
                    "Gender",
                    f"HeroType=1 时 Gender 须为 1 或 2，实际: {row.get('Gender')}",
                )
            point = parse_int(row.get("Point"))
            hp = parse_int(row.get("HpLimit"))
            if point is not None and hp is not None and point != hp:
                add(
                    issues,
                    row_id,
                    name,
                    "Point",
                    f"HeroType=1 时 Point 须等于 HpLimit，实际 Point={point} HpLimit={hp}",
                )
            equip = parse_int(row.get("EquipLimit"))
            if equip is not None and equip != 3:
                add(
                    issues,
                    row_id,
                    name,
                    "EquipLimit",
                    f"HeroType=1 时 EquipLimit 须为 3，实际: {equip}",
                )
            country = (
                "" if is_empty(row.get("Country")) else str(row.get("Country")).strip()
            )
            if country and country not in COUNTRIES:
                add(
                    issues,
                    row_id,
                    name,
                    "Country",
                    f"Country 不在枚举内: {country}",
                )
            pack = (
                ""
                if is_empty(row.get("BelongExpansionPack"))
                else str(row.get("BelongExpansionPack")).strip()
            )
            if pack and pack not in EXPANSION_PACKS:
                add(
                    issues,
                    row_id,
                    name,
                    "BelongExpansionPack",
                    f"BelongExpansionPack 不在枚举内: {pack}",
                )
            is_open = parse_bool(row.get("IsOpen"))
            if is_open is True:
                for field in TYPE1_OPEN_REQUIRED:
                    if field not in data.columns:
                        continue
                    require_present(
                        issues,
                        row_id,
                        name,
                        field,
                        row.get(field),
                    )
                skill = row.get("Skill")
                if not is_empty(skill):
                    text = str(skill).strip().replace(" ", "")
                    if not INT_ARRAY_RE.match(text):
                        add(
                            issues,
                            row_id,
                            name,
                            "Skill",
                            f"HeroType=1 且 IsOpen=1 时 Skill 须为逗号分隔 int，实际: {skill}",
                        )
                nmt = row.get("NotUseModeType")
                if not is_empty(nmt):
                    text = str(nmt).strip().replace(" ", "")
                    if not INT_ARRAY_RE.match(text):
                        add(
                            issues,
                            row_id,
                            name,
                            "NotUseModeType",
                            f"HeroType=1 且 IsOpen=1 时 NotUseModeType 须为逗号分隔 int，"
                            f"实际: {nmt}",
                        )

        elif hero_type == 3:
            for field in TYPE3_EMPTY:
                if field in data.columns:
                    require_empty(issues, row_id, name, field, row.get(field))

        elif hero_type in (2, 4, 5, 6):
            for field in SPECIAL_REQUIRED:
                if field in data.columns:
                    require_present(issues, row_id, name, field, row.get(field))
            for field in SPECIAL_EMPTY:
                if field in data.columns:
                    require_empty(issues, row_id, name, field, row.get(field))

    # AssociationHeroId → 本表 Id
    for _, row in data.iterrows():
        row_id = parse_int(row.get("Id"))
        if row_id is None:
            continue
        name = "" if is_empty(row.get("Name")) else str(row.get("Name")).strip()
        assoc = row.get("AssociationHeroId") if "AssociationHeroId" in data.columns else None
        if is_empty(assoc):
            continue
        text = str(assoc).strip().replace(" ", "")
        if not INT_ARRAY_RE.match(text):
            continue
        for part in text.split(","):
            ref = int(part)
            if ref not in all_ids:
                add(
                    issues,
                    row_id,
                    name,
                    "AssociationHeroId",
                    f"AssociationHeroId 引用 {ref} 不在本表 Id 中",
                )

    for duplicate_id, count in Counter(ids).items():
        if count > 1:
            add(issues, duplicate_id, "", "Id", f"Id 重复，出现 {count} 次")

    for eid, count in Counter(e_hero_ids).items():
        if count > 1:
            add(
                issues,
                "",
                "",
                "E#EHeroId",
                f"E#EHeroId={eid} 重复，出现 {count} 次（拼音不可重复）",
            )

    return issues


@dataclass
class SemanticRow:
    id: int
    name: str
    e_hero_id: str
    country: str
    melt_name: str
    belong_expansion_pack: str
    gender: object
    hero_type: object

    def to_dict(self) -> dict:
        return asdict(self)


def collect_semantic_rows(data: pd.DataFrame) -> list[SemanticRow]:
    """导出待 LLM 判断的行：Name↔拼音、Name→Country/MeltName/BelongExpansionPack/Gender。"""
    rows: list[SemanticRow] = []
    for _, row in data.iterrows():
        row_id = parse_int(row.get("Id"))
        if row_id is None:
            continue
        name = "" if is_empty(row.get("Name")) else str(row.get("Name")).strip()
        e_id = (
            ""
            if is_empty(row.get("E#EHeroId"))
            else str(row.get("E#EHeroId")).strip()
        )
        if not name or not e_id:
            continue
        country = (
            "" if is_empty(row.get("Country")) else str(row.get("Country")).strip()
        )
        melt_name = (
            "" if is_empty(row.get("MeltName")) else str(row.get("MeltName")).strip()
        )
        pack = (
            ""
            if is_empty(row.get("BelongExpansionPack"))
            else str(row.get("BelongExpansionPack")).strip()
        )
        gender = parse_int(row.get("Gender"))
        hero_type = parse_int(row.get("HeroType"))
        rows.append(
            SemanticRow(
                id=row_id,
                name=name,
                e_hero_id=e_id,
                country=country,
                melt_name=melt_name,
                belong_expansion_pack=pack,
                gender=gender if gender is not None else "",
                hero_type=hero_type if hero_type is not None else "",
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
    parser = argparse.ArgumentParser(description="校验 Hero.xlsx")
    parser.add_argument("path", help="Hero.xlsx 路径")
    parser.add_argument(
        "--skill",
        dest="skill_path",
        default=None,
        help="Skill.xlsx 路径（默认：与武将表同目录）",
    )
    parser.add_argument("--json", action="store_true", help="输出完整 JSON")
    parser.add_argument(
        "--semantic-json",
        action="store_true",
        help="仅输出待 LLM 语义的行（含 Country/MeltName/BelongExpansionPack/Gender）",
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
        f"（见 skill：Name ↔ E#EHeroId、Name → Country / MeltName / BelongExpansionPack / Gender）"
    )
    return 1 if issues else 0


if __name__ == "__main__":
    sys.exit(main())
