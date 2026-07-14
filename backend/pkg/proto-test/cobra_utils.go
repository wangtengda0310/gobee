package prototest

import "strings"

// deriveHTTPAddr 从 TCP 服务器地址推导默认 HTTP 认证地址。
// 规则：取 TCP 地址的 IP 部分，端口固定为 20144。
// 若 tcpAddr 不含端口，则直接追加 ":20144"。
func deriveHTTPAddr(tcpAddr string) string {
	ip := tcpAddr
	if idx := strings.LastIndex(ip, ":"); idx >= 0 {
		ip = ip[:idx]
	}
	return ip + ":20144"
}
