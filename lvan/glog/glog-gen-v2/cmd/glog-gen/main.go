package main

//go:generate go run main.go -xml ../../testdata/demo.xml -mapping ../../config/type_mapping.xml -out ../../output/logs
import (
	"flag"
	"fmt"
	"glog-gen-v2/internal"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("main开始")
	var xmlPath, excelPath, mappingPath, outDir, tmplDir string
	flag.StringVar(&xmlPath, "xml", "", "日志结构XML文件路径")
	flag.StringVar(&excelPath, "excel", "", "日志结构Excel文件路径")
	flag.StringVar(&mappingPath, "mapping", "./config/type_mapping.xml", "类型映射配置文件路径")
	flag.StringVar(&outDir, "out", "./output", "代码输出目录")
	flag.StringVar(&tmplDir, "tmplDir", "./templates", "模板文件夹路径")
	flag.Parse()

	internal.GlobalTemplateDir = tmplDir

	if (xmlPath == "" && excelPath == "") || (xmlPath != "" && excelPath != "") {
		log.Fatal("请且只能指定 -xml 或 -excel 参数中的一个")
	}
	if xmlPath != "" {
		if _, err := os.Stat(xmlPath); err != nil {
			fmt.Fprintf(os.Stderr, "XML文件不存在: %s\n", xmlPath)
			os.Exit(1)
		}
	}
	if excelPath != "" {
		if _, err := os.Stat(excelPath); err != nil {
			fmt.Fprintf(os.Stderr, "Excel文件不存在: %s\n", excelPath)
			os.Exit(1)
		}
	}
	if _, err := os.Stat(mappingPath); err != nil {
		fmt.Fprintf(os.Stderr, "类型映射文件不存在: %s\n", mappingPath)
		os.Exit(1)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("准备生成代码...")
	var genErr error
	if xmlPath != "" {
		parse := func() ([]internal.Struct, error) { return internal.ParseXML(xmlPath) }
		genErr = internal.Generate(parse, mappingPath, outDir)
	} else {
		parse := func() ([]internal.Struct, error) { return internal.ParseExcel(excelPath) }
		genErr = internal.Generate(parse, mappingPath, outDir)
	}
	if genErr != nil {
		fmt.Fprintf(os.Stderr, "代码生成失败: %+v\n", genErr)
		os.Exit(1)
	}
	fmt.Println("代码生成完成。请检查输出目录：", filepath.Clean(outDir))

	// 自动运行单元测试
	absOut, _ := filepath.Abs(outDir)
	cmd := exec.Command("go", "test", "-v", absOut)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("自动测试失败:", err)
		os.Exit(2)
	}
	fmt.Println("自动测试通过！")
}
