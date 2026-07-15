# -*- coding: utf-8 -*-
"""校验 Mail.xlsx：脚本负责结构化规则，语义匹配由 LLM 分析。"""
from __future__ import annotations

import argparse
import json
import re
import sys
from collections import Counter, defaultdict
from dataclasses import asdict, dataclass
from pathlib import Path

import pandas as pd

BODY_PREFIX = "<color=#FFFFFF00>占位</color>"
BODY_PREFIX_LEN = 200
ITEM_PATTERN = re.compile(r"^(\{\d+;\d+\})(,\{\d+;\d+\})*$")
SHANHE_TITLES = {"山河争锋守城将来信", "山河争锋攻城将来信"}
MELT_TITLE = "熔炼武将道具返还"
DATA_START_ROW = 5
HEADER_ROW = 2


@dataclass
class MailIssue:
    row_id: object
    title: str
    field: str
    message: str
    source: str = "structural"

    def format_line(self) -> str:
        title = self.title or "-"
        rid = self.row_id if self.row_id != "" else "-"
        return f"Id={rid} | Title={title} | {self.field} | {self.message}"


@dataclass
class SemanticRow:
    id: int
    title: str
    sender: str
    body_preview: str
    category: str

    def to_dict(self) -> dict:
        return asdict(self)


def is_empty(value: object) -> bool:
    if pd.isna(value):
        return True
    text = str(value).strip()
    return text in ("", "nan", "NaN")


def is_warrior_letter(title: str) -> bool:
    return title.endswith("来信") and title not in SHANHE_TITLES


def is_warrior_mail(title: str) -> bool:
    return title == MELT_TITLE or is_warrior_letter(title)


def needs_semantic_review(title: str) -> bool:
    return is_warrior_mail(title)


def expected_sender_type(title: str) -> int:
    if is_warrior_mail(title):
        return 2
    return 1


def expected_send_cond_type(title: str) -> int:
    if title == "举报成功":
        return 4
    if title == "被举报通知":
        return 5
    if is_warrior_letter(title):
        return 3
    return 0


def expected_sender(title: str) -> str | None:
    if title in ("举报成功", "被举报通知"):
        return "系统"
    if title == "山河争锋守城将来信":
        return "守城将"
    if title == "山河争锋攻城将来信":
        return "攻城将"
    if is_warrior_mail(title):
        return None
    return "名将杀运营团队"


def expected_receiver_type(receiver: object) -> int | None:
    if is_empty(receiver):
        return 1
    text = str(receiver).strip()
    if not text.isdigit():
        return None
    if len(text) <= 3:
        return 2
    if len(text) <= 5:
        return 3
    return None


def body_preview(body: object) -> str:
    if is_empty(body):
        return ""
    text = str(body).strip()
    if text.startswith(BODY_PREFIX):
        text = text[len(BODY_PREFIX) :]
    text = re.sub(r"\s+", " ", text)
    return text[:BODY_PREFIX_LEN]


def is_string_value(value: object) -> bool:
    return isinstance(value, str)


def validate_send_cond(
    row_id: int,
    title: str,
    send_cond_type: int | None,
    send_cond: object,
    issues: list[MailIssue],
) -> None:
    if send_cond_type == 0:
        if not is_empty(send_cond):
            issues.append(
                MailIssue(
                    row_id,
                    title,
                    "SendCond",
                    "SendCondType=0 时 SendCond 应为空",
                )
            )
        return

    if send_cond_type is None:
        return

    if is_empty(send_cond):
        if send_cond_type in (4, 5):
            issues.append(
                MailIssue(
                    row_id,
                    title,
                    "SendCond",
                    f"SendCondType={send_cond_type} 时 SendCond 应为 1 或 2",
                )
            )
        elif send_cond_type == 3:
            issues.append(
                MailIssue(
                    row_id,
                    title,
                    "SendCond",
                    "SendCondType=3 时 SendCond 应为 7 位道具 id（string）",
                )
            )
        return

    if not is_string_value(send_cond):
        issues.append(
            MailIssue(
                row_id,
                title,
                "SendCond",
                f"SendCond 必须是 string 类型，实际为 {type(send_cond).__name__}（值: {send_cond}）",
            )
        )
        return

    text = send_cond.strip()
    if send_cond_type in (4, 5):
        if text not in ("1", "2"):
            issues.append(
                MailIssue(
                    row_id,
                    title,
                    "SendCond",
                    f"SendCondType={send_cond_type} 时 SendCond 应为 1 或 2",
                )
            )
    elif send_cond_type == 3 and not re.fullmatch(r"\d{7}", text):
        issues.append(
            MailIssue(
                row_id,
                title,
                "SendCond",
                "SendCondType=3 时 SendCond 应为 7 位道具 id（string）",
            )
        )


def load_mail_rows(path: Path) -> pd.DataFrame:
    raw = pd.read_excel(path, header=None)
    columns = raw.iloc[HEADER_ROW].tolist()
    data = raw.iloc[DATA_START_ROW:].copy()
    data.columns = columns
    data = data[data["Id"].notna() & (data["Id"].astype(str) != "#")]
    return data.reset_index(drop=True)


def collect_semantic_rows(data: pd.DataFrame) -> list[SemanticRow]:
    rows: list[SemanticRow] = []
    for _, row in data.iterrows():
        title = "" if is_empty(row["Title"]) else str(row["Title"]).strip()
        if not needs_semantic_review(title):
            continue
        try:
            row_id = int(row["Id"])
        except (TypeError, ValueError):
            continue
        sender = "" if is_empty(row["Sender"]) else str(row["Sender"]).strip()
        category = "melt_return" if title == MELT_TITLE else "warrior_letter"
        rows.append(
            SemanticRow(
                id=row_id,
                title=title,
                sender=sender,
                body_preview=body_preview(row["Body"]),
                category=category,
            )
        )
    return rows


def validate_structural(path: Path) -> tuple[list[MailIssue], pd.DataFrame]:
    data = load_mail_rows(path)
    issues: list[MailIssue] = []
    ids: list[int] = []

    title_sender_table: dict[str, set[str]] = defaultdict(set)
    for _, row in data.iterrows():
        title = "" if is_empty(row["Title"]) else str(row["Title"]).strip()
        if not is_warrior_mail(title):
            continue
        sender = "" if is_empty(row["Sender"]) else str(row["Sender"]).strip()
        if sender:
            title_sender_table[title].add(sender)

    for title, senders in title_sender_table.items():
        if len(senders) > 1:
            issues.append(
                MailIssue(
                    "",
                    title,
                    "Sender",
                    f"同一 Title 对应多个 Sender: {', '.join(sorted(senders))}",
                )
            )

    for _, row in data.iterrows():
        raw_id = row["Id"]
        title = "" if is_empty(row["Title"]) else str(row["Title"]).strip()

        try:
            row_id = int(raw_id)
            ids.append(row_id)
        except (TypeError, ValueError):
            issues.append(MailIssue(raw_id, title, "Id", "Id 必须是 int"))
            continue

        if not title:
            issues.append(MailIssue(row_id, title, "Title", "Title 必须是 string 且非空"))

        body = row["Body"]
        if is_empty(body):
            issues.append(MailIssue(row_id, title, "Body", "Body 不能为空"))
        elif not str(body).startswith(BODY_PREFIX):
            issues.append(
                MailIssue(
                    row_id,
                    title,
                    "Body",
                    f"Body 必须以 {BODY_PREFIX} 开头并拼接后续文本",
                )
            )

        if not is_empty(row["SenderType"]):
            actual = int(row["SenderType"])
            expected = expected_sender_type(title)
            if actual != expected:
                issues.append(
                    MailIssue(
                        row_id,
                        title,
                        "SenderType",
                        f"期望 {expected}，实际 {actual}",
                    )
                )

        actual_sender = "" if is_empty(row["Sender"]) else str(row["Sender"]).strip()
        expected_sender_value = expected_sender(title)

        if expected_sender_value is not None:
            if actual_sender != expected_sender_value:
                issues.append(
                    MailIssue(
                        row_id,
                        title,
                        "Sender",
                        f"期望 {expected_sender_value}，实际 {actual_sender or '(空)'}",
                    )
                )
        elif is_warrior_mail(title) and is_empty(actual_sender):
            issues.append(
                MailIssue(
                    row_id,
                    title,
                    "Sender",
                    "武将来信/熔炼武将道具返还时 Sender 不能为空（具体武将名由语义校验确认）",
                )
            )

        expected_rt = expected_receiver_type(row["Receiver"])
        if expected_rt is None:
            issues.append(
                MailIssue(
                    row_id,
                    title,
                    "Receiver",
                    f"Receiver 参数无法识别: {row['Receiver']}",
                )
            )
        elif not is_empty(row["ReceiverType"]) and int(row["ReceiverType"]) != expected_rt:
            issues.append(
                MailIssue(
                    row_id,
                    title,
                    "ReceiverType",
                    f"期望 {expected_rt}，实际 {row['ReceiverType']}",
                )
            )

        expected_sct = expected_send_cond_type(title)
        if not is_empty(row["SendCondType"]) and int(row["SendCondType"]) != expected_sct:
            issues.append(
                MailIssue(
                    row_id,
                    title,
                    "SendCondType",
                    f"期望 {expected_sct}，实际 {row['SendCondType']}",
                )
            )

        send_cond_type = (
            int(row["SendCondType"]) if not is_empty(row["SendCondType"]) else None
        )
        validate_send_cond(row_id, title, send_cond_type, row["SendCond"], issues)

        if not is_empty(row["Item"]):
            item_text = str(row["Item"]).strip()
            if not ITEM_PATTERN.match(item_text):
                issues.append(
                    MailIssue(
                        row_id,
                        title,
                        "Item",
                        f"Item 格式应为 {{道具id;数量}} 或 {{a;b}},{{c;d}}，实际: {item_text}",
                    )
                )

    for duplicate_id, count in Counter(ids).items():
        if count > 1:
            issues.append(
                MailIssue(duplicate_id, "", "Id", f"Id 重复，出现 {count} 次")
            )

    return issues, data


def build_report(path: Path) -> dict:
    structural_issues, data = validate_structural(path)
    semantic_rows = collect_semantic_rows(data)
    return {
        "file": str(path),
        "structural_issue_count": len(structural_issues),
        "semantic_row_count": len(semantic_rows),
        "structural_issues": [asdict(issue) for issue in structural_issues],
        "semantic_rows": [row.to_dict() for row in semantic_rows],
    }


def print_structural_issues(issues: list[MailIssue]) -> None:
    print(f"结构化问题数量: {len(issues)}")
    for issue in issues:
        print(issue.format_line())


def main() -> int:
    parser = argparse.ArgumentParser(description="校验 Mail.xlsx 配置表")
    parser.add_argument(
        "path",
        help="Mail.xlsx 路径",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        help="输出完整 JSON（结构化问题 + 待语义分析行）",
    )
    parser.add_argument(
        "--semantic-json",
        action="store_true",
        help="仅输出待 LLM 语义分析的行（JSON）",
    )
    args = parser.parse_args()

    path = Path(args.path)
    if not path.is_file():
        print(f"文件不存在: {path}", file=sys.stderr)
        return 2

    if args.json:
        print(json.dumps(build_report(path), ensure_ascii=False, indent=2))
        return 0

    structural_issues, data = validate_structural(path)

    if args.semantic_json:
        rows = [row.to_dict() for row in collect_semantic_rows(data)]
        print(json.dumps(rows, ensure_ascii=False, indent=2))
        return 0

    print(f"检查文件: {path}")
    print_structural_issues(structural_issues)
    semantic_count = len(collect_semantic_rows(data))
    print(f"待语义分析行数: {semantic_count}（见 skill 流程，由 LLM 逐行分析 Title/Sender 是否同一人物）")
    return 1 if structural_issues else 0


if __name__ == "__main__":
    sys.exit(main())
