package main

// tos_test.go: TOS 渠道测试.
// 离线部分: 签名向量对拍 (期望值由参考实现生成) / 配置门 / 清单采纳门 / 转义.
// 实测部分 (TestTOSLive*): 需真实环境变量 WINBOOT_TOS_* 注入才运行, 否则 Skip ——
// 凭证只经环境注入, 绝不写入代码或测试数据 (仓库合规红线).

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestTOSV1SignVectors 签名原语与参考实现 (Python hmac) 的向量对拍:
// base64url(HMAC-SHA256(sk, canonical))[:30].
func TestTOSV1SignVectors(t *testing.T) {
	cases := []struct{ sk, canonical, want string }{
		// 带 query 段 (bust 参数纳入签名)
		{"unit-test-secret",
			"GET\n/example-bucket/example-prefix/obj.exe\n_=abc123\n\n3600",
			"fHpf-kyBJs4dz-AJkQVK3AHOeYr1uw"},
		// 无 query 段 (sign-in-query 的 presign 形态)
		{"unit-test-secret",
			"GET\n/example-bucket/example-prefix/obj.exe\n\n3600",
			"W4VmFGIm0O51PmiDZBfcbn966Gyq9n"},
	}
	for _, c := range cases {
		if got := tosV1Sign(c.sk, c.canonical); got != c.want {
			t.Errorf("tosV1Sign canonical=%q\n got %q\nwant %q", c.canonical, got, c.want)
		}
	}
}

// TestTOSConfigured 配置门: 必填项任一缺失 = 渠道禁用 (本机注入了真实凭证时跳过).
func TestTOSConfigured(t *testing.T) {
	if os.Getenv("WINBOOT_TOS_AK") != "" {
		t.Skip("环境中已有真实 TOS 配置, 跳过禁用态断言")
	}
	if c := tosConfigured(); c != nil {
		t.Fatalf("无任何环境变量时渠道应禁用, got %+v", c)
	}
	t.Setenv("WINBOOT_TOS_AK", "ak")
	t.Setenv("WINBOOT_TOS_SK", "sk")
	if c := tosConfigured(); c != nil {
		t.Fatal("缺 bucket/host/service 时渠道应禁用")
	}
	t.Setenv("WINBOOT_TOS_BUCKET", "example-bucket")
	t.Setenv("WINBOOT_TOS_HOST", "example-bucket.example-endpoint")
	t.Setenv("WINBOOT_TOS_SERVICE", "example-service")
	t.Setenv("WINBOOT_TOS_PREFIX", "winboot/")
	t.Setenv("WINBOOT_TOS_GATEWAY", " 10.0.0.1 , 10.0.0.2 ,")
	c := tosConfigured()
	if c == nil {
		t.Fatal("配置齐全时渠道应启用")
	}
	if len(c.gateways) != 2 {
		t.Fatalf("网关列表应解析为 2 项 (容忍空白与尾逗号), got %v", c.gateways)
	}
	if c.prefix != "winboot/" {
		t.Fatalf("前缀解析错误: %q", c.prefix)
	}
}

// TestTOSAdopt 清单采纳门: flavor 门 (防跨栈蚕食) + 版本门 + 伪协议 URL 化.
func TestTOSAdopt(t *testing.T) {
	mk := func(s string) []byte { return []byte(s) }
	// 无 flavor 字段 (异构栈清单): 拒绝 —— 这是防蚕食的关键断言
	if r := tosAdopt(mk(`{"version":"9.9.9","sha256":"ab","size":1}`), "0.4.0", TOOL_FLAVOR); r != nil {
		t.Fatalf("未标注 flavor 的清单应被拒绝, got %+v", r)
	}
	// flavor 为别栈: 拒绝
	if r := tosAdopt(mk(`{"version":"9.9.9","flavor":"other","size":1}`), "0.4.0", TOOL_FLAVOR); r != nil {
		t.Fatalf("异构 flavor 的清单应被拒绝, got %+v", r)
	}
	// 同 flavor 但版本不新: 不采纳
	if r := tosAdopt(mk(fmt.Sprintf(`{"version":"0.4.0","flavor":%q,"size":1}`, TOOL_FLAVOR)), "0.4.0", TOOL_FLAVOR); r != nil {
		t.Fatalf("同版本不应触发更新, got %+v", r)
	}
	// 同 flavor 且版本新: 采纳, URL 伪协议化 (显式 url 与缺省两种)
	r := tosAdopt(mk(fmt.Sprintf(`{"version":"0.5.0","url":"winboot.exe","flavor":%q,"size":1}`, TOOL_FLAVOR)), "0.4.0", TOOL_FLAVOR)
	if r == nil || r.URL != "tos:winboot.exe" || r.Version != "0.5.0" {
		t.Fatalf("应采纳同 flavor 新版本并伪协议化 URL, got %+v", r)
	}
	if r = tosAdopt(mk(fmt.Sprintf(`{"version":"0.5.0","flavor":%q,"size":1}`, TOOL_FLAVOR)), "0.4.0", TOOL_FLAVOR); r == nil || r.URL != "tos:winboot.exe" {
		t.Fatalf("缺省 url 应补 tos:winboot.exe, got %+v", r)
	}
	// 非法 JSON: 拒绝
	if r := tosAdopt(mk("not json"), "0.4.0", TOOL_FLAVOR); r != nil {
		t.Fatalf("非法 JSON 应返回 nil")
	}
}

// TestTOSQueryEscape 转义与参考实现 (Python quote, 默认斜杠保留) 字节一致.
func TestTOSQueryEscape(t *testing.T) {
	got := tosQueryEscape("EXAMPLEAK/20260929/cn-beijing/tos/sig_request")
	if got != "EXAMPLEAK/20260929/cn-beijing/tos/sig_request" {
		t.Fatalf("斜杠与字母数字不应被转义: %q", got)
	}
	if got := tosQueryEscape("a b+c"); got != "a%20b%2Bc" {
		t.Fatalf("空格应转 %%20 (非 +), 加号应转义: %q", got)
	}
}

// TestTOSLiveManifest 实测: 拉取桶内清单 (协议正确性验证: 签名/路由/缓存破除全链).
func TestTOSLiveManifest(t *testing.T) {
	c := liveTOS(t)
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := tosOpen(c, client, c.prefix+manifestName, fmt.Sprint(time.Now().UnixNano()))
	if err != nil {
		t.Fatalf("拉取清单失败: %v", err)
	}
	defer resp.Body.Close()
	body := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	for {
		n, rerr := resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		if rerr != nil {
			break
		}
	}
	if !strings.Contains(string(body), `"version"`) {
		t.Fatalf("清单内容异常 (无 version 字段): %.200s", body)
	}
	t.Logf("清单 (%d B): %.300s", len(body), body)
}

// TestTOSLivePresign 实测: 门票签发 -> 直接 GET 回环 (Host 头路由 + 签名时效).
func TestTOSLivePresign(t *testing.T) {
	c := liveTOS(t)
	if len(c.gateways) == 0 {
		t.Skip("未配网关 IP, 跳过门票回环测试")
	}
	url, host := TOSPresignURL(manifestName, 5*time.Minute)
	if url == "" {
		t.Fatal("配置齐全应能签发门票")
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Host = host
	resp, err := (&http.Client{Timeout: 8 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("门票 GET 失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("门票 GET 状态 %d", resp.StatusCode)
	}
	body := make([]byte, 300)
	n, _ := resp.Body.Read(body)
	if !strings.Contains(string(body[:n]), `"version"`) {
		t.Fatalf("门票取回内容异常: %.200s", body[:n])
	}
	t.Log("门票回环 OK:", url[:min(60, len(url))], "...")
}

// liveTOS 实测前置: 环境未注入凭证时跳过; 返回已校验的配置.
func liveTOS(t *testing.T) *tosConfig {
	t.Helper()
	if os.Getenv("WINBOOT_TOS_AK") == "" {
		t.Skip("未注入 WINBOOT_TOS_* 环境变量, 跳过实测")
	}
	c := tosConfigured()
	if c == nil {
		t.Fatal("环境变量已注入但配置校验未通过 (缺必填项)")
	}
	return c
}
