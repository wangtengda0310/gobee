# -*- coding: utf-8 -*-
"""校验 Recharge_充值表.xlsx：仅结构化规则，无语义阶段。"""
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
PRODUCT_ID_RE = re.compile(r"^qudao_\d+\.\d+$")
INT_ARRAY_RE = re.compile(r"^\d+(,\d+)*$")
SEASON_PACK_RE = re.compile(r"^S(\d+)(豪华版礼包|典藏版礼包)$")
NAME_PRICE_RE = re.compile(r"^(\d+)元")
GOLD_LIANG_RE = re.compile(r"^(\d+)两黄金$")
SEASON_ON_RE = re.compile(r"^(\d{4})-(\d{2})-15 (\d{2}):(\d{2}):(\d{2})$")
SEASON_OFF_RE = re.compile(r"^(\d{4})-(\d{2})-14 23:59:59$")
SEASON_SKIN_BOX_NAME = "赛季皮肤礼盒"
SEASON_SKIN_BOX_PLATFORMS = ("andriod", "pc", "ios")
SEASON_SKIN_BOX_EMPTY_FIELDS = ("Channel", "Depend", "Mutex", "GiftGoodID")

RECHARGE_TYPES = {
    "MonthlyCard",
    "SeasonPass",
    "ShopGood",
    "LimitGift",
    "PickGoods",
    "Activity",
    "UnSale",
}
PLATFORMS = {"andriod", "ios", "pc"}
LIMIT_TYPES = {1, 2, 3, 4}


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
    """连续空 3 行后截断：Id 连续为空达 3 行时，截断其后内容（业务不用）。"""
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


def has_edge_whitespace(value: object) -> bool:
    if is_empty(value):
        return False
    text = str(value)
    return text != text.strip() or "\n" in text or "\r" in text


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


def parse_float(value: object) -> float | None:
    if is_empty(value):
        return None
    if isinstance(value, (int, float)) and not isinstance(value, bool):
        return float(value)
    try:
        return float(str(value).strip())
    except ValueError:
        return None


def shelf_time_str(value: object) -> str | None:
    """规范化上/下架时间为 YYYY-MM-DD HH:MM:SS。"""
    if is_empty(value):
        return None
    if isinstance(value, datetime):
        return value.strftime(DATE_FMT)
    if hasattr(value, "to_pydatetime"):
        try:
            return value.to_pydatetime().strftime(DATE_FMT)
        except Exception:
            pass
    text = str(value).strip()
    if DATE_RE.match(text):
        return text
    for fmt in (DATE_FMT, "%Y-%m-%d %H:%M:%S.%f", "%Y/%m/%d %H:%M:%S"):
        try:
            return datetime.strptime(text, fmt).strftime(DATE_FMT)
        except ValueError:
            continue
    return text


PLATFORM_MAP_COL = {
    "andriod": "IdAndroid",
    "pc": "IdPc",
    "ios": "IdIos",
}


def resolve_map_path(recharge_path: Path, map_path: Path | None) -> Path:
    if map_path is not None:
        return map_path
    return recharge_path.parent / "Recharge_映射表.xlsx"


def load_recharge_id_map(map_path: Path) -> dict[str, set[int]]:
    """按端收集映射表中的充值 Id：Platform → {Id, ...}。

    映射表中段可能有空行、末尾行仍有数据，故不做「连续空 3 行截断」。
    """
    raw = pd.read_excel(map_path, header=None, dtype=object)
    columns = select_checkable_columns(raw)
    data = raw.iloc[DATA_START_ROW:, : len(columns)].copy()
    data.columns = columns
    data = data.loc[:, [c for c in data.columns if not is_empty(c)]]

    result: dict[str, set[int]] = {p: set() for p in PLATFORM_MAP_COL}
    for _, row in data.iterrows():
        for plat, col in PLATFORM_MAP_COL.items():
            if col not in data.columns:
                continue
            parsed = parse_int(row.get(col))
            if parsed is not None:
                result[plat].add(parsed)
    return result


def select_checkable_columns(raw: pd.DataFrame) -> list[object]:
    """第2行(类型)与第3行(字段名)皆空的列不检查。"""
    types = raw.iloc[TYPE_ROW].tolist()
    headers = raw.iloc[HEADER_ROW].tolist()
    width = max(len(types), len(headers))
    selected: list[object] = []
    for i in range(width):
        type_val = types[i] if i < len(types) else None
        header_val = headers[i] if i < len(headers) else None
        if is_empty(type_val) and is_empty(header_val):
            continue
        selected.append(header_val)
    return selected


def load_rows(path: Path) -> pd.DataFrame:
    raw = pd.read_excel(path, header=None, dtype=object)
    columns = select_checkable_columns(raw)
    data = raw.iloc[DATA_START_ROW:, : len(columns)].copy()
    data.columns = columns
    data = data.loc[:, [c for c in data.columns if not is_empty(c)]]
    data = truncate_after_three_blank_ids(data)

    # 注释行 `#xxx` 标记分区，供独有规则识别（如 #支付中心黄金）
    sections: list[str] = []
    current = ""
    for raw_id in data["Id"].tolist():
        if not is_blank_id(raw_id) and str(raw_id).strip().startswith("#"):
            current = str(raw_id).strip()
        sections.append(current)
    data = data.copy()
    data["_Section"] = sections

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


def require_int(
    issues: list[Issue],
    row_id: object,
    name: str,
    field: str,
    value: object,
    *,
    required: bool,
) -> int | None:
    if is_empty(value):
        if required:
            add(issues, row_id, name, field, f"{field} 不能为空")
        return None
    parsed = parse_int(value)
    if parsed is None:
        add(issues, row_id, name, field, f"{field} 必须是 int，实际: {value}")
    return parsed


def validate_structural(path: Path, map_path: Path | None = None) -> list[Issue]:
    data = load_rows(path)
    issues: list[Issue] = []
    ids: list[int] = []

    resolved_map = resolve_map_path(path, map_path)
    id_map: dict[str, set[int]] | None = None
    if not resolved_map.is_file():
        add(
            issues,
            "",
            "",
            "Id",
            f"缺少映射表，无法校验充值 Id 对应: {resolved_map}",
        )
    else:
        try:
            id_map = load_recharge_id_map(resolved_map)
        except Exception as exc:  # noqa: BLE001
            add(
                issues,
                "",
                "",
                "Id",
                f"读取映射表失败 ({resolved_map}): {exc}",
            )

    for _, row in data.iterrows():
        raw_id = row.get("Id")
        name = "" if is_empty(row.get("Name")) else str(row.get("Name")).strip()

        row_id = parse_int(raw_id)
        if row_id is None:
            add(issues, raw_id, name, "Id", f"Id 必须是 int，实际: {raw_id}")
            continue
        ids.append(row_id)

        if not name:
            add(issues, row_id, name, "Name", "Name 必须是 string 且非空")

        rtype = "" if is_empty(row.get("RechargeType")) else str(row.get("RechargeType")).strip()
        if not rtype:
            add(issues, row_id, name, "RechargeType", "RechargeType 不能为空")
        elif rtype not in RECHARGE_TYPES:
            add(
                issues,
                row_id,
                name,
                "RechargeType",
                f"RechargeType 不在枚举内: {rtype}",
            )

        require_int(issues, row_id, name, "RelateId", row.get("RelateId"), required=True)

        product_id = row.get("ProductId")
        if is_empty(product_id):
            add(issues, row_id, name, "ProductId", "ProductId 不能为空")
        else:
            if has_edge_whitespace(product_id):
                add(
                    issues,
                    row_id,
                    name,
                    "ProductId",
                    "ProductId 不得含首尾空白或换行",
                )
            pid_text = str(product_id).strip()
            if not PRODUCT_ID_RE.match(pid_text):
                add(
                    issues,
                    row_id,
                    name,
                    "ProductId",
                    f"ProductId 格式应为 qudao_<数字>.<数字>，实际: {pid_text}",
                )

        platform = "" if is_empty(row.get("Platform")) else str(row.get("Platform")).strip()
        if not platform:
            add(issues, row_id, name, "Platform", "Platform 不能为空")
        elif platform not in PLATFORMS:
            add(issues, row_id, name, "Platform", f"Platform 不在枚举内: {platform}")
        elif id_map is not None:
            mapped_ids = id_map.get(platform, set())
            if row_id not in mapped_ids:
                col = PLATFORM_MAP_COL[platform]
                add(
                    issues,
                    row_id,
                    name,
                    "Id",
                    f"Id={row_id}（Platform={platform}）未在映射表 {col} 列中找到",
                )

        require_int(issues, row_id, name, "Channel", row.get("Channel"), required=False)

        price = parse_float(row.get("Price"))
        if price is None:
            add(issues, row_id, name, "Price", f"Price 必须是数值且非空，实际: {row.get('Price')}")
        elif price < 0:
            add(issues, row_id, name, "Price", f"Price 不能为负: {price}")
        else:
            m_price = NAME_PRICE_RE.match(name)
            if m_price and price != float(m_price.group(1)):
                add(
                    issues,
                    row_id,
                    name,
                    "Price",
                    f"Name 以 {m_price.group(1)}元 开头时 Price 必须为 {m_price.group(1)}，实际: {price}",
                )
            m_gold = GOLD_LIANG_RE.match(name)
            if m_gold:
                section = str(row.get("_Section") or "")
                # #支付中心黄金 分区内黄金档价格不定规则，跳过
                if "支付中心黄金" not in section:
                    expect = int(m_gold.group(1)) / 10.0
                    if abs(price - expect) > 1e-6:
                        add(
                            issues,
                            row_id,
                            name,
                            "Price",
                            f"Name 为「N两黄金」时 Price 必须为 N/10={expect:g}，实际: {price}",
                        )

        require_int(
            issues, row_id, name, "BaseRelateId", row.get("BaseRelateId"), required=False
        )
        if not is_empty(row.get("OldPrice")) and parse_float(row.get("OldPrice")) is None:
            add(issues, row_id, name, "OldPrice", f"OldPrice 必须是数值，实际: {row.get('OldPrice')}")
        require_int(issues, row_id, name, "Discount", row.get("Discount"), required=False)
        require_int(
            issues, row_id, name, "RechargeMulti", row.get("RechargeMulti"), required=False
        )

        on_raw = row.get("OnShelfTime")
        off_raw = row.get("OffShelfTime")
        on_empty = is_empty(on_raw)
        off_empty = is_empty(off_raw)
        if on_empty != off_empty:
            add(
                issues,
                row_id,
                name,
                "OnShelfTime" if on_empty else "OffShelfTime",
                "OnShelfTime 与 OffShelfTime 须同时为空或同时有值",
            )
        elif not on_empty and not off_empty:
            on_text = str(on_raw).strip()
            off_text = str(off_raw).strip()
            on_ok = DATE_RE.match(on_text)
            off_ok = DATE_RE.match(off_text)
            if not on_ok:
                add(
                    issues,
                    row_id,
                    name,
                    "OnShelfTime",
                    f"格式应为 {DATE_FMT}，实际: {on_text}",
                )
            if not off_ok:
                add(
                    issues,
                    row_id,
                    name,
                    "OffShelfTime",
                    f"格式应为 {DATE_FMT}，实际: {off_text}",
                )
            if on_ok and off_ok:
                on_dt = datetime.strptime(on_text, DATE_FMT)
                off_dt = datetime.strptime(off_text, DATE_FMT)
                if on_dt > off_dt:
                    add(
                        issues,
                        row_id,
                        name,
                        "OnShelfTime",
                        f"上架时间不能晚于下架时间: {on_text} > {off_text}",
                    )

        limit_type_raw = row.get("LimitType")
        limit_count_raw = row.get("LimitCount")
        lt_empty = is_empty(limit_type_raw)
        lc_empty = is_empty(limit_count_raw)
        if lt_empty != lc_empty:
            add(
                issues,
                row_id,
                name,
                "LimitType" if lt_empty else "LimitCount",
                "LimitType 与 LimitCount 须同时为空或同时有值",
            )
        else:
            if not lt_empty:
                lt = parse_int(limit_type_raw)
                if lt is None:
                    add(
                        issues,
                        row_id,
                        name,
                        "LimitType",
                        f"LimitType 必须是 int，实际: {limit_type_raw}",
                    )
                elif lt not in LIMIT_TYPES:
                    add(
                        issues,
                        row_id,
                        name,
                        "LimitType",
                        f"LimitType 不在枚举 {{1,2,3,4}}，实际: {lt}",
                    )
            if not lc_empty:
                require_int(
                    issues, row_id, name, "LimitCount", limit_count_raw, required=True
                )

        for field in ("Depend", "Mutex"):
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

        require_int(
            issues, row_id, name, "RechargeGroup", row.get("RechargeGroup"), required=False
        )
        require_int(
            issues, row_id, name, "GiftGoodID", row.get("GiftGoodID"), required=False
        )

    # 表独有：赛季礼包档位（依赖同季豪华 RelateId / Depend，先建索引）
    luxury_relate_by_season: dict[str, set[int]] = {}
    luxury_id_by_season_plat: dict[str, dict[str, int]] = {}
    for _, row in data.iterrows():
        name = "" if is_empty(row.get("Name")) else str(row.get("Name")).strip()
        m = SEASON_PACK_RE.match(name)
        if not m or m.group(2) != "豪华版礼包":
            continue
        season = m.group(1)
        relate = parse_int(row.get("RelateId"))
        if relate is not None:
            luxury_relate_by_season.setdefault(season, set()).add(relate)
        lux_id = parse_int(row.get("Id"))
        platform = (
            "" if is_empty(row.get("Platform")) else str(row.get("Platform")).strip()
        )
        if lux_id is not None and platform in ("andriod", "pc", "ios"):
            luxury_id_by_season_plat.setdefault(season, {})[platform] = lux_id

    def expected_upgrade_depend(season: str) -> str | None:
        plats = luxury_id_by_season_plat.get(season, {})
        ordered = [plats.get(p) for p in ("andriod", "pc", "ios")]
        if any(x is None for x in ordered):
            return None
        return ",".join(str(x) for x in ordered)

    def depend_text(value: object) -> str | None:
        if is_empty(value):
            return None
        return str(value).strip().replace(" ", "")

    season_pack_rows: list[dict] = []
    for _, row in data.iterrows():
        raw_id = row.get("Id")
        row_id = parse_int(raw_id)
        if row_id is None:
            continue
        name = "" if is_empty(row.get("Name")) else str(row.get("Name")).strip()
        m = SEASON_PACK_RE.match(name)
        if not m:
            continue
        season, kind = m.group(1), m.group(2)
        season_n = int(season)
        price = parse_float(row.get("Price"))
        base = parse_int(row.get("BaseRelateId"))
        lux_relates = luxury_relate_by_season.get(season, set())
        dep = depend_text(row.get("Depend"))

        on_s = shelf_time_str(row.get("OnShelfTime"))
        off_s = shelf_time_str(row.get("OffShelfTime"))
        on_ym: tuple[int, int] | None = None
        if on_s is None:
            add(
                issues,
                row_id,
                name,
                "OnShelfTime",
                f"{name} 的 OnShelfTime 不能为空",
            )
        else:
            on_m = SEASON_ON_RE.match(on_s)
            if not on_m:
                add(
                    issues,
                    row_id,
                    name,
                    "OnShelfTime",
                    f"赛季礼包 OnShelfTime 必须为每月 15 日，实际: {on_s}",
                )
            else:
                on_ym = (int(on_m.group(1)), int(on_m.group(2)))
                # S1 不校 00:00:00，仅禁止写成未来时间；S2+ 必须 15 日 00:00:00
                if season_n == 1:
                    try:
                        on_dt = datetime.strptime(on_s, DATE_FMT)
                        if on_dt > datetime.now():
                            add(
                                issues,
                                row_id,
                                name,
                                "OnShelfTime",
                                f"S1 赛季礼包 OnShelfTime 不能为未来时间，实际: {on_s}",
                            )
                    except ValueError:
                        pass
                elif (on_m.group(3), on_m.group(4), on_m.group(5)) != ("00", "00", "00"):
                    add(
                        issues,
                        row_id,
                        name,
                        "OnShelfTime",
                        f"赛季礼包 OnShelfTime 必须为 15 日 00:00:00，实际: {on_s}",
                    )
        if off_s is None:
            add(
                issues,
                row_id,
                name,
                "OffShelfTime",
                f"{name} 的 OffShelfTime 不能为空",
            )
        elif not SEASON_OFF_RE.match(off_s):
            add(
                issues,
                row_id,
                name,
                "OffShelfTime",
                f"赛季礼包 OffShelfTime 必须为 YYYY-MM-14 23:59:59，实际: {off_s}",
            )

        season_pack_rows.append(
            {
                "row_id": row_id,
                "name": name,
                "season_n": season_n,
                "on_s": on_s,
                "on_ym": on_ym,
            }
        )

        if kind == "豪华版礼包":
            if price is not None and price != 38:
                add(
                    issues,
                    row_id,
                    name,
                    "Price",
                    f"{name} 的 Price 必须为 38，实际: {price}",
                )
            if not is_empty(row.get("BaseRelateId")):
                add(
                    issues,
                    row_id,
                    name,
                    "BaseRelateId",
                    f"{name} 的 BaseRelateId 应为空",
                )
            if dep is not None:
                add(
                    issues,
                    row_id,
                    name,
                    "Depend",
                    f"{name} 的 Depend 应为空，实际: {row.get('Depend')}",
                )
            continue

        # 典藏版礼包
        if is_empty(row.get("BaseRelateId")):
            if price is not None and price != 128:
                add(
                    issues,
                    row_id,
                    name,
                    "Price",
                    f"{name} 直购档（BaseRelateId 空）Price 必须为 128，实际: {price}",
                )
            if dep is not None:
                add(
                    issues,
                    row_id,
                    name,
                    "Depend",
                    f"{name} 直购档 Depend 应为空，实际: {row.get('Depend')}",
                )
        else:
            if price is not None and price != 90:
                add(
                    issues,
                    row_id,
                    name,
                    "Price",
                    f"{name} 升级档（BaseRelateId 有值）Price 必须为 90，实际: {price}",
                )
            if base is None:
                add(
                    issues,
                    row_id,
                    name,
                    "BaseRelateId",
                    f"{name} BaseRelateId 必须是 int，实际: {row.get('BaseRelateId')}",
                )
            elif not lux_relates:
                add(
                    issues,
                    row_id,
                    name,
                    "BaseRelateId",
                    f"缺少同季 S{season}豪华版礼包，无法校验 BaseRelateId",
                )
            elif base not in lux_relates:
                add(
                    issues,
                    row_id,
                    name,
                    "BaseRelateId",
                    f"应等于 S{season}豪华版礼包 的 RelateId {sorted(lux_relates)}，实际: {base}",
                )
            expect_dep = expected_upgrade_depend(season)
            if expect_dep is None:
                add(
                    issues,
                    row_id,
                    name,
                    "Depend",
                    f"缺少同季 S{season} 三端豪华版礼包 Id，无法校验 Depend",
                )
            elif dep != expect_dep:
                add(
                    issues,
                    row_id,
                    name,
                    "Depend",
                    f"升级档 Depend 必须为同季三端豪华 Id（andriod,pc,ios）"
                    f"={expect_dep}，实际: {row.get('Depend')}",
                )

    # 期数 ↔ 上架年月：升序推演 S1..Sn，且一对一不可重复
    season_to_ym: dict[int, set[tuple[int, int]]] = {}
    ym_to_season: dict[tuple[int, int], set[int]] = {}
    for item in season_pack_rows:
        ym = item["on_ym"]
        if ym is None:
            continue
        sn = item["season_n"]
        season_to_ym.setdefault(sn, set()).add(ym)
        ym_to_season.setdefault(ym, set()).add(sn)

    for sn, yms in sorted(season_to_ym.items()):
        if len(yms) > 1:
            sample = next(i for i in season_pack_rows if i["season_n"] == sn)
            add(
                issues,
                sample["row_id"],
                sample["name"],
                "OnShelfTime",
                f"S{sn} 期上架年月不一致（不可重复对应多月）: "
                f"{sorted(f'{y}-{m:02d}' for y, m in yms)}",
            )

    for ym, sns in sorted(ym_to_season.items()):
        if len(sns) > 1:
            sample = next(i for i in season_pack_rows if i["on_ym"] == ym)
            add(
                issues,
                sample["row_id"],
                sample["name"],
                "Name",
                f"上架年月 {ym[0]}-{ym[1]:02d} 对应多个期数 {sorted(sns)}（不可重复）",
            )

    ordered_yms = sorted(ym_to_season.keys())
    expected_by_ym = {ym: idx + 1 for idx, ym in enumerate(ordered_yms)}
    reported_mismatch: set[tuple[int, int, int]] = set()
    for item in season_pack_rows:
        ym = item["on_ym"]
        if ym is None:
            continue
        expect_n = expected_by_ym[ym]
        if item["season_n"] != expect_n:
            key = (item["season_n"], expect_n, ym[0] * 100 + ym[1])
            if key in reported_mismatch:
                continue
            reported_mismatch.add(key)
            add(
                issues,
                item["row_id"],
                item["name"],
                "Name",
                f"按上架年月升序推演应为 S{expect_n}（{ym[0]}-{ym[1]:02d}），"
                f"名称期数为 S{item['season_n']}",
            )

    # 表独有：赛季皮肤礼盒（不校外联 / 一期皮肤数 / 命名变体）
    skin_rows: list[dict] = []
    for _, row in data.iterrows():
        raw_id = row.get("Id")
        row_id = parse_int(raw_id)
        if row_id is None:
            continue
        name = "" if is_empty(row.get("Name")) else str(row.get("Name")).strip()
        if name != SEASON_SKIN_BOX_NAME:
            continue

        rtype = (
            "" if is_empty(row.get("RechargeType")) else str(row.get("RechargeType")).strip()
        )
        if rtype != "PickGoods":
            add(
                issues,
                row_id,
                name,
                "RechargeType",
                f"{SEASON_SKIN_BOX_NAME} 的 RechargeType 必须为 PickGoods，实际: {rtype or '(空)'}",
            )

        price = parse_float(row.get("Price"))
        if price is None or price != 128:
            add(
                issues,
                row_id,
                name,
                "Price",
                f"{SEASON_SKIN_BOX_NAME} 的 Price 必须为 128，实际: {row.get('Price')}",
            )

        base = require_int(
            issues, row_id, name, "BaseRelateId", row.get("BaseRelateId"), required=True
        )
        relate = parse_int(row.get("RelateId"))
        group = require_int(
            issues, row_id, name, "RechargeGroup", row.get("RechargeGroup"), required=True
        )

        limit_type = parse_int(row.get("LimitType"))
        if limit_type != 1:
            add(
                issues,
                row_id,
                name,
                "LimitType",
                f"{SEASON_SKIN_BOX_NAME} 的 LimitType 必须为 1，实际: {row.get('LimitType')}",
            )
        limit_count = parse_int(row.get("LimitCount"))
        if limit_count != 1:
            add(
                issues,
                row_id,
                name,
                "LimitCount",
                f"{SEASON_SKIN_BOX_NAME} 的 LimitCount 必须为 1，实际: {row.get('LimitCount')}",
            )

        for field in SEASON_SKIN_BOX_EMPTY_FIELDS:
            if not is_empty(row.get(field)):
                add(
                    issues,
                    row_id,
                    name,
                    field,
                    f"{SEASON_SKIN_BOX_NAME} 的 {field} 应为空，实际: {row.get(field)}",
                )

        platform = (
            "" if is_empty(row.get("Platform")) else str(row.get("Platform")).strip()
        )
        on_shelf = (
            None
            if is_empty(row.get("OnShelfTime"))
            else str(row.get("OnShelfTime")).strip()
        )
        off_shelf = (
            None
            if is_empty(row.get("OffShelfTime"))
            else str(row.get("OffShelfTime")).strip()
        )
        skin_rows.append(
            {
                "row_id": row_id,
                "name": name,
                "platform": platform,
                "relate": relate,
                "base": base,
                "group": group,
                "on": on_shelf,
                "off": off_shelf,
            }
        )

    by_relate: dict[int, list[dict]] = {}
    for item in skin_rows:
        if item["relate"] is None:
            continue
        by_relate.setdefault(item["relate"], []).append(item)

    for relate, items in by_relate.items():
        on_vals = {i["on"] for i in items}
        off_vals = {i["off"] for i in items}
        if len(on_vals) > 1 or len(off_vals) > 1:
            sample = items[0]
            add(
                issues,
                sample["row_id"],
                sample["name"],
                "OnShelfTime",
                f"RelateId={relate} 组内 On/OffShelfTime 必须一致，"
                f"On={sorted(str(x) for x in on_vals)} Off={sorted(str(x) for x in off_vals)}",
            )

        android_ids = [
            i["row_id"] for i in items if i["platform"] == "andriod" and i["row_id"] is not None
        ]
        if android_ids:
            expect_group = min(android_ids)
            for i in items:
                if i["group"] is not None and i["group"] != expect_group:
                    add(
                        issues,
                        i["row_id"],
                        i["name"],
                        "RechargeGroup",
                        f"RelateId={relate} 组内 RechargeGroup 必须等于 andriod 最小 Id "
                        f"{expect_group}，实际: {i['group']}",
                    )

    by_sku: dict[tuple[int, int], dict[str, list[dict]]] = {}
    for item in skin_rows:
        if item["relate"] is None or item["base"] is None:
            continue
        key = (item["relate"], item["base"])
        by_sku.setdefault(key, {}).setdefault(item["platform"], []).append(item)

    for (relate, base), by_plat in by_sku.items():
        for platform in SEASON_SKIN_BOX_PLATFORMS:
            rows_here = by_plat.get(platform, [])
            if len(rows_here) == 0:
                # 挂到同组任意已有行上报；若三端皆缺则无法挂（不应发生）
                sample = next((rs[0] for rs in by_plat.values() if rs), None)
                if sample is not None:
                    add(
                        issues,
                        sample["row_id"],
                        sample["name"],
                        "Platform",
                        f"RelateId={relate} BaseRelateId={base} 缺少 Platform={platform} 行",
                    )
            elif len(rows_here) > 1:
                for i in rows_here:
                    add(
                        issues,
                        i["row_id"],
                        i["name"],
                        "Platform",
                        f"RelateId={relate} BaseRelateId={base} Platform={platform} "
                        f"应恰有 1 行，实际 {len(rows_here)} 行",
                    )

    for duplicate_id, count in Counter(ids).items():
        if count > 1:
            add(issues, duplicate_id, "", "Id", f"Id 重复，出现 {count} 次")

    return issues


def build_report(path: Path, map_path: Path | None = None) -> dict:
    structural_issues = validate_structural(path, map_path=map_path)
    return {
        "file": str(path),
        "map_file": str(resolve_map_path(path, map_path)),
        "structural_issue_count": len(structural_issues),
        "semantic_row_count": 0,
        "structural_issues": [asdict(issue) for issue in structural_issues],
        "semantic_rows": [],
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="校验 Recharge_充值表.xlsx")
    parser.add_argument("path", help="Recharge_充值表.xlsx 路径")
    parser.add_argument(
        "--map",
        dest="map_path",
        default=None,
        help="Recharge_映射表.xlsx 路径（默认：与充值表同目录）",
    )
    parser.add_argument("--json", action="store_true", help="输出完整 JSON")
    parser.add_argument(
        "--semantic-json",
        action="store_true",
        help="本表无语义阶段；输出空数组以兼容 CLI",
    )
    args = parser.parse_args()

    path = Path(args.path)
    if not path.is_file():
        print(f"文件不存在: {path}", file=sys.stderr)
        return 2

    map_path = Path(args.map_path) if args.map_path else None

    if args.json:
        print(json.dumps(build_report(path, map_path=map_path), ensure_ascii=False, indent=2))
        return 0

    if args.semantic_json:
        print("[]")
        return 0

    issues = validate_structural(path, map_path=map_path)
    print(f"检查文件: {path}")
    print(f"映射表: {resolve_map_path(path, map_path)}")
    print(f"结构化问题数量: {len(issues)}")
    for issue in issues:
        print(issue.format_line())
    print("待语义分析行数: 0（本表仅结构化）")
    return 1 if issues else 0


if __name__ == "__main__":
    sys.exit(main())
