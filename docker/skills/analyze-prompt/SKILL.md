---
name: analyze-prompt
description: |
  分析策划配表提交的影响范围，产出可被飞书渲染的结构化分析报告，并按分支策略发送通知。

  **必须使用此技能的场景**：
  - 流水线通过 `CI_INVOKE_SKILL=analyze-prompt` 触发本次任务
  - 用户要求分析某次配表提交（Excel 变更）的影响范围
  - 用户要求执行 rain-excel-checker 校验并解读结果

  **不要触发**：
  - 仅需解析单个 Excel 文件内容（用 /excel-parser）
  - 仅需查询配表关系（用 /game-config-relations）
  - 仅需检查 commit message（用 /check-config-commit-message）

  **判定关键**：需要对一次配表提交产出「校验结果 + commit message 检查 + 影响分析」完整报告，并发送飞书通知时使用。
version: 2.0.0
tags: [excel, config-analysis, ci-pipeline, feishu-report]
---

# 配表提交影响分析

你是一名游戏策划配表分析专家。请分析本次配表提交的影响范围并按分支策略发送飞书通知。

## 1. 读取任务上下文

所有原始上下文由 `entrypoint.sh` 通过环境变量注入。**不要依赖 entrypoint 做业务计算**，所有分支判断、diff 过滤、通知策略都在本 skill 中完成。

```bash
# 基础路径
CONFIG_REPO="${CI_CONFIG_REPO:-/config-repo}"
CASE_DIR="${CI_CASE_DIR:-/workspace/cases}"
TARGET_COMMIT="${CI_TARGET_COMMIT:-}"

# CI 注入的原始分支/提交（可能为空）
BK_BRANCH="${BK_CI_HOOK_BRANCH:-}"
BK_REVISION="${BK_CI_HOOK_REVISION:-}"

# 通知凭证
FEISHU_ROBOT="${FEISHU_ROBOT:-none}"
FEISHU_DM_APP_ID="${FEISHU_DM_APP_ID:-}"
FEISHU_DM_APP_SECRET="${FEISHU_DM_APP_SECRET:-}"
MONITOR_EMAIL="${CI_CLAUDE_MONITOR_EMAIL:-v-wangtengda@ztgame.com}"

# 检查模式：full=强制全量检查（定时任务用），空=按分支/merge 判断
CHECK_MODE="${CI_CLAUDE_CHECK_MODE:-}"
```

## 2. 确定分析目标与分支信息

切换到 `$CONFIG_REPO`，使用 git 获取提交信息：

```bash
cd "$CONFIG_REPO"

if [ -n "$TARGET_COMMIT" ]; then
    COMMIT_HASH=$(git rev-parse "$TARGET_COMMIT")
    COMMIT_AUTHOR=$(git log -1 --format='%an' "$COMMIT_HASH")
    COMMIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    COMMIT_MESSAGE=$(git log -1 --format='%s' "$COMMIT_HASH")
else
    COMMIT_HASH=$(git rev-parse HEAD)
    COMMIT_AUTHOR=$(git log -1 --format='%an' HEAD)
    COMMIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    COMMIT_MESSAGE=$(git log -1 --format='%s' HEAD)
fi

# CI detached HEAD 场景：用流水线注入的分支名覆盖
if [ -n "$BK_BRANCH" ]; then
    COMMIT_BRANCH="${BK_BRANCH#refs/heads/}"
fi

COMMIT_AUTHOR_EMAIL=$(git log -1 --format='%ae' HEAD)
```

分支与 merge 判断：

```bash
IS_PRE_RELEASE="false"
if [ "$COMMIT_BRANCH" = "v0.0.8-pre-release" ]; then
    IS_PRE_RELEASE="true"
fi

IS_MERGE="false"
MERGE_PARENTS=$(git log -1 --format='%P' HEAD | wc -w)
if [ "$MERGE_PARENTS" -ge 2 ]; then
    IS_MERGE="true"
fi
```

## 3. 构建 rain-excel-checker 命令

所有分支都传 `-noDM -feishuRobot=none`（通知由本 skill 统一发送）。`-mode=full`（全量检查）触发条件，优先级从高到低：

1. `CI_CLAUDE_CHECK_MODE=full` → 定时任务或强制全量（优先级最高）
2. 非预发布分支的 merge 提交 → 全量检查
3. 其他 → 增量检查

```bash
BASE_CMD="/usr/local/bin/rain-excel-checker -feishuRobot=none -noDM -excelPath=$CONFIG_REPO -casePath=$CASE_DIR"
if [ "$CHECK_MODE" = "full" ]; then
    EXCEL_CHECKER_CMD="$BASE_CMD -mode=full"
elif [ "$IS_MERGE" = "true" ] && [ "$IS_PRE_RELEASE" = "false" ]; then
    EXCEL_CHECKER_CMD="$BASE_CMD -mode=full"
else
    EXCEL_CHECKER_CMD="$BASE_CMD"
fi
```

## 4. 获取变更 Excel 文件列表

```bash
if [ -n "$TARGET_COMMIT" ]; then
    DIFF_FILES=$(git -C "$CONFIG_REPO" diff --name-only "${COMMIT_HASH}~1" "${COMMIT_HASH}" 2>/dev/null || true)
else
    if git -C "$CONFIG_REPO" rev-parse HEAD~1 > /dev/null 2>&1; then
        DIFF_FILES=$(git -C "$CONFIG_REPO" diff --name-only HEAD~1 HEAD 2>/dev/null || true)
    else
        DIFF_FILES=$(git -C "$CONFIG_REPO" ls-files "*.xlsx" 2>/dev/null || true)
    fi
fi

EXCEL_FILES=$(echo "$DIFF_FILES" | grep -E '\.xlsx$' || true)
FILE_COUNT=$(echo "$EXCEL_FILES" | grep -c '.' || true)
```

如果无 Excel 变更，直接发送通知后退出：

```bash
if [ -z "$EXCEL_FILES" ]; then
    NO_CHANGE_MSG="**提交者**: ${COMMIT_AUTHOR}\n**分支**: ${COMMIT_BRANCH}\n**Commit**: ${COMMIT_HASH:0:8}\n\n本次提交无 Excel 文件变更，跳过分析。"
    if [ "$IS_PRE_RELEASE" = "true" ]; then
        /app/notify.sh group "配表检查 - 无变更" "$NO_CHANGE_MSG"
    else
        /app/notify.sh dm "$COMMIT_AUTHOR_EMAIL" "$NO_CHANGE_MSG"
    fi
    exit 0
fi
```

## 5. 调用其他 skill 辅助分析

如有需要，可调用：

- /check-config-commit-message — 检查 commit message 是否与修改内容匹配
- /excel-parser — 解析 Excel 文件内容
- /game-config-relations — 查询配表之间的关系和业务规则
  - 如遇知识盲区，该 skill 会自动通过 /multica-cli 创建 issue
- /multica-cli — 在流水线环境中创建 issue

## 6. 生成分析报告

### 6.1 rain-excel-checker 校验结果

执行 `bash -c "$EXCEL_CHECKER_CMD"`，原文输出程序校验结果，仅保留最终检查统计和表级变更通知，排除过程日志。

### 6.2 /check-config-commit-message 检查结果

输出匹配度评级、问题列表、改进建议。

### 6.3 影响分析

按以下固定结构输出：

- **影响的系统**：战斗/经济/活动/UI 等
- **跨表影响**：可能受影响的其他 sheet
- **风险评估**：高/中/低
- **建议**：需额外验证或关注的事项

完整飞书消息示例见本 skill 末尾附录。

## 7. 发送飞书通知

### 7.1 预发布分支（v0.0.8-pre-release）→ 群消息

如果是 merge 场景，在消息开头追加 @ 提交者列表：

```bash
AT_LIST=""
if [ "$IS_PRE_RELEASE" = "true" ] && [ "$IS_MERGE" = "true" ]; then
    MERGE_AUTHORS=$(git log --format='%ae' HEAD~1..HEAD 2>/dev/null | sort -u || true)
    for email in $MERGE_AUTHORS; do
        AT_LIST="${AT_LIST}<at email=\"${email}\"></at> "
    done
    [ -n "$AT_LIST" ] && AT_LIST="${AT_LIST}\n\n"
fi

FEISHU_TITLE="配表影响分析 - ${COMMIT_BRANCH}"
FEISHU_CONTENT="${AT_LIST}**提交者**: ${COMMIT_AUTHOR}  **分支**: ${COMMIT_BRANCH}  **Commit**: \`${COMMIT_HASH:0:8}\`\n**消息**: ${COMMIT_MESSAGE}\n**变更文件**: ${FILE_COUNT} 个 Excel\n\n---\n\n${CLAUDE_TEXT}"

# 截断到 15000 字符
if [ "${#FEISHU_CONTENT}" -gt 15000 ]; then
    FEISHU_CONTENT="${FEISHU_CONTENT:0:14900}\n\n...(内容已截断，超出 15000 字符限制)"
fi

/app/notify.sh group "$FEISHU_TITLE" "$FEISHU_CONTENT"
```

### 7.2 其他分支 → 私聊

```bash
/app/notify.sh dm "$COMMIT_AUTHOR_EMAIL" "$CLAUDE_TEXT"
```

## 8. 发送监控通知

无论群消息还是私聊，成功后都发送脚本层监控通知：

```bash
/app/notify.sh monitor "私聊" "$COMMIT_AUTHOR_EMAIL" "内容长度 ${#CLAUDE_TEXT}"
# 或
/app/notify.sh monitor "群消息" "$FEISHU_TITLE" "内容长度 ${#FEISHU_CONTENT}"
```

## 9. 内容约束

1. 言简意赅
2. 格式可以正确被飞书渲染
3. 发现的问题列表和改进建议如果没有就不显示，不要显示"无"
4. 影响分析中的建议如果没有就不显示，不要显示"无"

## 附录：飞书消息示例

```markdown
# 配表提交影响分析报告

## 配表校验结果

📊 检查统计:
• 列级检查: 0 项 (失败: 0)
• 表级检查: 5 项 (失败: 0)
• 执行的表级规则:

- ADDED_COL_NOTIFY
  • 解析错误: 0 项

📝 表级变更通知:
• 📋 表格名称: 无尽buff效果表|RougeEndlessNodeBuff
🔄 变更类型: 修改行
变更范围: 共 1 行记录发生变更
【变更记录 1】
第 84 行，Id 80
✏️ BeforeDelete:
[原值] 74
[新值] <span style="color: red; ">41,</span>74

⏰ 变更时间: 2026-05-22 10:51:42
👤 提交人: <at email="houzhensong@ztgame.com"></at>
🔗 对比版本: 38a8804507a7596302ae40a5cfc9bd9bd86d244a → 3fc93fc7f5074d878bde7c9854d86ef2c8fc2850

---

## 配表提交 Commit Message 检查报告

### 匹配度

✅ 完全匹配 / ⚠️ 部分匹配 / ❌ 不匹配 / 📝 过于笼统

### ⚠️发现问题

问题1 . 描述提到"调整千里单骑数值"但未说明具体字段变化
问题2. 新增了 buff 配置但描述未提及

### 原始提交日志

> **Commit**: 张飞技能配置
> **作者**: 张三
> **日期**: 2026-05-21 10:00

### 改进建议

建议的 commit message:
> 调整千里单骑无尽模式数值：
> - 动作表：降低精英怪触发概率 20%→15%
> - 计算器表：提高层数奖励系数 1.2→1.5
> - buff表：新增暴击buff效果

---

## 影响报告

**影响的系统**: *战斗系统*修改了 buff 的生命周期末尾效果，影响玩家在特定节点获得的 buff 到期后的效果<br/> *无尽模式*千里单骑无尽模式中战斗节点 80（单刀赴会）相关的 buff 效果
**跨表影响**: *道具表\|Item*道具到期时间不匹配 <br/> *皮肤抽奖\|DrawSkin*副产物道具ID配置的id丢失
**风险评估**: **中风险**  - 数值调整影响范围明确（仅影响节点 80 的 buff）<br>- 变更为增加 debuff 效果（初始手牌-2，出杀次数-2），可能降低玩家体验<br>- 需要验证 buff 到期逻辑是否正确处理多个效果
**建议**: *验证重点*：测试战斗节点 80（单刀赴会）的 buff 到期效果是否正确执行两个效果（41 和 74） <br/> *体验评估*：增加"初始手牌-2，出杀次数-2"的 debuff 可能影响玩家在该节点的体验，需要评估是否符合设计意图 <br/> *回归测试*：检查其他使用 buff 效果 41 或 74 的节点是否受到影响

```
