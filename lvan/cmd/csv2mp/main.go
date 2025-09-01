package main

import (
	"os"

	"github.com/wangtengda0310/gobee/lvan/internal/mp"
)

func main() {
	defer mp.Recover()
	mp.Maincsv(os.Args[1], os.Args[2])
}
