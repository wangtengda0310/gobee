# -*- coding: utf-8 -*-
"""校验 RougeEndlessZhuGeLiangNode_千里单骑无尽诸葛亮节点表.xlsx（表结构 A，主键 ZhuGeLiangId）。"""
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
SHEET_NAME = '无尽诸葛|RougeEndlessZhuGeLiangNode'
PK_FIELD = 'ZhuGeLiangId'
L1_FIELDS = ['GuaXiangTitle', 'GuaXiangDesc', 'BuffDesc']
DISPLAY_FIELD = 'GuaXiangTitle'
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

    def format_line(self) -> str:
        label = DISPLAY_FIELD or "-"
        return f"{PK_FIELD}={self.row_id if self.row_id != '' else '-'} | {label}={self.display or '-'} | {self.field} | {self.message}"


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


def validate_structural(path: Path):
    data = load_rows(path)
    issues, seen = [], []
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
            if (field.startswith("Is") or field.startswith("Can")) and t != "bool":
                if parse_bool(value) is None:
                    issues.append(Issue(row_id, disp, field, f"{field} 须为 bool 语义，实际: {value}"))
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
    parser = argparse.ArgumentParser(description="校验 RougeEndlessZhuGeLiangNode_千里单骑无尽诸葛亮节点表.xlsx")
    parser.add_argument("path")
    parser.add_argument("--json", action="store_true")
    parser.add_argument("--semantic-json", action="store_true")
    args = parser.parse_args()
    path = Path(args.path)
    if not path.is_file():
        print(f"文件不存在: {path}", file=sys.stderr)
        return 2
    issues, data = validate_structural(path)
    sem = semantic_rows(data)
    if args.semantic_json:
        print(json.dumps({"file": str(path), "pk_field": PK_FIELD, "l1_fields": L1_FIELDS, "semantic_rows": sem}, ensure_ascii=False, indent=2))
        return 0
    if args.json:
        print(json.dumps({"file": str(path), "pk_field": PK_FIELD, "l1_fields": L1_FIELDS, "structural_issue_count": len(issues), "structural_issues": [asdict(i) for i in issues], "semantic_rows": sem}, ensure_ascii=False, indent=2))
        return 0
    print(f"检查文件: {path}")
    print(f"表结构: {STRUCT_KIND} | 主键列: {PK_FIELD} | L1字段: {', '.join(L1_FIELDS) if L1_FIELDS else '无'}")
    print(f"结构化问题数量: {len(issues)}")
    for issue in issues:
        print(issue.format_line())
    print(f"待语义分析行数: {len(sem)}（L1 文案质量；由 Agent 审）")
    return 1 if issues else 0


if __name__ == "__main__":
    raise SystemExit(main())
