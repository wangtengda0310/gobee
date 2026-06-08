# -*- coding: utf-8 -*-
"""skill-tester 通用工具：解析 testcase_*.md、定位技能目录、校验报告 JSON。"""
from __future__ import annotations

import json
import re
import subprocess
from pathlib import Path
from typing import Any

SKILL_TESTER_ROOT = Path(__file__).resolve().parent.parent
REPO_ROOT_CANDIDATES = 8
TESTCASE_PREFIX = "testcase_"

ROW_KEYS = ("id", "title", "precondition", "expected", "actual", "result", "basis")
META_KEYS = (
    "skill_name",
    "hero",
    "faction",
    "category",
    "commit",
    "impl_path",
    "test_method",
    "diff_notes",
)


def testcase_filename(skill_name: str) -> str:
    """用例文件名：testcase_{技能名}.md"""
    return f"{TESTCASE_PREFIX}{skill_name.strip()}.md"


def skill_name_from_testcase_path(path: Path) -> str:
    stem = path.stem
    if stem.startswith(TESTCASE_PREFIX):
        return stem[len(TESTCASE_PREFIX) :]
    return stem


def resolve_testcase_md(skill_dir: Path) -> Path:
    """定位技能目录下的用例文件 testcase_{技能名}.md。"""
    skill_dir = skill_dir.resolve()
    if not skill_dir.is_dir():
        raise FileNotFoundError(f"技能目录不存在: {skill_dir}")

    by_folder = skill_dir / testcase_filename(skill_dir.name)
    if by_folder.is_file():
        return by_folder

    matches = sorted(skill_dir.glob(f"{TESTCASE_PREFIX}*.md"))
    if len(matches) == 1:
        return matches[0]
    if len(matches) > 1:
        for m in matches:
            if m.name == testcase_filename(skill_dir.name):
                return m
        names = ", ".join(m.name for m in matches)
        raise FileNotFoundError(f"{skill_dir} 下存在多个用例文件，请保留唯一 testcase_*.md: {names}")

    raise FileNotFoundError(
        f"未找到用例文件 {testcase_filename(skill_dir.name)}（目录: {skill_dir}）"
    )


def has_testcase_md(skill_dir: Path) -> bool:
    try:
        resolve_testcase_md(skill_dir)
        return True
    except FileNotFoundError:
        return False


def find_skill_dir(skill_name: str, root: Path | None = None) -> Path:
    """按文件夹名或 testcase 内「技能名称」定位技能目录。"""
    base = (root or SKILL_TESTER_ROOT).resolve()
    if not base.is_dir():
        raise FileNotFoundError(f"skill-tester 根目录不存在: {base}")

    name = skill_name.strip()

    # 1) 文件夹名完全匹配且存在对应用例文件
    direct = base / name
    if direct.is_dir() and has_testcase_md(direct):
        return direct

    # 2) 遍历：文件夹名或概览技能名称
    for child in sorted(base.iterdir()):
        if not child.is_dir() or child.name == "scripts":
            continue
        if not has_testcase_md(child):
            continue
        if child.name == name:
            return child
        overview = parse_test_md_overview(resolve_testcase_md(child))
        if overview.get("skill_name") == name:
            return child

    raise FileNotFoundError(
        f"未找到技能目录: {name!r}（查找文件夹名或 {testcase_filename(name)} 内技能名称）"
    )


def list_skill_dirs(root: Path | None = None) -> list[Path]:
    base = (root or SKILL_TESTER_ROOT).resolve()
    out: list[Path] = []
    for child in sorted(base.iterdir()):
        if child.is_dir() and child.name != "scripts" and has_testcase_md(child):
            out.append(child)
    return out


def parse_test_md_overview(test_md: Path) -> dict[str, str]:
    text = test_md.read_text(encoding="utf-8")
    out = {k: "" for k in ("skill_name", "hero", "faction", "category")}
    patterns = {
        "skill_name": r"\*\*技能名称\*\*[：:]\s*(.+)",
        "hero": r"\*\*所属武将\*\*[：:]\s*(.+)",
        "faction": r"\*\*武将势力\*\*[：:]\s*(.+)",
        "category": r"\*\*技能分类\*\*[：:]\s*(.+)",
    }
    for key, pat in patterns.items():
        m = re.search(pat, text)
        if m:
            out[key] = m.group(1).strip()
    return out


def parse_test_md_cases(test_md: Path) -> list[dict[str, str]]:
    """解析测试用例表。列顺序：编号|一级|二级|拆解|标题|前置|步骤|预期结果"""
    rows: list[dict[str, str]] = []
    in_table = False
    for line in test_md.read_text(encoding="utf-8").splitlines():
        if line.startswith("| TC-"):
            in_table = True
        if not in_table or not line.startswith("| TC-"):
            continue
        parts = [p.strip() for p in line.split("|")]
        if len(parts) < 10:
            continue
        rows.append(
            {
                "id": parts[1],
                "module_l1": parts[2],
                "module_l2": parts[3],
                "breakdown": parts[4],
                "title": parts[5],
                "precondition": parts[6].replace("<br>", "\n"),
                "steps": parts[7].replace("<br>", "\n"),
                "expected": parts[8].replace("<br>", "\n"),
            }
        )
    return rows


def git_commit(start: Path | None = None) -> str:
    repo = (start or SKILL_TESTER_ROOT).resolve()
    for _ in range(REPO_ROOT_CANDIDATES):
        if (repo / ".git").exists():
            try:
                return subprocess.check_output(
                    ["git", "log", "-1", "--oneline"],
                    cwd=repo,
                    text=True,
                    encoding="utf-8",
                    stderr=subprocess.DEVNULL,
                ).strip()
            except (subprocess.CalledProcessError, FileNotFoundError):
                return "unknown"
        if repo.parent == repo:
            break
        repo = repo.parent
    return "unknown"


def normalize_result(raw: str) -> str:
    s = (raw or "").strip()
    if s in ("✅", "通过", "PASS", "pass"):
        return "通过"
    if s in ("❌", "失败", "FAIL", "fail"):
        return "失败"
    if "待验证" in s or s in ("PENDING", "pending", "?"):
        return "待验证"
    return s or "待验证"


def row_from_obj(obj: dict[str, Any]) -> tuple[str, ...]:
    return (
        str(obj.get("id") or obj.get("case_id") or ""),
        str(obj.get("title") or ""),
        str(obj.get("precondition") or obj.get("前置条件") or ""),
        str(obj.get("expected") or obj.get("预期") or ""),
        str(obj.get("actual") or obj.get("实际") or ""),
        normalize_result(str(obj.get("result") or obj.get("结果") or "")),
        str(obj.get("basis") or obj.get("规则依据") or obj.get("规则依据（代码路径）") or ""),
    )


def validate_report_data(data: dict[str, Any]) -> None:
    rows = data.get("rows")
    if not isinstance(rows, list) or not rows:
        raise ValueError("report_data 缺少非空 rows 数组")
    for i, row in enumerate(rows, 1):
        if not isinstance(row, dict):
            raise ValueError(f"第 {i} 条用例应为对象")
        missing = [k for k in ROW_KEYS if k not in row or row[k] is None]
        if missing:
            raise ValueError(f"第 {i} 条用例缺少字段: {', '.join(missing)}")
        if normalize_result(str(row["result"])) not in ("通过", "失败", "待验证"):
            raise ValueError(f"第 {i} 条 result 无效: {row['result']!r}")


def merge_report_meta(skill_dir: Path, data: dict[str, Any]) -> dict[str, Any]:
    testcase_md = resolve_testcase_md(skill_dir)
    overview = parse_test_md_overview(testcase_md)
    meta = {**overview}
    for k in META_KEYS:
        if data.get(k):
            meta[k] = data[k]
    if not meta.get("skill_name"):
        meta["skill_name"] = skill_name_from_testcase_path(testcase_md) or skill_dir.name
    if not meta.get("commit"):
        meta["commit"] = git_commit(skill_dir)
    if not meta.get("test_method"):
        meta["test_method"] = "对照服务端代码推演（skill-tester）"
    if data.get("diff_notes"):
        meta["diff_notes"] = data["diff_notes"]
    if data.get("impl_path"):
        meta["impl_path"] = data["impl_path"]
    return meta


def load_report_json(path: Path | None, stdin_data: dict | None = None) -> dict[str, Any]:
    if stdin_data is not None:
        data = stdin_data
    else:
        data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError("报告 JSON 根节点必须是对象")
    validate_report_data(data)
    return data


def write_report_json(skill_dir: Path, data: dict[str, Any], filename: str = "report_data.json") -> Path:
    out = skill_dir / filename
    out.write_text(json.dumps(data, ensure_ascii=False, indent=2), encoding="utf-8")
    return out
