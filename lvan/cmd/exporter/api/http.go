package api

import (
	_ "embed"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/spf13/pflag"
	"github.com/wangtengda0310/gobee/lvan/internal/workdir"
	"github.com/wangtengda0310/gobee/lvan/pkg/logger"
)

var htmlDirFlag *string
var router = http.DefaultServeMux

func init() {

	// 创建路由器并应用中间件
	router.HandleFunc("/", HandleRootRequest)
	router.HandleFunc("/cmd", HandleCommandRequest)
	router.HandleFunc("/cancel/", HandleCancelRequest)
	router.HandleFunc("/cmd/", HandleCommandRequest)
	router.HandleFunc("/result/", HandleResultRequest)
	router.HandleFunc("/backup/", HandleBackupRequest)

}
func Pflag(getEnvString func(key string, defaultVal string) string) {
	htmlDirFlag = pflag.String("html", getEnvString("EXPORTER_HTML_DIR", "html"), "指定HTML静态文件目录，支持环境变量 EXPORTER_HTML_DIR")

}

// 处理根路径请求
func HandleRootRequest(w http.ResponseWriter, r *http.Request) {
	// 只处理根路径请求
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// 使用嵌入的HTTP文档内容
	_, _ = w.Write([]byte(httpDoc))
}

// 嵌入HTTP文档
//
//go:embed http-doc.txt
var httpDoc string

func Router() http.Handler {

	// 如果指定了HTML目录，则注册/html/路径
	if *htmlDirFlag != "" {
		htmlDir := *htmlDirFlag
		// 确保目录存在
		if err := os.MkdirAll(htmlDir, 0755); err != nil {
			logger.Warn("无法创建HTML目录 %s: %v", htmlDir, err)
		} else {
			// 创建文件服务器handler
			fileServer := http.FileServer(http.Dir(workdir.Join(htmlDir)))
			// 注册 /html/ 路径
			router.Handle("/html/", http.StripPrefix("/html/", fileServer))
			logger.Info("HTML静态文件服务已启用，目录: %s，URL路径: /html/", htmlDir)
		}
	}

	// 使用中间件包装路由
	protectedRouter := recoveryMiddleware(router)
	return protectedRouter
}

// 新增恢复中间件
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("HTTP处理崩溃恢复: %v\n%s", err, debug.Stack())
				http.Error(w, "内部服务器错误", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
