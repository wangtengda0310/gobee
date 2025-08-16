package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// StartMCPServer 启动MCP服务器
func StartMCPServer() error {
	// 创建MCP服务器
	s := server.NewMCPServer(
		"绿岸 FTP MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	//注册FTP上传文件工具
	s.AddResource(
		mcp.NewResource("ftp://list", "ftp服务器文件列表工具",
			mcp.WithResourceDescription("列出服务器上的文件列表"),
			mcp.WithMIMEType("application/json"),
		),
		listFtpFiles,
	)

	// 注册FTP上传文件工具
	s.AddTool(
		mcp.NewTool("绿岸ftp上传工具",
			mcp.WithDescription("上传文件到FTP服务器"),
			mcp.WithArray("文件", mcp.Required(), mcp.Description("需要上传的文件路径")),
			mcp.WithArray("ftp服务器", mcp.Description("需要上传的文件路径")),
			mcp.WithArray("账号", mcp.Description("需要上传的文件路径")),
			mcp.WithArray("密码", mcp.Description("需要上传的文件路径")),
		),
		handleFTPUpload,
	)

	// 启动标准输入输出服务器
	log.Println("FTP Client MCP服务器已启动...")
	return server.ServeStdio(s)
}

func listFtpFiles(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	var sb strings.Builder
	arguments := req.Params.Arguments
	reqhost := host
	requser := user
	reqpass := pass
	if h, ok := arguments["ftp服务器"]; ok {
		reqhost = h.(string)
	}
	if u, ok := arguments["账号"]; ok {
		requser = u.(string)
	}
	if p, ok := arguments["密码"]; ok {
		reqpass = p.(string)
	}
	e := openServer(
		reqhost, requser, reqpass,
		func(conn *ftp.ServerConn) {
			entries, err := conn.List(".")
			if err != nil {
				return
			}
			for _, entry := range entries {
				sb.WriteString(fmt.Sprintf("%s\n", entry.Name))
				log.Println(entry.Name)
			}
		},
	)
	if e != nil {
		return nil, e
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "application/json",
			Text:     sb.String(),
		},
	}, nil
}

func get_config(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("配置查询成功"), nil
}

// handleFTPUpload 处理FTP上传请求
func handleFTPUpload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	paths, err := req.RequireStringSlice("文件")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	var e error
	err = openServer(
		req.GetString("ftp服务器", host),
		req.GetString("账号", user),
		req.GetString("密码", pass),
		func(conn *ftp.ServerConn) {
			for _, path := range paths {
				e = uploadPath(conn, path)
			}
			sb.WriteString("上传完成")
		})

	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if e != nil {
		return mcp.NewToolResultError(e.Error()), nil
	}
	sb.WriteString("上传成功")
	return mcp.NewToolResultText(sb.String()), nil
}

// ftpDialAndLogin 连接并登录到FTP服务器
func ftpDialAndLogin(serverAddr, username, password string) (*ftp.ServerConn, error) {
	conn, err := ftp.Dial(serverAddr, ftp.DialWithTimeout(time.Duration(3)*time.Second))
	if err != nil {
		return nil, err
	}

	if err := conn.Login(username, password); err != nil {
		conn.Quit()
		return nil, err
	}

	return conn, nil
}
