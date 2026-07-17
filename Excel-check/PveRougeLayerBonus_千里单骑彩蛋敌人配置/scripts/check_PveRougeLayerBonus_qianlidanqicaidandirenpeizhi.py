# -*- coding: utf-8 -*-
"""校验 PveRougeLayerBonus_千里单骑彩蛋敌人配置.xlsx：通用结构化规则（表结构 A，主键 MapLayerId）。"""
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
DATA_START_ROW = 4
SHEET_NAME = '千里单骑彩蛋敌人配置|PveRougeLayerBonus'
PK_FIELD = 'MapLayerId'
DATE_RE = re.compile(r"^\d{4}-\d{2}-\d{2}( \d{2}:\d{2}:\d{2})?$")
INT_ARRAY_RE = re.compile(r"^\d+(,\d+)*$")
ID_STR_RE = re.compile(r"^[A-Za-z][A-Za-z0-9_]*$")
BOOL_TRUE = {"1", "true", "yes", "y"}
BOOL_FALSE = {"0", "false", "no", "n"}
ID_TYPE = 'int'
DISPLAY_FIELD = ''


@dataclass
class Issue:
    row_id: object
    display: str
    field: str
    message: str
    source: str = "structural"

    def format_line(self) -> str:
        label = DISPLAY_FIELD or "-"
        disp = self.display or "-"
        rid = self.row_id if self.row_id != "" else "-"
        return f"{PK_FIELD}={rid} | {label}={disp} | {self.field} | {self.message}"


def is_empty(value: object) -> bool:
    if pd.isna(value):
        return True
    text = str(value).strip()
    return text in ("", "nan", "NaN")


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


def load_rows(path: Path) -> pd.DataFrame:
    raw = pd.read_excel(path, sheet_name=SHEET_NAME, header=None, dtype=object)
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
            continue
        keep_cols.append(i)
        col_names.append(str(header_val).strip())

    if PK_FIELD not in col_names:
        raise ValueError(f"主键列 {PK_FIELD} 不在字段名行: {col_names[:20]}")
    pk_idx = keep_cols[col_names.index(PK_FIELD)]

    type_by_name = {}
    for col_i, name in zip(keep_cols, col_names):
        t = types[col_i] if col_i < len(types) else None
        type_by_name[name] = "" if is_empty(t) else str(t).strip()

    records: list[dict] = []
    blank_run = 0
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


def display_of(row: dict) -> str:
    if DISPLAY_FIELD and DISPLAY_FIELD in row and not is_empty(row.get(DISPLAY_FIELD)):
        return str(row.get(DISPLAY_FIELD)).strip()
    return ""


def validate_structural(path: Path) -> list[Issue]:
    data = load_rows(path)
    issues: list[Issue] = []
    seen: list[object] = []

    for _, row in data.iterrows():
        raw_id = row.get(PK_FIELD)
        disp = display_of(row)
        types = row.get("__types__") or {}
        if isinstance(types, float):
            types = {}

        if ID_TYPE == "int":
            parsed = parse_int(raw_id)
            if parsed is None:
                issues.append(
                    Issue(raw_id, disp, PK_FIELD, f"{PK_FIELD} 须为 int，实际: {raw_id}")
                )
                continue
            row_id: object = parsed
            seen.append(parsed)
        else:
            text = str(raw_id).strip()
            if ID_TYPE == "string" and not ID_STR_RE.match(text):
                issues.append(
                    Issue(text, disp, PK_FIELD, f"{PK_FIELD} 须为标识符，实际: {raw_id}")
                )
            row_id = text
            seen.append(text)

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
            if field.startswith("Is") or field.startswith("Can") or t == "bool":
                if parse_bool(value) is None:
                    issues.append(
                        Issue(row_id, disp, field, f"{field} 须为 bool 语义，实际: {value}")
                    )
            if t.endswith("[]") and "int" in t.lower():
                text = str(value).strip().replace(" ", "")
                if not INT_ARRAY_RE.match(text) and parse_int(value) is None:
                    issues.append(
                        Issue(
                            row_id,
                            disp,
                            field,
                            f"{field} 应为 int 或逗号分隔 int 列表，实际: {value}",
                        )
                    )
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

        for a, b in (
            ("OnShelfTime", "OffShelfTime"),
            ("StartTime", "EndTime"),
            ("BeginTime", "EndTime"),
        ):
            if a in data.columns and b in data.columns:
                ea, eb = is_empty(row.get(a)), is_empty(row.get(b))
                if ea != eb:
                    issues.append(Issue(row_id, disp, a, f"{a}/{b} 须同空或同有"))

    counts = Counter(seen)
    for dup, n in counts.items():
        if n > 1:
            issues.append(Issue(dup, "", PK_FIELD, f"{PK_FIELD} 重复出现 {n} 次"))
    return issues


def main() -> int:
    parser = argparse.ArgumentParser(description="校验 PveRougeLayerBonus_千里单骑彩蛋敌人配置.xlsx")
    parser.add_argument("path", help="PveRougeLayerBonus_千里单骑彩蛋敌人配置.xlsx 路径")
    parser.add_argument("--json", action="store_true", help="输出 JSON")
    args = parser.parse_args()
    path = Path(args.path)
    if not path.is_file():
        print(f"文件不存在: {path}", file=sys.stderr)
        return 2
    issues = validate_structural(path)
    if args.json:
        print(
            json.dumps(
                {
                    "file": str(path),
                    "pk_field": PK_FIELD,
                    "structural_issue_count": len(issues),
                    "structural_issues": [asdict(i) for i in issues],
                    "semantic_rows": [],
                },
                ensure_ascii=False,
                indent=2,
            )
        )
        return 0
    print(f"检查文件: {path}")
    print(f"表结构: A | 主键列: {PK_FIELD}")
    print(f"结构化问题数量: {len(issues)}")
    for issue in issues:
        print(issue.format_line())
    print("待语义分析行数: 0（本表首轮仅结构化）")
    return 1 if issues else 0


if __name__ == "__main__":
    raise SystemExit(main())
