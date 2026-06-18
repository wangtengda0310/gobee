#!/usr/bin/env python3
"""把一次技能/牌用例生成的产出统计写入飞书 Bitable「测试用例生成数据统计」。

供 SKILL.md 阶段 4.7 调用，是看板的上游写入方。
本表对技能/牌**不去重、全表累加**，每次生成追加一条记录（运行ID 天然唯一）。

用法:
    append_stats.py <testcases.json> <对象名> <skill|card> <xlsx_url> <timestamp>

其中 xlsx_url 是上传后那份 xlsx 用例表的访问链接（落入「飞书链接」列，
点开即用例表本身，而非文件夹）。

其中 timestamp 为阶段零的紧凑时间戳（%Y%m%d_%H%M%S），脚本据此派生 ISO 生成时间
与运行ID（`<对象名>_<timestamp>`，与 SKILL_DIR 命名一致）。

OUTPUT:
    成功时 stdout 最后一行输出 `RECORD_ID=<record_id>`，供调用方捕获后用
    user-feedback skill 发反馈卡片（record_id 埋进卡片按钮 value）；用户点击后
    由常驻服务回写到本条 Bitable 记录的「用户反馈」字段。

写表是非关键的统计收尾：任何失败都打印警告并以 exit 0 退出，绝不中断主交付。
"""

import json
import subprocess
import sys
from datetime import datetime


LARK_CLI = "lark-cli"
APP_TOKEN = "AyOlbNrnzavuCzscvYPchXwtnUU"
TABLE_ID = "tblnlTnPgX12x9gP"

PROJECT = "名将杀"
TYPE_BY_OBJECT = {"skill": "技能", "card": "牌"}
SOURCE_BY_OBJECT = {"skill": "技能用例", "card": "牌用例"}


def _to_iso(timestamp: str) -> str:
    """紧凑时间戳 %Y%m%d_%H%M%S → ISO 8601；无法解析则退回当前时间。"""
    try:
        return datetime.strptime(timestamp, "%Y%m%d_%H%M%S").isoformat()
    except ValueError:
        return datetime.now().isoformat(timespec="seconds")


def _case_count(testcases_path: str) -> int:
    with open(testcases_path, "r", encoding="utf-8") as f:
        testcases = json.load(f)
    return len(testcases.get("rows", []))


def _create_record(fields: dict) -> str:
    """调用 Bitable record create，返回 record_id。"""
    path = f"/open-apis/bitable/v1/apps/{APP_TOKEN}/tables/{TABLE_ID}/records"
    cmd = [
        LARK_CLI, "api", "POST", path,
        "--data", json.dumps({"fields": fields}, ensure_ascii=False),
        "--as", "bot",
    ]
    proc = subprocess.run(cmd, capture_output=True,
                          text=True, encoding="utf-8", errors="replace")
    payload = json.loads(proc.stdout) if proc.stdout else {}
    if proc.returncode != 0 or payload.get("code") != 0:
        raise RuntimeError(proc.stdout or proc.stderr or "未知错误")
    return payload["data"]["record"]["record_id"]


def main() -> None:
    if len(sys.argv) != 6:
        print(f"用法: {sys.argv[0]} <testcases.json> <对象名> <skill|card> <xlsx_url> <timestamp>")
        sys.exit(2)

    testcases_path, object_name, object_type, xlsx_url, timestamp = sys.argv[1:6]

    try:
        count = _case_count(testcases_path)
        fields = {
            "项目": PROJECT,
            "类型": TYPE_BY_OBJECT.get(object_type, object_type),
            "名称": object_name,
            "来源": SOURCE_BY_OBJECT.get(object_type, ""),
            "用例数": count,
            "生成时间": _to_iso(timestamp),
            "飞书链接": xlsx_url,
            "运行ID": f"{object_name}_{timestamp}",
        }
        # 已覆盖 / 覆盖有误 / 部分覆盖 / 未覆盖 / 覆盖率 / 用户反馈 / 操作人 — 留空不传

        record_id = _create_record(fields)
        print(f"✅ 已写入统计表：{object_name}（{count} 条用例）")
        print(f"RECORD_ID={record_id}")
    except Exception as e:
        # 非关键收尾：失败只警告，不中断主交付（文件与链接已交付）。
        try:
            print(f"⚠️ 统计表写入失败，已跳过（不影响产物交付）：{e}")
        except Exception:
            pass
        sys.exit(0)


if __name__ == "__main__":
    main()
