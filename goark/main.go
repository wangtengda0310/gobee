package main

import (
	"github.com/go-spring/spring-core/gs"
	"github.com/wangtengda0310/gobee/ark/csv"
	_ "github.com/wangtengda0310/gobee/ark/demo/httpsvr"
	_ "github.com/wangtengda0310/gobee/ark/demo/log"
	_ "github.com/wangtengda0310/gobee/ark/prometheus"
)

var d = BookDao{}

type BookDao struct {
	csv.Loader `autowire:""`
}

func (b BookDao) Run() error {
	b.Load()
	return nil
}

func main() {
	gs.EnableSimpleHttpServer(true)

	gs.Run(func() error {
		println(d.Loader)
		return nil
	})
}

func init() {
	gs.Object(&d).AsRunner()
	gs.SetActiveProfiles("online")
	gs.Property("csvfile", "testdata/")
}
