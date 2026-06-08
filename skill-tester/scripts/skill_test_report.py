#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
skill-tester 统一命令行入口（所有技能共用，禁止为单个技能再写 gen_report.py）。

子命令:
  export       将 report_data.json 导出为 Excel（默认删除临时 JSON）
  parse-cases  从 testcase_{技能名}.md 导出用例列表（JSON，供 Agent 核对）
  list-skills  列出已有 testcase_*.md 的技能目录
  find-dir     打印技能目录绝对路径（解决 Windows 中文路径）

示例（在仓库根目录）:
  python .cursor/skills/skill-tester/scripts/skill_test_report.py export --skill-name "宫" --input .cursor/skills/skill-tester/宫/report_data.json
  python .cursor/skills/skill-tester/scripts/skill_test_report.py export --skill-name "宫" --stdin < report_data.json
"""
from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

# 保证同目录模块可导入
sys.path.insert(0, str(Path(__file__).resolve().parent))

from export_test_report import export_xlsx
from skill_test_common import (
    find_skill_dir,
    list_skill_dirs,
    load_report_json,
    parse_test_md_cases,
    parse_test_md_overview,
    resolve_testcase_md,
    write_report_json,
)


def _resolve_skill_dir(args: argparse.Namespace) -> Path:
    if getattr(args, "skill_dir", None):
        return Path(args.skill_dir).resolve()
    if getattr(args, "skill_name", None):
        return find_skill_dir(args.skill_name)
    raise SystemExit("必须指定 --skill-name 或 --skill-dir")


def cmd_list_skills(_: argparse.Namespace) -> None:
    for d in list_skill_dirs():
        overview = parse_test_md_overview(resolve_testcase_md(d))
        print(f"{d.name}\t{overview.get('skill_name', '')}\t{overview.get('hero', '')}")


def cmd_find_dir(args: argparse.Namespace) -> None:
    print(_resolve_skill_dir(args))


def cmd_parse_cases(args: argparse.Namespace) -> None:
    skill_dir = _resolve_skill_dir(args)
    testcase_md = resolve_testcase_md(skill_dir)
    cases = parse_test_md_cases(testcase_md)
    payload = {
        "skill_dir": str(skill_dir),
        "testcase_file": testcase_md.name,
        "overview": parse_test_md_overview(testcase_md),
        "cases": cases,
    }
    text = json.dumps(payload, ensure_ascii=False, indent=2 if args.pretty else None)
    if args.output:
        Path(args.output).write_text(text or "{}", encoding="utf-8")
        print(args.output)
    else:
        print(text)


def cmd_export(args: argparse.Namespace) -> None:
    skill_dir = _resolve_skill_dir(args)

    if args.stdin:
        data = load_report_json(None, json.load(sys.stdin))
    elif args.input:
        inp = Path(args.input)
        data = load_report_json(inp)
    else:
        raise SystemExit("export 需要 --input <file> 或 --stdin")

    temp_json = skill_dir / "report_data.json"
    wrote_temp = False
    if args.stdin or (args.input and Path(args.input).resolve() != temp_json.resolve()):
        write_report_json(skill_dir, data)
        wrote_temp = True
        json_path = temp_json
    else:
        json_path = Path(args.input)

    out = Path(args.output) if args.output else None
    xlsx_path = export_xlsx(skill_dir, data, out)
    print(xlsx_path)

    if wrote_temp and not args.keep_json:
        temp_json.unlink(missing_ok=True)
    elif args.keep_json and wrote_temp:
        print(f"保留临时 JSON: {json_path}", file=sys.stderr)


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="skill-tester 测试报告工具（通用）")
    sub = p.add_subparsers(dest="command", required=True)

    def add_skill_args(sp: argparse.ArgumentParser) -> None:
        g = sp.add_mutually_exclusive_group(required=True)
        g.add_argument("--skill-name", help="技能名（文件夹名或 testcase 内技能名称）")
        g.add_argument("--skill-dir", help="技能目录路径")

    sp = sub.add_parser("list-skills", help="列出所有含 testcase_*.md 的技能")
    sp.set_defaults(func=cmd_list_skills)

    sp = sub.add_parser("find-dir", help="输出技能目录绝对路径")
    add_skill_args(sp)
    sp.set_defaults(func=cmd_find_dir)

    sp = sub.add_parser("parse-cases", help="解析 testcase_{技能名}.md 用例为 JSON")
    add_skill_args(sp)
    sp.add_argument("--output", "-o", help="写入文件（默认 stdout）")
    sp.add_argument("--pretty", action="store_true", help="缩进 JSON")
    sp.set_defaults(func=cmd_parse_cases)

    sp = sub.add_parser("export", help="导出 Excel 测试报告")
    add_skill_args(sp)
    src = sp.add_mutually_exclusive_group(required=True)
    src.add_argument("--input", "-i", help="report_data.json 路径")
    src.add_argument("--stdin", action="store_true", help="从标准输入读取 JSON")
    sp.add_argument("--output", "-o", help="xlsx 输出路径")
    sp.add_argument(
        "--keep-json",
        action="store_true",
        help="保留技能目录下的 report_data.json（默认导出后删除临时文件）",
    )
    sp.set_defaults(func=cmd_export)

    return p


def main() -> None:
    parser = build_parser()
    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
