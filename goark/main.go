package main

import (
	"github.com/go-spring/spring-core/gs"
	_ "github.com/wangtengda0310/gobee/ark/demo/httpsvr"
	_ "github.com/wangtengda0310/gobee/ark/demo/log"
	_ "github.com/wangtengda0310/gobee/ark/prometheus"
)

func main() {
	gs.EnableSimpleHttpServer(true)
	gs.Run()
}
func init() {
	gs.SetActiveProfiles("online")
}
