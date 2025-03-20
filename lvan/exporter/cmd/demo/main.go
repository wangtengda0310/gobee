package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
)

func main() {
	// 全局参数（可选）
	help := pflag.Bool("help", false, "显示帮助信息")
	dir, _ := os.Getwd()
	w := pflag.StringP("workdir", "w", dir, "显示帮助信息")
	println(w)
	pflag.Parse()

	if *help {
		pflag.PrintDefaults()
		return
	}

	// 获取非标志参数（即不带--或-的参数）
	args := pflag.Args()
	if len(args) == 0 {
		fmt.Println("请提供子命令")
		os.Exit(1)
	}

	// 根据第一个参数判断子命令
	switch args[0] {
	case "commit":
		handleCommit(args[1:]) // 传递剩余参数给子命令
	case "clone":
		handleClone(args[1:])
	default:
		fmt.Printf("未知子命令: %s\n", args[0])
		os.Exit(1)
	}
}
func handleCommit(args []string) {
	cmd := pflag.NewFlagSet("commit", pflag.ExitOnError)
	message := cmd.StringP("message", "m", "", "提交信息")
	author := cmd.String("author", "", "提交者")
	cmd.Parse(args)

	// 必要参数检查
	if *message == "" {
		fmt.Println("必须提供提交信息（-m）")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("提交信息: %s\n作者: %s\n", *message, *author)
}
func handleClone(args []string) {
	cmd := pflag.NewFlagSet("clone", pflag.ExitOnError)
	repo := cmd.String("repo", "", "仓库地址")
	depth := cmd.Int("depth", 1, "克隆深度")
	cmd.Parse(args)

	if *repo == "" {
		fmt.Println("必须提供仓库地址（--repo）")
		cmd.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("克隆仓库: %s\n深度: %d\n", *repo, *depth)
}
