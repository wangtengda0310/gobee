package feishu

import (
	"fmt"

	. "github.com/wangtengda0310/gobee/office-toolbox/robot/im/feishu"
)

// LJH 垃圾话机器人id
// var LJH = "36732a0b-9b65-4456-8294-17044223114f"
// https://open.feishu.cn/open-apis/bot/v2/hook/db06f82a-4dad-43f6-bbef-97503e0b953a
var LJH = "db06f82a-4dad-43f6-bbef-97503e0b953a"

func SendFeiShuRobotText(robot, format string, v ...interface{}) {
	Send(robot, &MsgData{
		MsgType: "text",
		Content: &Content{Text: fmt.Sprintf(format, v...)},
	})
}

func SendFeiShuRobotCardTemplate(robot, template, version string, v map[string]interface{}) {
	Send(robot, &MsgData{
		MsgType: "interactive",
		Card: &TemplateCard{
			Type: "template",
			Data: CardData{
				TemplateId:          template,
				TemplateVersionName: version,
				TemplateVariable:    v,
			},
		},
	})
}

func SendFeiShuRobotCardJson(robot string, jsonCard *JsonCard) {
	Send(robot, &MsgData{
		MsgType: "interactive",
		Card:    jsonCard,
	})
}
