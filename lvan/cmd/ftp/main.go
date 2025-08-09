package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

type StdinMode int

const (
	StdinModeNone  StdinMode = iota // 无标准输入
	StdinModeInput                  // 无标准输入
	StdinModePipe
	StdinModeRedirect // 标准输入被重定向
)

func SwitchStdinMode() StdinMode {
	stat, e := os.Stdin.Stat()
	if e != nil {
		panic(e)
	}
	charDevice := (stat.Mode() & os.ModeCharDevice) != 0
	switch {
	case stat.Size() > 0 && !charDevice:
		return StdinModeRedirect
	case stat.Size() == 0 && !charDevice:
		return StdinModePipe
	case stat.Size() == 0 && charDevice: // /dev/urandom 也落在这里
		return StdinModeInput
	default:
		return StdinModeNone
	}
}
func main() {
	// 解析命令行参数
	server := flag.String("h", fmt.Sprintf("%s:21", host), "FTP服务器地址和端口")
	username := flag.String("u", user, "FTP用户名")
	password := flag.String("p", pass, "FTP密码")
	filename := flag.String("f", "", "远程文件名")
	getfile := flag.String("g", "", "下载远程文件到标准输出")
	flag.Parse()

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

	if *getfile != "" {
		err := streamToStdout(conn, *getfile)
		if err != nil {
			log.Fatalf("下载文件失败: %v", err)
		}
		return
	}

	// 获取非flag参数
	args := flag.Args()

	if *filename == "" {

		hasFiles := len(args) > 0

		mode := SwitchStdinMode()
		switch {
		// 模式1: 有文件参数时忽略 stdin
		case hasFiles && mode == StdinModeInput:
			println("只处理文件参数，不处理标准输入")
			storFiles(conn, args)

		// 模式2: 有文件参数且有重定向的 stdin
		case hasFiles:
			println("处理文件参数和重定向的标准输入")

			storStdin(conn)
			storFiles(conn, args)
		// 模式3: 无参数时等待用户输入
		default:
			if mode == StdinModeInput {
				fmt.Println("Waiting for user input (press Ctrl+D to finish):")
			}
			storStdin(conn)
		}
	} else {
		mergeReader(conn, filename, args)
	}

	//
	//// 处理文件和文件夹参数
	//if len(args) > 0 {
	//	for _, path := range args {
	//		if err := uploadPath(conn, path, *filename); err != nil {
	//			log.Printf("上传 %s 失败: %v", path, err)
	//		} else {
	//			fmt.Printf("%s 上传成功\n", path)
	//		}
	//	}
	//}
}

// 核心函数：将FTP文件流式输出到stdout
func streamToStdout(conn *ftp.ServerConn, remotePath string) error {
	// 获取文件读取流
	r, err := conn.Retr(remotePath)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer r.Close()

	// 将内容直接复制到标准输出
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		return fmt.Errorf("输出到标准输出失败: %w", err)
	}

	return nil
}

func mergeReader(conn *ftp.ServerConn, filename *string, args []string) {
	var r []io.Reader
	hasFiles := len(args) > 0

	if !hasFiles || SwitchStdinMode() == StdinModePipe || SwitchStdinMode() == StdinModeRedirect {
		r = append(r, os.Stdin)
	}
	var closer []func() error
	for _, file := range args {
		fr, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		closer = append(closer, fr.Close)
		r = append(r, fr)
	}
	defer func() {
		for _, f := range closer {
			_ = f()
		}
	}()
	reader := io.MultiReader(r...)
	err := conn.Stor(*filename, reader)
	if err != nil {
		panic(err)
	}
}

func storStdin(conn *ftp.ServerConn) {
	var remoteName = uuid.New().String()
	err := conn.Stor(remoteName, os.Stdin)
	if err != nil {
		panic(fmt.Errorf("%s %w", remoteName, err))
	}
	fmt.Printf("标准输入内容已上传为 %s\n", remoteName)
}

// 上传标准输入内容
func storFiles(conn *ftp.ServerConn, args []string) {
	// 添加args中的文件内容
	for _, path := range args {
		uploadPath(conn, path)
	}

	return
}

// 上传文件或文件夹
func uploadPath(conn *ftp.ServerConn, localPath string) error {
	// 检查本地路径是否存在
	fi, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		// 处理文件夹
		return uploadDirectory(conn, localPath)
	} else {
		return uploadFile(conn, localPath, fi.Name())
	}
}

// 上传文件
func uploadFile(conn *ftp.ServerConn, localPath, remoteName string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = conn.Stor(remoteName, file)
	if err == nil {
		fmt.Printf("已上传文件 %s\n", remoteName)
	}
	return err
}

// 上传文件夹（递归）
func uploadDirectory(conn *ftp.ServerConn, localPath string) error {
	// 获取文件夹名称
	baseName := filepath.Base(localPath)

	// 创建远程文件夹
	if err := conn.MakeDir(baseName); err != nil {
		// 忽略文件夹已存在的错误
		// if !ftp.IsErrorPermission(err) {
		// 	return err
		// }
		panic(err)
	}

	// 切换到远程文件夹
	if err := conn.ChangeDir(baseName); err != nil {
		return err
	}
	defer conn.ChangeDir("..")

	// 遍历本地文件夹
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(localPath, entry.Name())
		if entry.IsDir() {
			// 递归上传子文件夹
			if err := uploadDirectory(conn, entryPath); err != nil {
				return err
			}
		} else {
			// 上传文件
			file, err := os.Open(entryPath)
			if err != nil {
				return err
			}
			defer file.Close()

			if err := conn.Stor(entry.Name(), file); err != nil {
				return err
			}
		}
	}

	return nil
}
