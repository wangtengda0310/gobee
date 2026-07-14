package feishu

import (
	"strings"

	"github.com/wangtengda0310/gobee/office-toolbox/robot/im/feishu"
)

// DecodeOctalEscape 解码字符串中的八进制转义序列
// 将 \ddd 格式的八进制转义序列转换为对应的 ASCII 字符
//
// 参数：
//   - s: 包含八进制转义序列的原始字符串
//
// 返回：
//   - 解码后的字符串，所有 \ddd 序列已转换为对应字符
//
// 示例：
//
//	DecodeOctalEscape("\\243\\274") → "£¼" (八进制 243=163='£', 274=188='¼')
//
// 使用场景：飞书消息中的中文等非 ASCII 字符在传输时会被编码为八进制转义序列
func DecodeOctalEscape(s string) string {
	var result strings.Builder
	i := 0

	for i < len(s) {
		if i+3 < len(s) && s[i] == '\\' {
			// 检查是否是八进制转义序列 \ddd
			b1, b2, b3 := s[i+1], s[i+2], s[i+3]
			if isOctal(b1) && isOctal(b2) && isOctal(b3) {
				// 计算八进制值: d1*64 + d2*8 + d3
				value := (b1-'0')*64 + (b2-'0')*8 + (b3 - '0')
				result.WriteByte(value)
				i += 4
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}

	return result.String()
}

// isOctal 判断字符是否为八进制数字 (0-7)
func isOctal(c byte) bool {
	return c >= '0' && c <= '7'
}

// WarningRed 发送红色样式卡片
func WarningRed(robot, title, subTitle, content string) {
	prefab := &feishu.JsonCardPrefab{}
	SendFeiShuRobotCardJson(robot, prefab.WarningRed(title, subTitle, &feishu.MDElement{
		Tag:       "markdown",
		Content:   content,
		TextAlign: "left",
		TextSize:  "normal_v2",
		Margin:    "0px 0px 0px 0px",
	}))
}

// SuccessGreen 发送绿色样式卡片
func SuccessGreen(robot, title, subTitle, content string) {
	prefab := &feishu.JsonCardPrefab{}
	SendFeiShuRobotCardJson(robot, prefab.SuccessGreen(title, subTitle, &feishu.MDElement{
		Tag:       "markdown",
		Content:   content,
		TextAlign: "left",
		TextSize:  "normal_v2",
		Margin:    "0px 0px 0px 0px",
	}))
}
