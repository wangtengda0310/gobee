package main

import (
	"github.com/wangtengda0310/gobee/lvan/internal/mp"
	"os"
)

func main() {
	defer mp.Recover()
	mp.Maincsv(os.Args[1], os.Args[2])
}
