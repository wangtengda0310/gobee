# -*- coding: utf-8 -*-
"""Excel-check 公共 S12：按类型行校验字段值形态。"""
from __future__ import annotations

import re
from typing import Callable

INT_RE = re.compile(r"^-?\d+(\.0+)?$")
FLOAT_RE = re.compile(r"^-?\d+(\.\d+)?([eE][+-]?\d+)?$")
INT_ARRAY_RE = re.compile(r"^-?\d+(\.0+)?(,-?\d+(\.0+)?)*$")
ID_STR_RE = re.compile(r"^[A-Za-z][A-Za-z0-9_]*$")
BOOL_TRUE = {"1", "true", "yes", "y"}
BOOL_FALSE = {"0", "false", "no", "n"}

INT_TYPE_NAMES = {
    "int",
    "long",
    "int32",
    "int64",
    "uint",
    "uint32",
    "uint64",
    "short",
    "byte",
}
FLOAT_TYPE_NAMES = {"float", "double", "single", "number"}


def _norm_type(declared: str) -> str:
    t = (declared or "").strip()
    # 表内常见写法 E#ESkillId → ESkillId
    if t.startswith("E#"):
        t = t[2:].strip()
    return t


def type_map_from_raw(
    raw,
    *,
    type_row: int = 1,
    header_row: int = 2,
    is_empty: Callable[[object], bool],
) -> dict[str, str]:
    """从 A 表 raw DataFrame 提取 字段名 → 类型行 声明。"""
    types = raw.iloc[type_row].tolist()
    headers = raw.iloc[header_row].tolist()
    width = max(len(types), len(headers), len(raw.columns))
    out: dict[str, str] = {}
    for i in range(width):
        hv = headers[i] if i < len(headers) else None
        tv = types[i] if i < len(types) else None
        if is_empty(hv):
            continue
        name = str(hv).strip()
        out[name] = "" if is_empty(tv) else str(tv).strip()
    return out


def collect_s12_field_errors(
    row: dict | object,
    type_by_name: dict[str, str],
    *,
    pk_field: str,
    is_empty: Callable[[object], bool],
    skip_fields: set[str] | None = None,
) -> list[tuple[str, str]]:
    """返回 [(field, message), ...]；跳过主键与 skip_fields。"""
    skip = {pk_field, "__types__", "_Section"} | (skip_fields or set())
    errors: list[tuple[str, str]] = []
    getter = row.get if hasattr(row, "get") else lambda k, d=None: row[k] if k in row else d
    for field, declared in type_by_name.items():
        if field in skip or not field:
            continue
        try:
            value = getter(field)
        except Exception:
            continue
        if is_empty(value):
            continue
        err = check_value_against_type(declared, value, is_empty=is_empty)
        if err:
            errors.append((field, err))
    return errors


def _is_empty(value: object, is_empty: Callable[[object], bool]) -> bool:
    return is_empty(value)


def _as_text(value: object) -> str:
    return str(value).strip()


def _parse_bool(value: object) -> bool | None:
    t = _as_text(value).lower()
    if t in BOOL_TRUE:
        return True
    if t in BOOL_FALSE:
        return False
    return None


def _is_int_text(text: str) -> bool:
    if INT_RE.match(text):
        return True
    try:
        f = float(text)
        return f == int(f)
    except (TypeError, ValueError):
        return False


def _is_float_text(text: str) -> bool:
    if FLOAT_RE.match(text.replace(" ", "")):
        return True
    try:
        float(text)
        return True
    except (TypeError, ValueError):
        return False


def _is_enum_token(text: str) -> bool:
    if not text:
        return False
    if ID_STR_RE.match(text):
        return True
    return _is_int_text(text)


def _element_base(declared: str) -> str:
    """Strip one trailing [] from type name."""
    t = declared.strip()
    if t.endswith("[]"):
        return t[:-2].strip()
    return t


def check_value_against_type(
    declared_type: str,
    value: object,
    *,
    is_empty: Callable[[object], bool],
) -> str | None:
    """
    若不符合类型行声明，返回问题说明；符合或跳过则返回 None。
    空值由调用方跳过（本函数若收到空值也返回 None）。
    """
    if _is_empty(value, is_empty):
        return None
    declared = _norm_type(declared_type)
    if not declared or declared in {"#", "client", "server", "client/server"}:
        return None

    text = _as_text(value)
    compact = text.replace(" ", "")

    # arrays
    if declared.endswith("[]"):
        base = _element_base(declared)
        base_l = base.lower()
        if not compact:
            return f"类型为 {declared}，不能为空白"

        # int[] / long[] …
        if base_l in INT_TYPE_NAMES or base_l == "int":
            if not INT_ARRAY_RE.match(compact):
                return f"类型为 {declared}，应为 int 或逗号分隔 int 列表，实际: {value}"
            return None

        # string[]
        if base_l == "string":
            parts = [p.strip() for p in text.split(",")]
            if not parts or any(not p for p in parts):
                return f"类型为 {declared}，应为逗号分隔非空 string，实际: {value}"
            return None

        # E*[]
        if base.startswith("E") and len(base) > 1:
            parts = [p.strip() for p in text.split(",")]
            if not parts or any(not _is_enum_token(p) for p in parts):
                return (
                    f"类型为 {declared}，每项应为标识串或 int，实际: {value}"
                )
            return None

        # ItemCfg[] / Pair3[] / {…}[] / 其它复合数组：仅非空
        return None

    declared_l = declared.lower()

    if declared_l in INT_TYPE_NAMES:
        if not _is_int_text(compact):
            return f"类型为 {declared}，须为 int，实际: {value}"
        return None

    if declared_l in FLOAT_TYPE_NAMES:
        if not _is_float_text(compact):
            return f"类型为 {declared}，须为数字，实际: {value}"
        return None

    if declared_l == "bool":
        if _parse_bool(value) is None:
            return f"类型为 bool，实际: {value}"
        return None

    if declared_l == "string":
        return None

    # E* scalar enums
    if declared.startswith("E") and len(declared) > 1 and not declared.endswith("[]"):
        if not _is_enum_token(text):
            return f"类型为 {declared}，须为标识串或 int，实际: {value}"
        return None

    # ItemCfg / 复合标量：有值即过
    if declared in {"ItemCfg"} or "Cfg" in declared or declared.startswith("{"):
        return None

    # 未识别类型：不报
    return None
