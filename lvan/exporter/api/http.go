package api

import (
	_ "embed"
	"net/http"
)

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
	w.Write([]byte(httpDoc))
}

// 嵌入HTTP文档
//
//go:embed http-doc.txt
var httpDoc string
