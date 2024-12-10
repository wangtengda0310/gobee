package main

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type protoMessage struct {
	Proto string
	Msg   interface{}
}
type testOneCase struct {
	Req protoMessage
	Ack protoMessage
}
type Testcase struct {
	Name string
	Seq  []testOneCase
}

func (pm *protoMessage) UnmarshalYAML(n *yaml.Node) error {
	// 创建一个匿名结构体用于通用解析
	type rawProtoMessage struct {
		Proto string      `yaml:"proto"`
		Msg   interface{} `yaml:"msg"`
	}

	// 使用通用结构体进行初步解析
	var raw rawProtoMessage
	if err := n.Decode(&raw); err != nil {
		return err
	}

	// 复制解析结果
	pm.Proto = raw.Proto

	// 根据Proto值进行特殊处理
	switch raw.Proto {
	case "sendReq":
		// 假设对于sendReq，我们需要将Msg解析为字符串
		pm.Msg = fmt.Sprintf("Processed: %v", raw.Msg)
	case "receiveAck":
		// 假设对于receiveAck，我们需要将Msg解析为map
		var msgMap map[string]string
		if err := yaml.Unmarshal([]byte(fmt.Sprintf("%v", raw.Msg)), &msgMap); err != nil {
			return err
		}
		pm.Msg = msgMap
	default:
		// 默认情况下，直接使用原始Msg
		pm.Msg = raw.Msg
	}

	return nil
}

func marshal(test *Testcase) string {
	out, err := yaml.Marshal(test)
	if err != nil {
		panic(err)
	}

	return string(out)
}
func unmarshal(data []byte) Testcase {
	var testcase Testcase

	err := yaml.Unmarshal(data, &testcase)
	if err != nil {
		fmt.Println("error:", err)
		return Testcase{}
	}

	return testcase
}
