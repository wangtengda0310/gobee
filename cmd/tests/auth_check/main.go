package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// 关键诊断：对比 streamproto AuthLogin 和 游戏客户端 LoginServer 的响应
// 看 open_id 字段是什么
func main() {
	testUIDs := []string{"test9", "test95", "test96"}
	httpAddr := "10.254.114.204:20144"

	log.SetFlags(log.Ltime)

	for _, uid := range testUIDs {
		log.Printf("=== AuthLogin uid=%s ===", uid)
		body := fmt.Sprintf(`{"uid":"%s","platform":13,"sdk":0}`, uid)
		resp, err := http.Post("http://"+httpAddr+"/authlogin",
			"application/json", bytes.NewBufferString(body))
		if err != nil {
			log.Printf("  HTTP失败: %v", err)
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// 打印原始响应
		var raw map[string]interface{}
		json.Unmarshal(respBody, &raw)
		prettyJSON, _ := json.MarshalIndent(raw, "  ", "  ")
		fmt.Fprintf(os.Stderr, "  Response:\n  %s\n", string(prettyJSON))

		// 解析关键字段
		var r struct {
			Code       int    `json:"code"`
			OpenID     string `json:"open_id"`
			Token      string `json:"token"`
			ServerAddr string `json:"server_addr"`
		}
		json.Unmarshal(respBody, &r)
		log.Printf("  code=%d open_id=%q token=%s... server=%s",
			r.Code, r.OpenID, r.Token[:minInt(16, len(r.Token))], r.ServerAddr)

	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
