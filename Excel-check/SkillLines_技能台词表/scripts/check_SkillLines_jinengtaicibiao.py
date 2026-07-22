# -*- coding: utf-8 -*-
"""校验 SkillLines_技能台词表.xlsx（表结构 A，主键 Id）。"""
from __future__ import annotations

import argparse
import json
import re
import sys
from collections import Counter, defaultdict
from dataclasses import asdict, dataclass
from pathlib import Path

_LIB_DIR = Path(__file__).resolve().parents[2] / "_lib"
if str(_LIB_DIR) not in sys.path:
    sys.path.insert(0, str(_LIB_DIR))
from type_check import check_value_against_type  # noqa: E402

import pandas as pd

TYPE_ROW = 1
HEADER_ROW = 2
DATA_START_ROW = 4
SHEET_NAME = "技能台词配置表|SkillLines"
PK_FIELD = "Id"
L1_FIELDS = []  # Skill*Line 为 int[] 音效 ID，L1 不适用
DISPLAY_FIELD = "SkillFirstLine"
ID_TYPE = "int"
STRUCT_KIND = "A"
LINE_FIELDS = (
    "SkillFirstLine",
    "SkillSecondLine",
    "SkillThirdLine",
    "SkillForthLine",
)
HERO_SKIN_ITEM_NAME = "HeroSkinItem_英雄皮肤.xlsx"
HERO_SKIN_ITEM_SHEET = "英雄皮肤|HeroSkinItem"
HERO_SKIN_ITEM_PK = "SkinItemId"
HERO_SKIN_ITEM_DATA_START = 4
HERO_LINES_NAME = "HeroLines_武将台词表.xlsx"
HERO_LINES_SHEET = "武将台词|HeroLines"
HERO_LINES_DATA_START = 4
SKILL_NAME = "Skill.xlsx"
SKILL_SHEET = "技能表|Skill"
SKILL_DATA_START = 5
DATE_RE = re.compile(r"^\d{4}-\d{2}-\d{2}( \d{2}:\d{2}:\d{2})?$")
INT_ARRAY_RE = re.compile(r"^\d+(,\d+)*$")
ID_STR_RE = re.compile(r"^[A-Za-z][A-Za-z0-9_]*$")
BOOL_TRUE = {"1", "true", "yes", "y"}
BOOL_FALSE = {"0", "false", "no", "n"}


@dataclass
class Issue:
    row_id: object
    display: str
    field: str
    message: str
    source: str = "structural"

    def format_line(self) -> str:
        label = DISPLAY_FIELD or "-"
        return (
            f"{PK_FIELD}={self.row_id if self.row_id != '' else '-'} | "
            f"{label}={self.display or '-'} | {self.field} | {self.message}"
        )


def is_empty(value: object) -> bool:
    if pd.isna(value):
        return True
    return str(value).strip() in ("", "nan", "NaN")


def parse_int(value: object):
    if is_empty(value):
        return None
    try:
        return int(float(str(value).strip()))
    except (TypeError, ValueError):
        return None


def parse_int_list(value: object) -> list[int]:
    if is_empty(value):
        return []
    text = str(value).strip().replace(" ", "")
    out: list[int] = []
    for part in text.split(","):
        if not part:
            continue
        parsed = parse_int(part)
        if parsed is not None:
            out.append(parsed)
    return out


def parse_bool(value: object):
    if is_empty(value):
        return None
    t = str(value).strip().lower()
    if t in BOOL_TRUE:
        return True
    if t in BOOL_FALSE:
        return False
    return None


def load_rows(path: Path) -> pd.DataFrame:
    raw = pd.read_excel(path, sheet_name=SHEET_NAME, header=None, dtype=object)
    if TYPE_ROW is None:
        headers = raw.iloc[HEADER_ROW].tolist()
        types = [None] * len(headers)
    else:
        types = raw.iloc[TYPE_ROW].tolist()
        headers = raw.iloc[HEADER_ROW].tolist()
    width = max(len(types), len(headers), len(raw.columns))
    keep_cols, col_names = [], []
    for i in range(width):
        hv = headers[i] if i < len(headers) else None
        tv = types[i] if i < len(types) else None
        if TYPE_ROW is not None and is_empty(tv) and is_empty(hv):
            continue
        if is_empty(hv):
            continue
        keep_cols.append(i)
        col_names.append(str(hv).strip())
    if PK_FIELD not in col_names:
        raise ValueError(f"主键列 {PK_FIELD} 不在字段名行")
    pk_idx = keep_cols[col_names.index(PK_FIELD)]
    type_by_name = {}
    for col_i, name in zip(keep_cols, col_names):
        t = types[col_i] if col_i < len(types) else None
        type_by_name[name] = "" if is_empty(t) else str(t).strip()
    records, blank_run = [], 0
    for r in range(DATA_START_ROW, len(raw)):
        raw_id = raw.iloc[r, pk_idx]
        if is_empty(raw_id):
            blank_run += 1
            if blank_run >= 3:
                break
            continue
        blank_run = 0
        if str(raw_id).strip().startswith("#"):
            continue
        row = {name: raw.iloc[r, col_i] for col_i, name in zip(keep_cols, col_names)}
        row["__types__"] = type_by_name
        records.append(row)
    return pd.DataFrame.from_records(records)


def display_of(row) -> str:
    if DISPLAY_FIELD and DISPLAY_FIELD in row and not is_empty(row.get(DISPLAY_FIELD)):
        return str(row.get(DISPLAY_FIELD)).strip()
    return ""


def resolve_path(base: Path, override: Path | None, filename: str) -> Path:
    if override is not None:
        return override
    return base.parent / filename


def load_skin_item_ids(hero_skin_item_path: Path) -> set[int]:
    raw = pd.read_excel(
        hero_skin_item_path, sheet_name=HERO_SKIN_ITEM_SHEET, header=None, dtype=object
    )
    headers = raw.iloc[HEADER_ROW].tolist()
    pk_idx = None
    for i, hv in enumerate(headers):
        if not is_empty(hv) and str(hv).strip() == HERO_SKIN_ITEM_PK:
            pk_idx = i
            break
    if pk_idx is None:
        raise ValueError(f"主键列 {HERO_SKIN_ITEM_PK} 不在字段名行")
    ids: set[int] = set()
    blank_run = 0
    for r in range(HERO_SKIN_ITEM_DATA_START, len(raw)):
        raw_id = raw.iloc[r, pk_idx]
        if is_empty(raw_id):
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


def load_hero_lines_tab_by_id(hero_lines_path: Path) -> dict[int, str]:
    raw = pd.read_excel(
        hero_lines_path, sheet_name=HERO_LINES_SHEET, header=None, dtype=object
    )
    headers = raw.iloc[HEADER_ROW].tolist()
    id_idx = tab_idx = None
    for i, hv in enumerate(headers):
        if is_empty(hv):
            continue
        name = str(hv).strip()
        if name == "Id":
            id_idx = i
        elif name == "TabName":
            tab_idx = i
    if id_idx is None:
        raise ValueError("主键列 Id 不在 HeroLines 字段名行")
    if tab_idx is None:
        raise ValueError("TabName 不在 HeroLines 字段名行")
    out: dict[int, str] = {}
    blank_run = 0
    for r in range(HERO_LINES_DATA_START, len(raw)):
        raw_id = raw.iloc[r, id_idx]
        if is_empty(raw_id):
            blank_run += 1
            if blank_run >= 3:
                break
            continue
        blank_run = 0
        if str(raw_id).strip().startswith("#"):
            continue
        parsed = parse_int(raw_id)
        if parsed is None:
            continue
        tab = raw.iloc[r, tab_idx]
        out[parsed] = "" if is_empty(tab) else str(tab).strip()
    return out


def load_skill_name_to_eids(skill_path: Path) -> dict[str, set[str]]:
    raw = pd.read_excel(skill_path, sheet_name=SKILL_SHEET, header=None, dtype=object)
    headers = raw.iloc[HEADER_ROW].tolist()
    types = raw.iloc[TYPE_ROW].tolist()
    name_idx = None
    e_idx = None
    id_idx = 0
    width = max(len(headers), len(types), len(raw.columns))
    for i in range(width):
        hv = headers[i] if i < len(headers) else None
        tv = types[i] if i < len(types) else None
        if not is_empty(hv) and str(hv).strip() == "Id":
            id_idx = i
        if not is_empty(hv) and str(hv).strip() == "SkillName":
            name_idx = i
        if not is_empty(tv) and str(tv).strip() == "E#ESkillId":
            e_idx = i
    if name_idx is None:
        raise ValueError("SkillName 不在 Skill.xlsx 字段名行")
    if e_idx is None:
        raise ValueError("Skill.xlsx 缺少类型行 E#ESkillId 列（技能拼音 ID）")
    out: dict[str, set[str]] = defaultdict(set)
    blank_run = 0
    for r in range(SKILL_DATA_START, len(raw)):
        raw_id = raw.iloc[r, id_idx]
        if is_empty(raw_id):
            blank_run += 1
            if blank_run >= 3:
                break
            continue
        blank_run = 0
        if str(raw_id).strip().startswith("#"):
            continue
        name = raw.iloc[r, name_idx]
        eid = raw.iloc[r, e_idx]
        if is_empty(name) or is_empty(eid):
            continue
        out[str(name).strip()].add(str(eid).strip())
    return out


def validate_structural(
    path: Path,
    hero_skin_item_path: Path | None = None,
    hero_lines_path: Path | None = None,
    skill_path: Path | None = None,
):
    data = load_rows(path)
    issues, seen = [], []

    resolved_skin = resolve_path(path, hero_skin_item_path, HERO_SKIN_ITEM_NAME)
    skin_item_ids: set[int] | None = None
    if not resolved_skin.is_file():
        issues.append(
            Issue(
                "",
                "",
                "SkinId",
                f"缺少 {HERO_SKIN_ITEM_NAME}，无法校验 SkinId 外联: {resolved_skin}",
            )
        )
    else:
        try:
            skin_item_ids = load_skin_item_ids(resolved_skin)
        except Exception as exc:  # noqa: BLE001
            issues.append(
                Issue(
                    "",
                    "",
                    "SkinId",
                    f"读取 {HERO_SKIN_ITEM_NAME} 失败 ({resolved_skin}): {exc}",
                )
            )

    resolved_hero_lines = resolve_path(path, hero_lines_path, HERO_LINES_NAME)
    hero_tab_by_id: dict[int, str] | None = None
    if not resolved_hero_lines.is_file():
        issues.append(
            Issue(
                "",
                "",
                "SkillFirstLine",
                f"缺少 {HERO_LINES_NAME}，无法校验 Skill*Line 外联: {resolved_hero_lines}",
            )
        )
    else:
        try:
            hero_tab_by_id = load_hero_lines_tab_by_id(resolved_hero_lines)
        except Exception as exc:  # noqa: BLE001
            issues.append(
                Issue(
                    "",
                    "",
                    "SkillFirstLine",
                    f"读取 {HERO_LINES_NAME} 失败 ({resolved_hero_lines}): {exc}",
                )
            )

    resolved_skill = resolve_path(path, skill_path, SKILL_NAME)
    skill_name_to_eids: dict[str, set[str]] | None = None
    if not resolved_skill.is_file():
        issues.append(
            Issue(
                "",
                "",
                "SkillId",
                f"缺少 {SKILL_NAME}，无法校验 TabName 拼音与 SkillId: {resolved_skill}",
            )
        )
    else:
        try:
            skill_name_to_eids = load_skill_name_to_eids(resolved_skill)
        except Exception as exc:  # noqa: BLE001
            issues.append(
                Issue(
                    "",
                    "",
                    "SkillId",
                    f"读取 {SKILL_NAME} 失败 ({resolved_skill}): {exc}",
                )
            )

    line_id_locs: dict[int, list[tuple[object, str, str]]] = defaultdict(list)

    for _, row in data.iterrows():
        raw_id = row.get(PK_FIELD)
        disp = display_of(row)
        types = row.get("__types__") or {}
        if isinstance(types, float):
            types = {}
        if ID_TYPE == "int":
            parsed = parse_int(raw_id)
            if parsed is None:
                issues.append(Issue(raw_id, disp, PK_FIELD, f"{PK_FIELD} 须为 int，实际: {raw_id}"))
                continue
            row_id, seen_val = parsed, parsed
        else:
            text = str(raw_id).strip()
            row_id, seen_val = text, text
        seen.append(seen_val)
        skill_id = ""
        if "SkillId" in data.columns:
            if is_empty(row.get("SkillId")):
                issues.append(Issue(row_id, disp, "SkillId", "SkillId 不能为空"))
            else:
                skill_id = str(row.get("SkillId")).strip()
        for field in ("Name", "Title", "SkillName"):
            if field in data.columns and is_empty(row.get(field)):
                issues.append(Issue(row_id, disp, field, f"{field} 不能为空"))
        for field in data.columns:
            if field in (PK_FIELD, "__types__"):
                continue
            value = row.get(field)
            if is_empty(value):
                continue
            t = str(types.get(field, "") if isinstance(types, dict) else "")
            type_err = check_value_against_type(t, value, is_empty=is_empty)
            if type_err:
                issues.append(Issue(row_id, disp, field, type_err))
            if "Time" in field or field.endswith("Date"):
                text = str(value).strip()
                if not DATE_RE.match(text) and re.search(r"\d{4}", text):
                    issues.append(
                        Issue(
                            row_id,
                            disp,
                            field,
                            f"{field} 时间格式应为 YYYY-MM-DD[ HH:MM:SS]，实际: {value}",
                        )
                    )
        if (
            skin_item_ids is not None
            and "SkinId" in data.columns
            and not is_empty(row.get("SkinId"))
        ):
            skin_id = parse_int(row.get("SkinId"))
            if skin_id is not None and skin_id not in skin_item_ids:
                issues.append(
                    Issue(
                        row_id,
                        disp,
                        "SkinId",
                        f"SkinId={skin_id} 在 HeroSkinItem.SkinItemId 中不存在",
                    )
                )

        for field in LINE_FIELDS:
            if field not in data.columns or is_empty(row.get(field)):
                continue
            for line_id in parse_int_list(row.get(field)):
                line_id_locs[line_id].append((row_id, disp, field))
                if hero_tab_by_id is not None and line_id not in hero_tab_by_id:
                    issues.append(
                        Issue(
                            row_id,
                            disp,
                            field,
                            f"台词Id={line_id} 在 HeroLines.Id 中不存在",
                        )
                    )
                    continue
                if hero_tab_by_id is None or not skill_id or skill_name_to_eids is None:
                    continue
                tab_name = hero_tab_by_id.get(line_id, "")
                eids = skill_name_to_eids.get(tab_name)
                if not eids:
                    issues.append(
                        Issue(
                            row_id,
                            disp,
                            field,
                            f"TabName={tab_name or '(空)'} 在 Skill.xlsx 找不到，无法校验与 SkillId 一致",
                        )
                    )
                    continue
                if skill_id not in eids:
                    shown = ",".join(sorted(eids))
                    issues.append(
                        Issue(
                            row_id,
                            disp,
                            field,
                            f"台词Id={line_id} TabName={tab_name} 对应拼音={shown} "
                            f"与本行 SkillId={skill_id} 不一致",
                        )
                    )

        for a, b in (("OnShelfTime", "OffShelfTime"), ("StartTime", "EndTime"), ("BeginTime", "EndTime")):
            if a in data.columns and b in data.columns:
                if is_empty(row.get(a)) != is_empty(row.get(b)):
                    issues.append(Issue(row_id, disp, a, f"{a}/{b} 须同空或同有"))

    for dup, n in Counter(seen).items():
        if n > 1:
            issues.append(Issue(dup, "", PK_FIELD, f"{PK_FIELD} 重复出现 {n} 次"))

    for line_id, locs in line_id_locs.items():
        if len(locs) <= 1:
            continue
        n = len(locs)
        for row_id, disp, field in locs:
            issues.append(
                Issue(
                    row_id,
                    disp,
                    field,
                    f"台词Id={line_id} 在 Skill*Line 中重复出现 {n} 次",
                )
            )

    return issues, data


def semantic_rows(data: pd.DataFrame):
    rows = []
    for _, row in data.iterrows():
        item = {"id": None if is_empty(row.get(PK_FIELD)) else str(row.get(PK_FIELD)).strip()}
        if DISPLAY_FIELD:
            item["display"] = display_of(row)
        fields = {}
        for f in L1_FIELDS:
            if f in data.columns and not is_empty(row.get(f)):
                fields[f] = str(row.get(f)).strip()
        if not fields:
            continue
        item["l1_fields"] = fields
        rows.append(item)
    return rows


def main() -> int:
    parser = argparse.ArgumentParser(description="校验 SkillLines_技能台词表.xlsx")
    parser.add_argument("path")
    parser.add_argument(
        "--hero-skin-item",
        dest="hero_skin_item_path",
        default=None,
        help=f"{HERO_SKIN_ITEM_NAME} 路径（默认：与技能台词表同目录）",
    )
    parser.add_argument(
        "--hero-lines",
        dest="hero_lines_path",
        default=None,
        help=f"{HERO_LINES_NAME} 路径（默认：与技能台词表同目录）",
    )
    parser.add_argument(
        "--skill",
        dest="skill_path",
        default=None,
        help=f"{SKILL_NAME} 路径（默认：与技能台词表同目录）",
    )
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--semantic-json", action="store_true")
    args = parser.parse_args()
    path = Path(args.path)
    if not path.is_file():
        print(f"文件不存在: {path}", file=sys.stderr)
        return 2
    hero_skin_item_path = Path(args.hero_skin_item_path) if args.hero_skin_item_path else None
    hero_lines_path = Path(args.hero_lines_path) if args.hero_lines_path else None
    skill_path = Path(args.skill_path) if args.skill_path else None
    issues, data = validate_structural(
        path,
        hero_skin_item_path=hero_skin_item_path,
        hero_lines_path=hero_lines_path,
        skill_path=skill_path,
    )
    sem = semantic_rows(data)
    resolved_skin = resolve_path(path, hero_skin_item_path, HERO_SKIN_ITEM_NAME)
    resolved_hero_lines = resolve_path(path, hero_lines_path, HERO_LINES_NAME)
    resolved_skill = resolve_path(path, skill_path, SKILL_NAME)
    if args.semantic_json:
        print(
            json.dumps(
                {
                    "file": str(path),
                    "pk_field": PK_FIELD,
                    "l1_fields": L1_FIELDS,
                    "semantic_rows": sem,
                },
                ensure_ascii=False,
                indent=2,
            )
        )
        return 0
    if args.json:
        print(
            json.dumps(
                {
                    "file": str(path),
                    "hero_skin_item_file": str(resolved_skin),
                    "hero_lines_file": str(resolved_hero_lines),
                    "skill_file": str(resolved_skill),
                    "pk_field": PK_FIELD,
                    "l1_fields": L1_FIELDS,
                    "structural_issue_count": len(issues),
                    "structural_issues": [asdict(i) for i in issues],
                    "semantic_rows": sem,
                },
                ensure_ascii=False,
                indent=2,
            )
        )
        return 0
    print(f"检查文件: {path}")
    print(f"英雄皮肤表: {resolved_skin}")
    print(f"武将台词表: {resolved_hero_lines}")
    print(f"技能表: {resolved_skill}")
    print(
        f"表结构: {STRUCT_KIND} | 主键列: {PK_FIELD} | "
        f"L1字段: {', '.join(L1_FIELDS) if L1_FIELDS else '无'}"
    )
    print(f"结构化问题数量: {len(issues)}")
    for issue in issues:
        print(issue.format_line())
    print(f"待语义分析行数: {len(sem)}（L1 文案质量；由 Agent 审）")
    return 1 if issues else 0


if __name__ == "__main__":
    raise SystemExit(main())
