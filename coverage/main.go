package main

import (
	"flag"
	"fmt"
	"os"
)

var token = flag.String("token", "c1be2fb9af823da8a1d5d5b7a616a6cad3646b39b86ce0c170fbd387936689a7", "机器人token")
var secret = flag.String("secret", "SECbdc49e29fba225a5c4ba50e4786e81664e5058d88765c8fba3ed54779a04d6b1", "机器人secret")

var dingding = flag.Bool("dingding", false, "是否发送钉钉消息")

var title = flag.String("title", "新提交代码测试覆盖率统计", "消息标题")

var alarmUrl = flag.String("alarmUrl", "http://alarm.iwgame.com/alarm/dingtalk/sendTemplate", "报警url")

func main() {
	if "sumPackage" == os.Args[1] {
		lines, parentDirs := scanStdin(os.Stdin)
		for _, line := range lines {
			fmt.Println("-	", line)
		}
		for _, dir := range parentDirs {
			fmt.Println("total	", dir)
		}
		return
	}
	flag.Parse()
	var analyser = gotoolcover{}
	j := analyser.analyseCoverage()
	alarmJson(j)

}

type gotoolcover struct{}

func (gotoolcover) analyseCoverage() *JSONData {

	return analyseGoToolsCover()
}
