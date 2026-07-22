# -*- coding: utf-8 -*-
"""校验 HeroLines_武将台词表.xlsx（表结构 A，主键 Id）。"""
from __future__ import annotations

import argparse
import json
import re
import sys
from collections import Counter
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
SHEET_NAME = '武将台词|HeroLines'
PK_FIELD = 'Id'
L1_FIELDS = ['TabName', 'Text']
DISPLAY_FIELD = 'TabName'
ID_TYPE = 'int'
STRUCT_KIND = 'A'
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
    warning: bool = False

    def format_line(self) -> str:
        label = DISPLAY_FIELD or "-"
        prefix = "[警告] " if self.warning else ""
        return (
            f"{prefix}{PK_FIELD}={self.row_id if self.row_id != '' else '-'} | "
            f"{label}={self.display or '-'} | {self.field} | {self.message}"
        )


@dataclass(frozen=True)
class HeroOwner:
    hero_id: int
    name: str
    is_open: bool
    eid: str


AUDIO_HERO_EID_RE = re.compile(r"^Vo_Hero_([A-Za-z]+)")
HERO_SHEET = "武将表|Hero"

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


FIXED_TAB_TYPE = {
    "登场": "LinesType_Dengchang",
    "击杀": "LinesType_Kill",
    "阵亡": "LinesType_Dead",
}
EMPTY_TYPE_TABS = {"自选", "重伤", "退场"}
SKILL_LINE_TYPE = "LinesType_Skill"
SKILL_SHEET = "技能表|Skill"
SKILL_LINES_SHEET = "技能台词配置表|SkillLines"
SKILL_LINES_LINE_FIELDS = (
    "SkillFirstLine",
    "SkillSecondLine",
    "SkillThirdLine",
    "SkillForthLine",
)

# 年兽 Boss 行：AudioId 前缀；无台词文案，不适用 Skill 外联
NIANSHOU_AUDIO_PREFIX = "Vo_Boss_NianShou_"


def is_nianshou_row(audio_id: str) -> bool:
    return bool(audio_id) and audio_id.startswith(NIANSHOU_AUDIO_PREFIX)

def resolve_skill_table_path(hero_lines_path: Path, skill_path: Path | None) -> Path:
    if skill_path is not None:
        return skill_path
    return hero_lines_path.parent / "Skill.xlsx"


def resolve_skill_lines_path(hero_lines_path: Path, skill_lines_path: Path | None) -> Path:
    if skill_lines_path is not None:
        return skill_lines_path
    return hero_lines_path.parent / "SkillLines_技能台词表.xlsx"


def resolve_hero_table_path(hero_lines_path: Path, hero_path: Path | None) -> Path:
    if hero_path is not None:
        return hero_path
    return hero_lines_path.parent / "Hero.xlsx"


def parse_int_list_cell(value: object) -> set[int]:
    """解析 int / 逗号分隔 int 列表（兼容分号）。"""
    if is_empty(value):
        return set()
    text = str(value).strip().replace(";", ",")
    out: set[int] = set()
    for part in text.split(","):
        part = part.strip()
        if not part:
            continue
        n = parse_int(part)
        if n is not None:
            out.add(n)
    return out


def load_skill_name_meta(skill_path: Path) -> tuple[dict[str, list[str]], dict[str, list[int]]]:
    """SkillName → (E#ESkillId 列表, 数值 Skill Id 列表)。同名多技能保留全部。"""
    raw = pd.read_excel(skill_path, sheet_name=SKILL_SHEET, header=None, dtype=object)
    types = raw.iloc[TYPE_ROW].tolist()
    headers = raw.iloc[HEADER_ROW].tolist()
    width = max(len(types), len(headers), len(raw.columns))
    id_col = 0
    name_col: int | None = None
    eid_col: int | None = None
    for i in range(width):
        hv = headers[i] if i < len(headers) else None
        tv = types[i] if i < len(types) else None
        if not is_empty(hv) and str(hv).strip() == "Id":
            id_col = i
        if not is_empty(hv) and str(hv).strip() == "SkillName":
            name_col = i
        if not is_empty(tv) and str(tv).strip() == "E#ESkillId":
            eid_col = i
    if name_col is None:
        raise ValueError(f"{skill_path} 缺少 SkillName 列")
    if eid_col is None:
        raise ValueError(f"{skill_path} 缺少类型行 E#ESkillId 列（技能拼音 ID）")

    name_to_eids: dict[str, list[str]] = {}
    name_to_ids: dict[str, list[int]] = {}
    blank_run = 0
    for r in range(DATA_START_ROW, len(raw)):
        raw_id = raw.iloc[r, id_col]
        if is_empty(raw_id):
            blank_run += 1
            if blank_run >= 3:
                break
            continue
        blank_run = 0
        if str(raw_id).strip().startswith("#"):
            continue
        name_val = raw.iloc[r, name_col]
        if is_empty(name_val):
            continue
        name = str(name_val).strip()
        sid = parse_int(raw_id)
        if sid is not None:
            id_bucket = name_to_ids.setdefault(name, [])
            if sid not in id_bucket:
                id_bucket.append(sid)
        eid_val = raw.iloc[r, eid_col]
        if not is_empty(eid_val):
            eid = str(eid_val).strip()
            eid_bucket = name_to_eids.setdefault(name, [])
            if eid not in eid_bucket:
                eid_bucket.append(eid)
    return name_to_eids, name_to_ids


def load_skill_name_to_eids(skill_path: Path) -> dict[str, list[str]]:
    """兼容：仅返回 SkillName → E#ESkillId。"""
    name_to_eids, _ = load_skill_name_meta(skill_path)
    return name_to_eids


def load_skill_names(skill_path: Path) -> set[str]:
    """从 Skill.xlsx 收集非空 SkillName。"""
    name_to_eids, name_to_ids = load_skill_name_meta(skill_path)
    return set(name_to_eids.keys()) | set(name_to_ids.keys())


def load_hero_skill_owners(hero_path: Path) -> tuple[dict[int, list[HeroOwner]], dict[str, list[HeroOwner]]]:
    """Hero.Skill 反查：数值技能 Id → 武将；E#EHeroId → 武将。"""
    raw = pd.read_excel(hero_path, sheet_name=HERO_SHEET, header=None, dtype=object)
    types = raw.iloc[TYPE_ROW].tolist()
    headers = raw.iloc[HEADER_ROW].tolist()
    width = max(len(types), len(headers), len(raw.columns))
    id_col = 0
    name_col: int | None = None
    open_col: int | None = None
    skill_col: int | None = None
    eid_col: int | None = None
    for i in range(width):
        hv = headers[i] if i < len(headers) else None
        tv = types[i] if i < len(types) else None
        if not is_empty(hv) and str(hv).strip() == "Id":
            id_col = i
        if not is_empty(hv) and str(hv).strip() == "Name":
            name_col = i
        if not is_empty(hv) and str(hv).strip() == "IsOpen":
            open_col = i
        if not is_empty(hv) and str(hv).strip() == "Skill":
            skill_col = i
        if not is_empty(tv) and str(tv).strip() == "E#EHeroId":
            eid_col = i
    if name_col is None or open_col is None or skill_col is None:
        raise ValueError(f"{hero_path} 缺少 Name/IsOpen/Skill 列")

    by_skill: dict[int, list[HeroOwner]] = {}
    by_eid: dict[str, list[HeroOwner]] = {}
    blank_run = 0
    for r in range(DATA_START_ROW, len(raw)):
        raw_id = raw.iloc[r, id_col]
        if is_empty(raw_id):
            blank_run += 1
            if blank_run >= 3:
                break
            continue
        blank_run = 0
        if str(raw_id).strip().startswith("#"):
            continue
        hero_id = parse_int(raw_id)
        if hero_id is None:
            continue
        name = "" if is_empty(raw.iloc[r, name_col]) else str(raw.iloc[r, name_col]).strip()
        is_open = parse_bool(raw.iloc[r, open_col])
        # IsOpen 空或无法解析按未开放处理
        open_flag = True if is_open is True else False
        eid = ""
        if eid_col is not None and not is_empty(raw.iloc[r, eid_col]):
            eid = str(raw.iloc[r, eid_col]).strip()
        owner = HeroOwner(hero_id=hero_id, name=name, is_open=open_flag, eid=eid)
        if eid:
            by_eid.setdefault(eid, []).append(owner)
        for sid in parse_int_list_cell(raw.iloc[r, skill_col]):
            by_skill.setdefault(sid, []).append(owner)
    return by_skill, by_eid


def resolve_row_heroes(
    tab_name: str,
    audio_id: str,
    skill_name_to_ids: dict[str, list[int]] | None,
    skill_id_to_heroes: dict[int, list[HeroOwner]] | None,
    eid_to_heroes: dict[str, list[HeroOwner]] | None,
) -> list[HeroOwner]:
    """技能 TabName → Hero.Skill 归属；否则用 AudioId 中武将拼音匹配 E#EHeroId。"""
    found: dict[int, HeroOwner] = {}
    if skill_name_to_ids is not None and skill_id_to_heroes is not None:
        if (
            tab_name
            and tab_name not in FIXED_TAB_TYPE
            and tab_name not in EMPTY_TYPE_TABS
        ):
            for sid in skill_name_to_ids.get(tab_name, []):
                for h in skill_id_to_heroes.get(sid, []):
                    found[h.hero_id] = h
    if not found and eid_to_heroes is not None and audio_id:
        m = AUDIO_HERO_EID_RE.match(audio_id)
        if m:
            for h in eid_to_heroes.get(m.group(1), []):
                found[h.hero_id] = h
    return list(found.values())


def closed_hero_note(heroes: list[HeroOwner]) -> str:
    closed = [h.name or str(h.hero_id) for h in heroes if not h.is_open]
    if not closed:
        return ""
    # 去重保序
    uniq: list[str] = []
    for n in closed:
        if n not in uniq:
            uniq.append(n)
    return "武将还未开放（" + "、".join(uniq) + "）"


def heroes_all_not_open(heroes: list[HeroOwner]) -> bool:
    """有归属武将且全部 IsOpen=0。"""
    return bool(heroes) and all(not h.is_open for h in heroes)


def annotate_issues_with_hero_open(
    issues: list[Issue],
    row_heroes: dict[object, list[HeroOwner]],
) -> None:
    """所有 Issue 若归属武将 IsOpen=0，在说明末尾标明「武将还未开放」。"""
    for issue in issues:
        if issue.row_id == "" or issue.row_id is None:
            continue
        note = closed_hero_note(row_heroes.get(issue.row_id, []))
        if not note:
            continue
        if "武将还未开放" in issue.message:
            continue
        issue.message = f"{issue.message}；{note}"


def demote_skilllines_missing_for_closed_heroes(
    issues: list[Issue],
    row_heroes: dict[object, list[HeroOwner]],
) -> None:
    """未开放武将：SkillLines 找不到 SkillId 降为警告（不阻断）。"""
    for issue in issues:
        if "SkillLines 中找不到 SkillId=" not in issue.message:
            continue
        if heroes_all_not_open(row_heroes.get(issue.row_id, [])):
            issue.warning = True


def load_skill_lines_id_index(skill_lines_path: Path) -> dict[str, set[int]]:
    """SkillId → 其各 Skill*Line 列中出现的 HeroLines Id 并集（含皮肤行）。"""
    raw = pd.read_excel(skill_lines_path, sheet_name=SKILL_LINES_SHEET, header=None, dtype=object)
    types = raw.iloc[TYPE_ROW].tolist()
    headers = raw.iloc[HEADER_ROW].tolist()
    width = max(len(types), len(headers), len(raw.columns))
    keep: dict[str, int] = {}
    id_col = 0
    for i in range(width):
        hv = headers[i] if i < len(headers) else None
        tv = types[i] if i < len(types) else None
        if is_empty(hv) and is_empty(tv):
            continue
        if is_empty(hv):
            continue
        name = str(hv).strip()
        keep[name] = i
        if name == "Id":
            id_col = i
    for need in ("SkillId", *SKILL_LINES_LINE_FIELDS):
        if need not in keep:
            raise ValueError(f"{skill_lines_path} 缺少列 {need}")

    index: dict[str, set[int]] = {}
    blank_run = 0
    for r in range(DATA_START_ROW, len(raw)):
        raw_id = raw.iloc[r, id_col]
        if is_empty(raw_id):
            blank_run += 1
            if blank_run >= 3:
                break
            continue
        blank_run = 0
        if str(raw_id).strip().startswith("#"):
            continue
        sid_val = raw.iloc[r, keep["SkillId"]]
        if is_empty(sid_val):
            continue
        sid = str(sid_val).strip()
        bucket = index.setdefault(sid, set())
        for field in SKILL_LINES_LINE_FIELDS:
            bucket |= parse_int_list_cell(raw.iloc[r, keep[field]])
    return index


def validate_structural(
    path: Path,
    skill_path: Path | None = None,
    skill_lines_path: Path | None = None,
    hero_path: Path | None = None,
):
    data = load_rows(path)
    issues, seen = [], []
    text_owners: dict[str, list[tuple[object, str]]] = {}
    audio_owners: dict[str, list[tuple[object, str]]] = {}
    row_heroes: dict[object, list[HeroOwner]] = {}

    resolved_skill = resolve_skill_table_path(path, skill_path)
    skill_names: set[str] | None = None
    skill_name_to_eids: dict[str, list[str]] | None = None
    skill_name_to_ids: dict[str, list[int]] | None = None
    if not resolved_skill.is_file():
        issues.append(
            Issue(
                "",
                "",
                "Skill",
                f"缺少 Skill.xlsx，无法校验技能类 TabName 外联: {resolved_skill}",
            )
        )
    else:
        try:
            skill_name_to_eids, skill_name_to_ids = load_skill_name_meta(resolved_skill)
            skill_names = set(skill_name_to_eids.keys()) | set(skill_name_to_ids.keys())
        except Exception as exc:  # noqa: BLE001
            issues.append(
                Issue(
                    "",
                    "",
                    "Skill",
                    f"读取 Skill.xlsx 失败 ({resolved_skill}): {exc}",
                )
            )

    resolved_skill_lines = resolve_skill_lines_path(path, skill_lines_path)
    skill_lines_index: dict[str, set[int]] | None = None
    if not resolved_skill_lines.is_file():
        issues.append(
            Issue(
                "",
                "",
                "SkillLines",
                f"缺少 SkillLines_技能台词表.xlsx，无法校验 LinesType_Skill 外联: {resolved_skill_lines}",
            )
        )
    else:
        try:
            skill_lines_index = load_skill_lines_id_index(resolved_skill_lines)
        except Exception as exc:  # noqa: BLE001
            issues.append(
                Issue(
                    "",
                    "",
                    "SkillLines",
                    f"读取 SkillLines 失败 ({resolved_skill_lines}): {exc}",
                )
            )

    resolved_hero = resolve_hero_table_path(path, hero_path)
    skill_id_to_heroes: dict[int, list[HeroOwner]] | None = None
    eid_to_heroes: dict[str, list[HeroOwner]] | None = None
    if not resolved_hero.is_file():
        issues.append(
            Issue(
                "",
                "",
                "Hero",
                f"缺少 Hero.xlsx，无法标注武将开放态: {resolved_hero}",
            )
        )
    else:
        try:
            skill_id_to_heroes, eid_to_heroes = load_hero_skill_owners(resolved_hero)
        except Exception as exc:  # noqa: BLE001
            issues.append(
                Issue(
                    "",
                    "",
                    "Hero",
                    f"读取 Hero.xlsx 失败 ({resolved_hero}): {exc}",
                )
            )

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
        for field in ("Name", "Title", "SkillName"):
            if field in data.columns and is_empty(row.get(field)):
                issues.append(Issue(row_id, disp, field, f"{field} 不能为空"))

        # 独有：TabName / Type / SkillName 外联
        tab_name = "" if is_empty(row.get("TabName")) else str(row.get("TabName")).strip()
        type_val = "" if is_empty(row.get("Type")) else str(row.get("Type")).strip()
        audio_id = "" if is_empty(row.get("AudioId")) else str(row.get("AudioId")).strip()
        nianshou = is_nianshou_row(audio_id)
        row_heroes[row_id] = resolve_row_heroes(
            tab_name,
            audio_id,
            skill_name_to_ids,
            skill_id_to_heroes,
            eid_to_heroes,
        )
        if not tab_name:
            issues.append(Issue(row_id, disp or "-", "TabName", "TabName 不能为空"))
        elif tab_name in FIXED_TAB_TYPE:
            expected = FIXED_TAB_TYPE[tab_name]
            if type_val != expected:
                issues.append(
                    Issue(
                        row_id,
                        tab_name,
                        "Type",
                        f"TabName={tab_name} 时 Type 须为 {expected}，实际: {type_val or '(空)'}",
                    )
                )
        elif tab_name in EMPTY_TYPE_TABS:
            if type_val:
                issues.append(
                    Issue(
                        row_id,
                        tab_name,
                        "Type",
                        f"TabName={tab_name} 时 Type 须为空，实际: {type_val}",
                    )
                )
        elif not nianshou:
            # 年兽 Boss 技能 TabName（如 烈焰噬心1）不在 Skill.xlsx，跳过外联
            if skill_names is not None:
                if tab_name not in skill_names:
                    issues.append(
                        Issue(
                            row_id,
                            tab_name,
                            "TabName",
                            f"在 Skill.xlsx 的 SkillName 中找不到此技能: {tab_name}",
                        )
                    )
                elif type_val != SKILL_LINE_TYPE:
                    issues.append(
                        Issue(
                            row_id,
                            tab_name,
                            "Type",
                            f"TabName 为技能名时 Type 须为 {SKILL_LINE_TYPE}，实际: {type_val or '(空)'}",
                        )
                    )

        # 独有：LinesType_Skill → SkillLines Skill*Line 须引用本行 Id
        if (
            type_val == SKILL_LINE_TYPE
            and not nianshou
            and tab_name
            and skill_lines_index is not None
            and skill_name_to_eids is not None
        ):
            eids = skill_name_to_eids.get(tab_name) or []
            if eids:
                matched_eids = [e for e in eids if e in skill_lines_index]
                if not matched_eids:
                    issues.append(
                        Issue(
                            row_id,
                            tab_name,
                            "Id",
                            "SkillLines 中找不到 SkillId="
                            + "/".join(eids)
                            + f"（由 TabName={tab_name} → Skill.xlsx E#ESkillId）",
                        )
                    )
                else:
                    line_ids: set[int] = set()
                    for e in matched_eids:
                        line_ids |= skill_lines_index.get(e, set())
                    if row_id not in line_ids:
                        eid_shown = "/".join(matched_eids)
                        issues.append(
                            Issue(
                                row_id,
                                tab_name,
                                "Id",
                                f"Id={row_id} 须出现在 SkillLines SkillId={eid_shown} 行的 "
                                "SkillFirstLine/SkillSecondLine/SkillThirdLine/SkillForthLine 中",
                            )
                        )

        # 独有：Text 必填（年兽无台词，可空）
        line_text = "" if is_empty(row.get("Text")) else str(row.get("Text")).strip()
        if not line_text:
            if not nianshou:
                issues.append(Issue(row_id, tab_name or disp or "-", "Text", "Text 不能为空"))
        else:
            text_owners.setdefault(line_text, []).append((row_id, tab_name or disp or "-"))

        # 独有：AudioId 必填 + 全表唯一（不做 Vo_Hero 格式 / 页签段校验）
        display_key = tab_name or disp or "-"
        if not audio_id:
            issues.append(Issue(row_id, display_key, "AudioId", "AudioId 不能为空"))
        else:
            audio_owners.setdefault(audio_id, []).append((row_id, display_key))

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
                    issues.append(Issue(row_id, disp, field, f"{field} 时间格式应为 YYYY-MM-DD[ HH:MM:SS]，实际: {value}"))
        for a, b in (("OnShelfTime", "OffShelfTime"), ("StartTime", "EndTime"), ("BeginTime", "EndTime")):
            if a in data.columns and b in data.columns:
                if is_empty(row.get(a)) != is_empty(row.get(b)):
                    issues.append(Issue(row_id, disp, a, f"{a}/{b} 须同空或同有"))
    for dup, n in Counter(seen).items():
        if n > 1:
            issues.append(Issue(dup, "", PK_FIELD, f"{PK_FIELD} 重复出现 {n} 次"))
    for text, owners in text_owners.items():
        if len(owners) <= 1:
            continue
        id_list = ",".join(str(rid) for rid, _ in owners)
        first_id, first_tab = owners[0]
        preview = text if len(text) <= 40 else text[:40] + "…"
        issues.append(
            Issue(
                first_id,
                first_tab,
                "Text",
                f"Text 重复出现 {len(owners)} 次（Id: {id_list}），文案: {preview}",
            )
        )
    for audio, owners in audio_owners.items():
        if len(owners) <= 1:
            continue
        id_list = ",".join(str(rid) for rid, _ in owners)
        first_id, first_tab = owners[0]
        issues.append(
            Issue(
                first_id,
                first_tab,
                "AudioId",
                f"AudioId 重复出现 {len(owners)} 次（Id: {id_list}），值: {audio}",
            )
        )
    annotate_issues_with_hero_open(issues, row_heroes)
    demote_skilllines_missing_for_closed_heroes(issues, row_heroes)
    return issues, data, row_heroes


def semantic_rows(data: pd.DataFrame, row_heroes: dict[object, list[HeroOwner]] | None = None):
    rows = []
    for _, row in data.iterrows():
        rid = None if is_empty(row.get(PK_FIELD)) else str(row.get(PK_FIELD)).strip()
        item = {"id": rid}
        if DISPLAY_FIELD:
            item["display"] = display_of(row)
        fields = {}
        for f in L1_FIELDS:
            if f in data.columns and not is_empty(row.get(f)):
                fields[f] = str(row.get(f)).strip()
        if not fields:
            continue
        item["l1_fields"] = fields
        if row_heroes is not None and rid is not None:
            # int key may be used in row_heroes
            key: object = rid
            parsed = parse_int(rid)
            if parsed is not None and parsed in row_heroes:
                key = parsed
            note = closed_hero_note(row_heroes.get(key, []))
            if note:
                item["hero_not_open"] = note
        rows.append(item)
    return rows


def main() -> int:
    parser = argparse.ArgumentParser(description="校验 HeroLines_武将台词表.xlsx")
    parser.add_argument("path")
    parser.add_argument(
        "--skill",
        dest="skill_path",
        default=None,
        help="Skill.xlsx 路径（默认：与台词表同目录）",
    )
    parser.add_argument(
        "--skill-lines",
        dest="skill_lines_path",
        default=None,
        help="SkillLines_技能台词表.xlsx 路径（默认：与台词表同目录）",
    )
    parser.add_argument(
        "--hero",
        dest="hero_path",
        default=None,
        help="Hero.xlsx 路径（默认：与台词表同目录）",
    )
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--semantic-json", action="store_true")
    args = parser.parse_args()
    path = Path(args.path)
    if not path.is_file():
        print(f"文件不存在: {path}", file=sys.stderr)
        return 2
    skill_path = Path(args.skill_path) if args.skill_path else None
    skill_lines_path = Path(args.skill_lines_path) if args.skill_lines_path else None
    hero_path = Path(args.hero_path) if args.hero_path else None
    issues, data, row_heroes = validate_structural(
        path,
        skill_path=skill_path,
        skill_lines_path=skill_lines_path,
        hero_path=hero_path,
    )
    sem = semantic_rows(data, row_heroes=row_heroes)
    errors = [i for i in issues if not i.warning]
    warnings = [i for i in issues if i.warning]
    if args.semantic_json:
        print(json.dumps({"file": str(path), "pk_field": PK_FIELD, "l1_fields": L1_FIELDS, "semantic_rows": sem}, ensure_ascii=False, indent=2))
        return 0
    if args.json:
        print(
            json.dumps(
                {
                    "file": str(path),
                    "skill_file": str(resolve_skill_table_path(path, skill_path)),
                    "skill_lines_file": str(resolve_skill_lines_path(path, skill_lines_path)),
                    "hero_file": str(resolve_hero_table_path(path, hero_path)),
                    "pk_field": PK_FIELD,
                    "l1_fields": L1_FIELDS,
                    "structural_issue_count": len(errors),
                    "structural_warning_count": len(warnings),
                    "structural_issues": [asdict(i) for i in errors],
                    "structural_warnings": [asdict(i) for i in warnings],
                    "semantic_rows": sem,
                },
                ensure_ascii=False,
                indent=2,
            )
        )
        return 0
    print(f"检查文件: {path}")
    print(f"技能表: {resolve_skill_table_path(path, skill_path)}")
    print(f"技能台词表: {resolve_skill_lines_path(path, skill_lines_path)}")
    print(f"武将表: {resolve_hero_table_path(path, hero_path)}")
    print(
        f"表结构: {STRUCT_KIND} | 主键列: {PK_FIELD} | "
        f"L1字段: {', '.join(L1_FIELDS) if L1_FIELDS else '无'}"
    )
    print(f"结构化问题数量: {len(errors)} | 警告数量: {len(warnings)}")
    for issue in errors:
        print(issue.format_line())
    if warnings:
        print("--- 警告（未开放武将等，不阻断）---")
        for issue in warnings:
            print(issue.format_line())
    print(f"待语义分析行数: {len(sem)}（L1 文案质量；由 Agent 审）")
    return 1 if errors else 0


if __name__ == "__main__":
    raise SystemExit(main())
