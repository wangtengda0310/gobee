// +build ignore

// Redis 测试辅助工具
// 用途：启动 miniredis 并输出连接信息供回归测试脚本使用
// 使用方式：go run tests/redis_test_helper.go > redis_env.txt &
//
// 输出格式（环境变量）：
// MINIREDIS_STARTED=true
// MINIREDIS_HOST=127.0.0.1
// MINIREDIS_PORT=<动态端口>
// MINIREDIS_ADDR=127.0.0.1:<动态端口>

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alicebob/miniredis/v2"
)

func main() {
	// 启动 miniredis
	s := miniredis.RunMiniServer(nil)

	// 输出环境变量格式
	fmt.Printf("MINIREDIS_STARTED=true\n")
	fmt.Printf("MINIREDIS_HOST=127.0.0.1\n")
	fmt.Printf("MINIREDIS_PORT=%d\n", s.Port())
	fmt.Printf("MINIREDIS_ADDR=%s\n", s.Addr())
	fmt.Printf("\n# 使用方式:\n")
	fmt.Printf("# source redis_env.txt\n")
	fmt.Printf("# export REDIS_PORT=$MINIREDIS_PORT\n")
	fmt.Printf("\n# 等待 Ctrl+C 停止...\n")

	// 设置信号处理，等待终止信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 清理
	s.Close()
	fmt.Println("\nMINIREDIS_STOPPED=true")
}
