#!/bin/bash
# ============================================================
# 飞书通知公共脚本
# 供 Docker 容器内各 skill 调用，封装群消息、私聊、监控通知。
#
# 用法:
#   /app/notify.sh dm <收件人邮箱> <内容>           # 发送私聊
#   /app/notify.sh group <标题> <内容>              # 发送群卡片消息
#   /app/notify.sh monitor <类型> <目标> <摘要>     # 发送脚本层监控通知
#
# 依赖环境变量:
#   FEISHU_DM_APP_ID / FEISHU_DM_APP_SECRET  # 私聊/监控必需
#   FEISHU_ROBOT                             # 群消息必需（不能为 none）
#   MONITOR_EMAIL                            # 监控通知接收邮箱（默认 v-wangtengda@ztgame.com）
# ============================================================

set -uo pipefail

# 通知失败不阻塞调用方
notify_exit_ok() { exit 0; }

log_info() {
	echo "[$(date '+%Y-%m-%d %H:%M:%S')] [INFO] $*"
}

log_error() {
	echo "[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $*" >&2
}

# 检测传入值是否为示例占位符，防止 prompt/测试中的示例命令被误执行
# 返回 0 表示是占位符
is_placeholder_value() {
	local value="$1"
	case "$value" in
		"标题"|"markdown 内容"|"收件人邮箱"|"消息内容"|"类型"|"目标"|"摘要"|\
		"<标题>"|"<markdown 内容>"|"<收件人邮箱>"|"<消息内容>"|"<类型>"|"<目标>"|"<摘要>")
			return 0
			;;
	esac
	return 1
}

# 获取飞书 tenant_access_token（stdin 传参避免命令行泄露凭证）
# 成功输出 token，失败输出空
get_feishu_token() {
	if [ -z "${FEISHU_DM_APP_ID:-}" ] || [ -z "${FEISHU_DM_APP_SECRET:-}" ]; then
		return 1
	fi

	local token_response=""
	token_response=$(printf '{"app_id":"%s","app_secret":"%s"}' \
		"$FEISHU_DM_APP_ID" "$FEISHU_DM_APP_SECRET" | \
		curl -s -m 10 -X POST \
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal" \
		-H "Content-Type: application/json" \
		-d @- 2>/dev/null) || true

	echo "$token_response" | jq -r '.tenant_access_token // empty' || true
}

# 发送飞书私聊消息
# 参数: $1=收件人邮箱 $2=消息内容
send_dm() {
	local receive_email="$1"
	local content="$2"

	if is_placeholder_value "$receive_email" || is_placeholder_value "$content"; then
		log_info "[DEBUG-DM] 飞书私聊已跳过 (参数为示例占位符)"
		notify_exit_ok
	fi

	log_info "[DEBUG-DM] send_feishu_dm 被调用 | 收件人: $receive_email | 内容长度: ${#content}"

	if [ -z "${FEISHU_DM_APP_ID:-}" ] || [ -z "${FEISHU_DM_APP_SECRET:-}" ]; then
		log_info "[DEBUG-DM] 飞书私聊已跳过 (未配置 FEISHU_DM_APP_ID/SECRET)"
		notify_exit_ok
	fi

	log_info "[DEBUG-DM] 正在获取 tenant_access_token..."
	local token=""
	token=$(get_feishu_token)
	if [ -z "$token" ]; then
		log_error "[DEBUG-DM] 获取飞书 token 失败"
		notify_exit_ok
	fi
	log_info "[DEBUG-DM] tenant_access_token 获取成功"

	# 截断内容（飞书文本限制 150KB，保守取 100KB）
	local truncated_content="$content"
	if [ "${#content}" -gt 100000 ]; then
		truncated_content="${content:0:99000}

...(内容已截断)"
		log_info "[DEBUG-DM] 内容已截断 (原长度: ${#content})"
	fi

	log_info "[DEBUG-DM] 正在发送私聊消息 → $receive_email"
	local inner_content
	inner_content=$(printf '%s' "$truncated_content" | jq -Rs .)
	local request_body
	request_body=$(jq -n \
		--arg rid "$receive_email" \
		--arg content "{\"text\":$inner_content}" \
		'{"receive_id": $rid, "msg_type": "text", "content": $content}')

	local send_response=""
	send_response=$(printf '%s' "$request_body" | \
		curl -s -m 30 -X POST \
		"https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=email" \
		-H "Authorization: Bearer $token" \
		-H "Content-Type: application/json" \
		-d @- 2>/dev/null) || true

	local code=""
	code=$(echo "$send_response" | jq -r '.code // -1') || true
	if [ "$code" = "0" ]; then
		log_info "[DEBUG-DM] 飞书私聊发送成功 → $receive_email"
	else
		log_error "[DEBUG-DM] 飞书私聊发送失败 | code: $code | 响应: $send_response"
	fi

	notify_exit_ok
}

# 发送脚本层监控私聊（与 backend DMMonitorDecorator 区分）
# 参数: $1=通知类型 $2=目标 $3=摘要
send_monitor() {
	local notify_type="$1"
	local target="$2"
	local summary="$3"
	local monitor_email="${MONITOR_EMAIL:-v-wangtengda@ztgame.com}"

	if is_placeholder_value "$target" || is_placeholder_value "$summary"; then
		log_info "[DEBUG-MONITOR] 监控通知已跳过 (参数为示例占位符)"
		notify_exit_ok
	fi

	if [ -z "$monitor_email" ]; then
		notify_exit_ok
	fi

	log_info "[DEBUG-MONITOR] 发送监控通知 → $monitor_email | 类型: $notify_type"

	if [ -z "${FEISHU_DM_APP_ID:-}" ] || [ -z "${FEISHU_DM_APP_SECRET:-}" ]; then
		log_info "[DEBUG-MONITOR] 监控通知已跳过 (未配置 FEISHU_DM_APP_ID/SECRET)"
		notify_exit_ok
	fi

	local token=""
	token=$(get_feishu_token)
	if [ -z "$token" ]; then
		log_error "[DEBUG-MONITOR] 获取飞书 token 失败"
		notify_exit_ok
	fi

	local branch="${CI_COMMIT_BRANCH:-${COMMIT_BRANCH:-?}}"
	local hash="${CI_COMMIT_HASH:-${COMMIT_HASH:-?}}"
	local monitor_content
	monitor_content=$(printf '[DEBUG-MONITOR] [entrypoint] 通知已分发\n类型: %s\n目标: %s\n分支: %s\n提交: %s\n消息摘要: %s' \
		"$notify_type" "$target" "$branch" "$hash" "$summary")

	local inner_content
	inner_content=$(printf '%s' "$monitor_content" | jq -Rs .)
	local request_body
	request_body=$(jq -n \
		--arg rid "$monitor_email" \
		--arg content "{\"text\":$inner_content}" \
		'{"receive_id": $rid, "msg_type": "text", "content": $content}')

	local send_response=""
	send_response=$(printf '%s' "$request_body" | \
		curl -s -m 30 -X POST \
		"https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=email" \
		-H "Authorization: Bearer $token" \
		-H "Content-Type: application/json" \
		-d @- 2>/dev/null) || true

	local code=""
	code=$(echo "$send_response" | jq -r '.code // -1') || true
	if [ "$code" = "0" ]; then
		log_info "[DEBUG-MONITOR] 监控通知发送成功 → $monitor_email"
	else
		log_error "[DEBUG-MONITOR] 监控通知发送失败 | code: $code | 响应: $send_response"
	fi

	notify_exit_ok
}

# 发送飞书群卡片通知
# 参数: $1=标题 $2=内容(markdown)
send_group() {
	local title="$1"
	local content="$2"

	if is_placeholder_value "$title" || is_placeholder_value "$content"; then
		log_info "[DEBUG-群消息] 飞书通知已跳过 (参数为示例占位符)"
		notify_exit_ok
	fi

	log_info "[DEBUG-群消息] send_feishu 被调用 | 标题: $title | 内容长度: ${#content}"

	if [ "${FEISHU_ROBOT:-none}" = "none" ] || [ -z "${FEISHU_ROBOT:-}" ]; then
		log_info "[DEBUG-群消息] 飞书通知已跳过 (FEISHU_ROBOT=none)"
		notify_exit_ok
	fi

	local webhook_url="https://open.feishu.cn/open-apis/bot/v2/hook/${FEISHU_ROBOT}"

	# 将 \n 转换为真实换行符，避免飞书显示字面量 \n
	local content_real
	content_real=$(printf '%b' "$content")
	local card_json
	card_json=$(jq -n \
		--arg title "$title" \
		--arg content "$content_real" \
		'{
			"msg_type": "interactive",
			"card": {
				"schema": "2.0",
				"config": {"wide_screen_mode": true},
				"header": {
					"title": {"tag": "plain_text", "content": $title},
					"template": "blue"
				},
				"body": {
					"elements": [
						{"tag": "markdown", "content": $content}
					]
				}
			}
		}')

	local response
	response=$(curl -s -m 30 -X POST "$webhook_url" \
		-H "Content-Type: application/json" \
		-d "$card_json")

	if echo "$response" | jq -e '.code == 0' > /dev/null 2>&1; then
		log_info "[DEBUG-群消息] 飞书通知发送成功"
	else
		log_error "[DEBUG-群消息] 飞书通知发送失败: $response"
	fi

	notify_exit_ok
}

# ==================== 入口 ====================

if [ $# -lt 1 ]; then
	echo "用法: $0 {dm|group|monitor} ..." >&2
	exit 1
fi

COMMAND="$1"
shift

case "$COMMAND" in
	dm)
		if [ $# -lt 2 ]; then
			echo "用法: $0 dm <收件人邮箱> <内容>" >&2
			exit 1
		fi
		send_dm "$1" "$2"
		;;
	group)
		if [ $# -lt 2 ]; then
			echo "用法: $0 group <标题> <内容>" >&2
			exit 1
		fi
		send_group "$1" "$2"
		;;
	monitor)
		if [ $# -lt 3 ]; then
			echo "用法: $0 monitor <类型> <目标> <摘要>" >&2
			exit 1
		fi
		send_monitor "$1" "$2" "$3"
		;;
	*)
		echo "未知命令: $COMMAND" >&2
		echo "用法: $0 {dm|group|monitor} ..." >&2
		exit 1
		;;
esac
