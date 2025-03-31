package main

import (
	"github.com/wangtengda/gobee/lvan/exporter/internal/mp"
	"os"
)

func main() {
	defer mp.Recover()
	mp.Mainjson(os.Args[1], os.Args[2])
}
