package main

import (
	"flag"
	"fmt"
	"github.com/wangtengda0310/gobee/lvan/cmd/coverage/internal"
	"os"
)

type StringList []string

func (l *StringList) String() string {
	return fmt.Sprint(*l)
}
func (l *StringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

type StringMap map[string]string

func (m *StringMap) String() string {
	return fmt.Sprint(*m)
}
func (m *StringMap) Set(value string) error {
	(*m)[value] = value
	return nil
}

func main() {
	if len(os.Args) > 1 && "sumPackage" == os.Args[1] {
		lines, parentDirs := internal.ScanStdin(os.Stdin)
		for _, line := range lines {
			fmt.Println("-	", line)
		}
		for _, dir := range parentDirs {
			fmt.Println("total	", dir)
		}
		return
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var names = customFlagArray("key")
	flag.Parse()
	var maps = customFlagMap(*names...)
	fmt.Println("keys", names)
	fmt.Println("values", maps)
	var analyser = gotoolcover{}
	j := analyser.analyseCoverage()
	internal.AlarmJson(j)

}

func customFlagArray(argName string) *StringList {
	var names StringList
	flag.Var(&names, argName, argName+"可指定多次,对应的value作为参数会2次解析")
	return &names
}
func customFlagMap(argNames ...string) *StringMap {
	fs := flag.NewFlagSet("name", -1)
	var ks = make(map[string]*string)
	m := &StringMap{}
	for _, argName := range argNames {
		var value = fs.String(argName, "", "--"+argName+"=val 会被解析入map")

		ks[argName] = value
	}
	err := fs.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
	}
	for argName, argValue := range ks {
		//fs.Var(m, argName, "--"+argName+"会被解析入map")

		fmt.Println(argName, *argValue)
		ks[argName] = argValue
	}
	return m
}

type gotoolcover struct{}

func (gotoolcover) analyseCoverage() *internal.JSONData {

	return internal.AnalyseGoToolsCover()
}
