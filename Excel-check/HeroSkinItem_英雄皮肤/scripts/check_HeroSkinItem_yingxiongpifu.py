# -*- coding: utf-8 -*-
"""校验 HeroSkinItem_英雄皮肤.xlsx（表结构 A，主键 SkinItemId）。"""
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
SHEET_NAME = "英雄皮肤|HeroSkinItem"
PK_FIELD = "SkinItemId"
L1_FIELDS = [
    "Name",
    "GetWay",
    "Lines",
    "DebutLines",
    "KillLines",
    "DeadLines",
    "LinesDubbed",
    "OriginalArtDesigner",
    "BodyOffset",
]
DISPLAY_FIELD = "Name"
ID_TYPE = "int"
STRUCT_KIND = "A"
HERO_NAME = "Hero.xlsx"
HERO_SHEET = "武将表|Hero"
HERO_DATA_START = 4
# 上线武将须具备的四套基础皮肤
REQUIRED_OPEN_HERO_SKIN_TYPES = (
    "SkinNormalSkin",
    "SkinLineSkin",
    "SkinNormalDynamicsSkin",
    "SkinHKComicsSkin",
)
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


@dataclass(frozen=True)
class HeroMeta:
    eid: str
    name: str
    is_open: bool
    hero_type: int | None


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


def resolve_hero_path(skin_path: Path, hero_path: Path | None) -> Path:
    if hero_path is not None:
        return hero_path
    return skin_path.parent / HERO_NAME


def load_hero_meta(hero_path: Path) -> dict[int, HeroMeta]:
    """Hero.Id → E#EHeroId / Name / IsOpen。"""
    raw = pd.read_excel(hero_path, sheet_name=HERO_SHEET, header=None, dtype=object)
    types = raw.iloc[TYPE_ROW].tolist()
    headers = raw.iloc[HEADER_ROW].tolist()
    width = max(len(types), len(headers), len(raw.columns))
    id_idx = e_idx = name_idx = open_idx = hero_type_idx = None
    for i in range(width):
        hv = headers[i] if i < len(headers) else None
        tv = types[i] if i < len(types) else None
        if not is_empty(hv) and str(hv).strip() == "Id":
            id_idx = i
        if not is_empty(hv) and str(hv).strip() == "Name":
            name_idx = i
        if not is_empty(hv) and str(hv).strip() == "IsOpen":
            open_idx = i
        if not is_empty(hv) and str(hv).strip() == "HeroType":
            hero_type_idx = i
        if not is_empty(tv) and str(tv).strip() == "E#EHeroId":
            e_idx = i
    if id_idx is None:
        raise ValueError(f"{hero_path} 缺少主键列 Id")
    if e_idx is None:
        raise ValueError(f"{hero_path} 缺少类型行 E#EHeroId 列")
    if name_idx is None:
        raise ValueError(f"{hero_path} 缺少 Name 列")
    if open_idx is None:
        raise ValueError(f"{hero_path} 缺少 IsOpen 列")
    out: dict[int, HeroMeta] = {}
    blank_run = 0
    for r in range(HERO_DATA_START, len(raw)):
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
        ev = raw.iloc[r, e_idx]
        nv = raw.iloc[r, name_idx]
        ov = raw.iloc[r, open_idx]
        is_open = parse_bool(ov) is True
        hero_type = None
        if hero_type_idx is not None:
            hero_type = parse_int(raw.iloc[r, hero_type_idx])
        out[parsed] = HeroMeta(
            eid="" if is_empty(ev) else str(ev).strip(),
            name="" if is_empty(nv) else str(nv).strip(),
            is_open=is_open,
            hero_type=hero_type,
        )
    return out


def skin_pinyin_matches_hero_eid(skin_pinyin: str, hero_eid: str) -> bool:
    """SkinPinYin 与 E#EHeroId 对齐（大小写不敏感）。"""
    pl = skin_pinyin.strip().lower()
    el = hero_eid.strip().lower()
    if not pl or not el:
        return False

    def ok(base: str) -> bool:
        if not base:
            return False
        if pl == base or pl.startswith(base + "_"):
            return True
        # 无下划线粘连后缀（如 ZhuRong → zhurongfuren）
        return pl.startswith(base) and len(pl) > len(base)

    if ok(el):
        return True
    if "_" in el:
        base = el.split("_", 1)[0]
        if ok(base):
            return True
    return False


def validate_structural(path: Path, hero_path: Path | None = None):
    data = load_rows(path)
    issues, seen = [], []

    resolved_hero = resolve_hero_path(path, hero_path)
    hero_meta: dict[int, HeroMeta] | None = None
    if not resolved_hero.is_file():
        issues.append(
            Issue(
                "",
                "",
                "HeroId",
                f"缺少 {HERO_NAME}，无法校验 HeroId/SkinPinYin/上线四套皮肤: {resolved_hero}",
            )
        )
    else:
        try:
            hero_meta = load_hero_meta(resolved_hero)
        except Exception as exc:  # noqa: BLE001
            issues.append(
                Issue(
                    "",
                    "",
                    "HeroId",
                    f"读取 {HERO_NAME} 失败 ({resolved_hero}): {exc}",
                )
            )

    hero_skin_types: dict[int, set[str]] = defaultdict(set)

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

        # 独有：HeroId → Hero.Id；SkinPinYin ↔ E#EHeroId
        if "HeroId" in data.columns:
            if is_empty(row.get("HeroId")):
                issues.append(Issue(row_id, disp, "HeroId", "HeroId 不能为空"))
            else:
                hero_id = parse_int(row.get("HeroId"))
                if hero_id is None:
                    issues.append(
                        Issue(row_id, disp, "HeroId", f"HeroId 须为 int，实际: {row.get('HeroId')}")
                    )
                elif hero_meta is not None:
                    if hero_id not in hero_meta:
                        issues.append(
                            Issue(
                                row_id,
                                disp,
                                "HeroId",
                                f"HeroId={hero_id} 在 Hero.Id 中不存在",
                            )
                        )
                    else:
                        if "SkinType" in data.columns and not is_empty(row.get("SkinType")):
                            hero_skin_types[hero_id].add(str(row.get("SkinType")).strip())
                        skin_py = (
                            ""
                            if "SkinPinYin" not in data.columns or is_empty(row.get("SkinPinYin"))
                            else str(row.get("SkinPinYin")).strip()
                        )
                        if not skin_py:
                            issues.append(Issue(row_id, disp, "SkinPinYin", "SkinPinYin 不能为空"))
                        else:
                            eid = hero_meta[hero_id].eid
                            if not eid:
                                issues.append(
                                    Issue(
                                        row_id,
                                        disp,
                                        "SkinPinYin",
                                        f"HeroId={hero_id} 在 Hero 表无 E#EHeroId，无法校验 SkinPinYin",
                                    )
                                )
                            elif not skin_pinyin_matches_hero_eid(skin_py, eid):
                                issues.append(
                                    Issue(
                                        row_id,
                                        disp,
                                        "SkinPinYin",
                                        f"SkinPinYin={skin_py} 与 Hero.E#EHeroId={eid}（HeroId={hero_id}）对不上",
                                    )
                                )

        for a, b in (("OnShelfTime", "OffShelfTime"), ("StartTime", "EndTime"), ("BeginTime", "EndTime")):
            if a in data.columns and b in data.columns:
                if is_empty(row.get(a)) != is_empty(row.get(b)):
                    issues.append(Issue(row_id, disp, a, f"{a}/{b} 须同空或同有"))

    # 独有：正式上线武将（HeroType=1 且 IsOpen=1）须具备四套 SkinType
    if hero_meta is not None:
        for hero_id, meta in sorted(hero_meta.items()):
            if not meta.is_open or meta.hero_type != 1:
                continue
            have = hero_skin_types.get(hero_id, set())
            missing = [t for t in REQUIRED_OPEN_HERO_SKIN_TYPES if t not in have]
            if not missing:
                continue
            name = meta.name or str(hero_id)
            issues.append(
                Issue(
                    "",
                    name,
                    "SkinType",
                    f"上线正式武将 HeroId={hero_id}（Name={name}）缺少皮肤类型: {','.join(missing)}",
                )
            )

    for dup, n in Counter(seen).items():
        if n > 1:
            issues.append(Issue(dup, "", PK_FIELD, f"{PK_FIELD} 重复出现 {n} 次"))
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
    parser = argparse.ArgumentParser(description="校验 HeroSkinItem_英雄皮肤.xlsx")
    parser.add_argument("path")
    parser.add_argument(
        "--hero",
        dest="hero_path",
        default=None,
        help=f"{HERO_NAME} 路径（默认：与皮肤表同目录）",
    )
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--semantic-json", action="store_true")
    args = parser.parse_args()
    path = Path(args.path)
    if not path.is_file():
        print(f"文件不存在: {path}", file=sys.stderr)
        return 2
    hero_path = Path(args.hero_path) if args.hero_path else None
    issues, data = validate_structural(path, hero_path=hero_path)
    sem = semantic_rows(data)
    resolved_hero = resolve_hero_path(path, hero_path)
    if args.semantic_json:
        print(json.dumps({"file": str(path), "pk_field": PK_FIELD, "l1_fields": L1_FIELDS, "semantic_rows": sem}, ensure_ascii=False, indent=2))
        return 0
    if args.json:
        print(
            json.dumps(
                {
                    "file": str(path),
                    "hero_file": str(resolved_hero),
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
    print(f"武将表: {resolved_hero}")
    print(f"表结构: {STRUCT_KIND} | 主键列: {PK_FIELD} | L1字段: {', '.join(L1_FIELDS) if L1_FIELDS else '无'}")
    print(f"结构化问题数量: {len(issues)}")
    for issue in issues:
        print(issue.format_line())
    print(f"待语义分析行数: {len(sem)}（L1 文案质量；由 Agent 审）")
    return 1 if issues else 0


if __name__ == "__main__":
    raise SystemExit(main())
