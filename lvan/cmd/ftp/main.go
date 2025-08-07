package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jlaffaye/ftp"
)

//go:embed host
var host string

//go:embed username
var user string

//go:embed password
var pass string

func main() {
	// 解析命令行参数（保留原逻辑）
	server := flag.String("h", fmt.Sprintf("%s:21", host), "FTP服务器地址和端口")
	username := flag.String("u", user, "FTP用户名")
	password := flag.String("p", pass, "FTP密码")
	filename := flag.String("f", "", "远程文件名")
	flag.Parse()

	if *filename == "" {
		s := uuid.New().String()
		filename = &s
	}

	// 连接FTP服务器
	conn, err := ftp.Dial(*server, ftp.DialWithTimeout(time.Duration(3)*time.Second))
	if err != nil {
		log.Fatalf("连接FTP服务器失败: %v", err)
	}
	defer conn.Quit()

	// 登录
	if err := conn.Login(*username, *password); err != nil {
		log.Fatalf("登录失败: %v", err)
	}

	// 上传文件（直接使用字符串的 Reader）
	if err := conn.Stor(*filename, os.Stdin); err != nil {
		log.Fatalf("文件上传失败: %v", err)
	}

	fmt.Printf("%s 文件上传成功", *filename)
}
