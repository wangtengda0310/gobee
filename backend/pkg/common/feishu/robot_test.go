package feishu

import (
	"testing"

	. "github.com/wangtengda0310/gobee/office-toolbox/robot/im/feishu"
)

// @别人
// <at email="v-tangfangda@ztgame.com"></at>
// <at user_id="v-tangfangda"></at>

// TestSendMsg 测试发送文本消息并 @指定用户
// 验证飞书机器人文本消息发送功能，支持 @用户 功能
func TestSendMsg(t *testing.T) {
	t.Skip("避免自动测试发送消息污染飞书")
	SendFeiShuRobotText(LJH, "<at user_id=\"%s\"></at>", "v-tangfangda")
}

// TestSendCardTemplateMsg 测试发送卡片模板消息（带饼图）
// 使用预定义模板发送包含图表的飞书卡片消息
// 跳过原因：避免自动化测试时发送大量消息污染飞书群
func TestSendCardTemplateMsg(t *testing.T) {
	t.Skip("避免自动测试发送消息污染飞书")
	SendFeiShuRobotCardTemplate(LJH, "AAqv2pptBvoUw", "1.0.1", map[string]interface{}{
		"title":   "1232352345",
		"content": "我是内容",
		"pie_data": map[string]interface{}{
			"type": "pie",
			"title": map[string]interface{}{
				"text": "图表",
			},
			"data": map[string]interface{}{
				"values": []map[string]interface{}{
					{"type": "A", "value": "111"},
					{"type": "B", "value": "222"},
					{"type": "C", "value": "333"},
					{"type": "D", "value": "444"},
					{"type": "E", "value": "555"},
				},
			},
			"valueField":    "value",
			"categoryField": "type",
			"outerRadius":   0.9,
			"legends": map[string]interface{}{
				"visible": true,
				"orient":  "right",
			},
			"padding": map[string]interface{}{
				"left":   10,
				"top":    10,
				"bottom": 5,
				"right":  0,
			},
			"label": map[string]interface{}{
				"visible": true,
			},
		},
	})
}

// TestSendCardJsonMsg 测试发送 JSON 格式的卡片消息
// 使用 JsonCardPrefab 构建 Markdown 内容的蓝色卡片，支持 @指定邮箱用户
func TestSendCardJsonMsg(t *testing.T) {
	t.Skip("避免自动测试发送消息污染飞书")

	myOu := "v-tangfangda@ztgame.com"

	prefab := &JsonCardPrefab{}
	jsonCard := prefab.NormalBlue("标题", "副标题", &MDElement{
		Tag:       "markdown",
		Content:   "<at email=\"" + myOu + "\"></at>",
		TextAlign: "left",
		TextSize:  "normal_v2",
		Margin:    "0px 0px 0px 0px",
	})

	SendFeiShuRobotCardJson(LJH, jsonCard)
}
