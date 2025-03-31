package main

import (
	"github.com/spf13/pflag"
	"github.com/wangtengda/gobee/lvan/exporter/internal/mp"
)

func main() {
	defer mp.Recover()
	jsondir := pflag.String("jsondir", "", "json和xml所在目录")
	csvdir := pflag.String("csvdir", "", "csv所在目录")
	pflag.Parse()
	if jsondir != nil && *jsondir != "" {
		mp.Mainjson(*jsondir, pflag.Args()[0])
	}
	if csvdir != nil && *csvdir != "" {
		mp.Maincsv(*csvdir, pflag.Args()[0])
	}
}
