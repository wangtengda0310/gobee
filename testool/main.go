package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var Mock = flag.Bool("Mock", true, "isMock")
var Test = flag.Bool("Test", true, "isTest")

func main() {
	flag.Parse()
	fmt.Println(*Mock, *Test)
	if *Mock {
		for _, f := range flag.Args() {
			gen(f, nil)
		}

		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		gen("", string(data))
	}
	if *Test {
		generateTestCasesForFuncOrMethod(nil)
	}
}
