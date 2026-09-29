package main

// tos.go: TOS 对象存储更新渠道 (纯标准库 V1 签名, 零第三方依赖).
//
// 渠道语义 (与部署中心渠道并存, 优先级见 hub.updateWorker):
//   清单 = 桶内 <PREFIX>winboot-manifest.json, exe = <PREFIX>winboot.exe;
//   check 有新版返回 manifest (url 为 "tos:<file>" 伪协议), download 按伪协议分流.
//
// 配置全部经环境变量注入 (本仓库为个人项目, 不携带任何环境相关信息;
// 任一必填项缺失 = 渠道禁用, 与"未配置部署中心 = 关闭"同构):
//   WINBOOT_TOS_AK / WINBOOT_TOS_SK  个人 AccessKey/SecretKey (V1 签名用)
//   WINBOOT_TOS_BUCKET               桶名
//   WINBOOT_TOS_HOST                 virtual-hosted 访问域名 (如 <bucket>.<endpoint>)
//   WINBOOT_TOS_SERVICE              Destination-Service 头值 (api 网关按此路由)
//   WINBOOT_TOS_PREFIX               发布前缀, 如 "winboot/" (可空 = 桶根)
//   WINBOOT_TOS_GATEWAY              网关 IP 直连 fallback, 逗号分隔 (可空 = 仅域名)
//
// 防跨栈蚕食: TOSCheckUpdate 只采纳 flavor 为本栈 (TOOL_FLAVOR) 的清单 ——
// 异构栈或未标注 flavor 的清单一律忽略, 防止把自己"升级"成别栈可执行文件.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// tosConfig 环境变量注入的渠道配置; tosConfigured 返回 nil 表示渠道禁用.
type tosConfig struct {
	ak, sk, bucket, host, service, prefix string
	gateways                              []string
}

// tosConfigured 读取并校验配置; 任一必填项缺失返回 nil (调用方据此静默禁用).
func tosConfigured() *tosConfig {
	c := &tosConfig{
		ak:      strings.TrimSpace(os.Getenv("WINBOOT_TOS_AK")),
		sk:      strings.TrimSpace(os.Getenv("WINBOOT_TOS_SK")),
		bucket:  strings.TrimSpace(os.Getenv("WINBOOT_TOS_BUCKET")),
		host:    strings.TrimSpace(os.Getenv("WINBOOT_TOS_HOST")),
		service: strings.TrimSpace(os.Getenv("WINBOOT_TOS_SERVICE")),
		prefix:  os.Getenv("WINBOOT_TOS_PREFIX"),
	}
	if g := os.Getenv("WINBOOT_TOS_GATEWAY"); g != "" {
		for _, ip := range strings.Split(g, ",") {
			if ip = strings.TrimSpace(ip); ip != "" {
				c.gateways = append(c.gateways, ip)
			}
		}
	}
	if c.ak == "" || c.sk == "" || c.bucket == "" || c.host == "" || c.service == "" {
		return nil
	}
	return c
}

// tosV1Sign V1 签名原语: HMAC-SHA256(SK, canonical) 的 base64url 前 30 字符.
// 独立成纯函数便于与参考实现做向量对拍 (见 tos_test.go).
func tosV1Sign(sk, canonical string) string {
	mac := hmac.New(sha256.New, []byte(sk))
	mac.Write([]byte(canonical))
	return base64.URLEncoding.EncodeToString(mac.Sum(nil))[:30]
}

// tosV1Auth 组装 X-Tos-Signature 头值.
// canonical 各段: METHOD / /bucket/key / 排序后的已签 query / 签名者(空) / 过期秒.
// 已签 query = 请求中排除集外的一切参数 —— 自加的破缓存参数 "_" 必须纳入,
// 而 timeout 与 tos-* 类参数在服务端排除集内 (不签, 也不出现在 canonical).
func tosV1Auth(c *tosConfig, method, key string, expired int64, signedQuery map[string]string) string {
	items := []string{method, "/" + c.bucket + "/" + key}
	if len(signedQuery) > 0 {
		keys := make([]string, 0, len(signedQuery))
		for k := range signedQuery {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		pairs := make([]string, 0, len(keys))
		for _, k := range keys {
			pairs = append(pairs, k+"="+signedQuery[k])
		}
		items = append(items, strings.Join(pairs, "&"))
	}
	items = append(items, "", fmt.Sprint(expired))
	sign := tosV1Sign(c.sk, strings.Join(items, "\n"))
	cred := fmt.Sprintf("%s/%s/cn-beijing/tos/sig_request", c.ak, time.Now().Format("20060102"))
	return fmt.Sprintf("TOS-HMAC-SHA256 expiration=%d,signame=,signature=%s,credentials=%s",
		expired, sign, cred)
}

// tosOpen 打开桶内对象的 HTTP 响应 (调用方负责 Close).
// 目标地址: 域名优先 (部分网络可解析), 失败逐个网关 IP 直连 (Host 头伪装成域名 ——
// V1 签名不覆盖 host, 网关按 Host 路由). 注意直连 IP 走大流量不受内网域名限速.
// bust: 破缓存参数 —— 网关默认给对象 30 天 Cache-Control, 同名覆盖后 CDN 会继续
// 发旧内容; manifest 用时间戳 (每次必回源), exe 用内容 sha (同内容同 URL, 内容变
// URL 必变). 全部目标失败返回错误 (调用方静默处理).
func tosOpen(c *tosConfig, client *http.Client, key, bust string) (*http.Response, error) {
	targets := append([]string{c.host}, c.gateways...)
	var lastErr error
	for _, t := range targets {
		u := fmt.Sprintf("http://%s/%s?timeout=10s", t, key)
		if bust != "" {
			u += "&_=" + bust
		}
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			lastErr = err
			continue
		}
		var sq map[string]string
		if bust != "" {
			sq = map[string]string{"_": bust}
		}
		req.Header.Set("X-Tos-Signature", tosV1Auth(c, "GET", key, time.Now().Unix()+600, sq))
		req.Header.Set("Destination-Service", c.service)
		req.Header.Set("Cache-Control", "no-cache")
		if t != c.host {
			req.Host = c.host // IP 直连时伪装域名 (网关按 Host 路由)
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("%s: HTTP %d", t, resp.StatusCode)
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("TOS 渠道不可达: %v", lastErr)
}

// TOSCheckUpdate 查 TOS 桶的发布清单; 有新版返回 manifest (URL 置为 "tos:<file>"
// 伪协议, 供 DownloadUpdate 分流), 无新版/未配置/失败一律返回 nil (静默, 不影响
// 其余渠道).
func TOSCheckUpdate() *updateInfo {
	c := tosConfigured()
	if c == nil {
		return nil
	}
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := tosOpen(c, client, c.prefix+manifestName, fmt.Sprint(time.Now().UnixNano()))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil
	}
	return tosAdopt(body, TOOL_VERSION, TOOL_FLAVOR)
}

// tosAdopt 解析并采纳清单: flavor 门 (清单 flavor 非本栈, 含未标注, 一律不采纳 ——
// 防跨栈蚕食) + 版本门 (须比 current 新) + URL 伪协议化 ("tos:<file>", 缺省文件名
// winboot.exe). 独立成纯函数便于离线测试.
func tosAdopt(body []byte, current, flavor string) *updateInfo {
	var info updateInfo
	if json.Unmarshal(body, &info) != nil {
		return nil
	}
	if info.Flavor != flavor {
		return nil
	}
	if !VersionGt(info.Version, current) {
		return nil
	}
	if info.URL == "" {
		info.URL = "tos:winboot.exe"
	} else {
		info.URL = "tos:" + info.URL
	}
	return &info
}

// tosQueryEscape 等价参考实现的 query 转义 (斜杠保留): credentials 串含 '/',
// 通用 QueryEscape 会转成 %2F —— 服务端虽然通常能解, 但保持与实测可用形态
// 字节一致更稳妥, 不引入未验证变量.
func tosQueryEscape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || ch >= '0' && ch <= '9' ||
			ch == '-' || ch == '_' || ch == '.' || ch == '~' || ch == '/' {
			b.WriteByte(ch)
		} else {
			fmt.Fprintf(&b, "%%%02X", ch)
		}
	}
	return b.String()
}

// TOSPresignURL 为桶内对象签发临时直读 URL (sign-in-query 形态).
//
// 引导场景 ("首份可执行文件"获取链) 的凭证零传播: 持有凭证的实例随局域网发现
// 应答现场签发, 消费方拿到的只是 "ttl 内可下载这一个文件" 的门票, 凭证永不
// 离开本进程. 返回 (url, host): URL 为网关 IP 直连形态 (V1 签名不覆盖 host,
// IP+Host 头依然有效), host 是消费方下载时必须设置的 Host 头值.
// 未配置渠道或未配网关 IP 时返回 ("", "") —— 门票必须是直连形态, 消费方
// (引导脚本) 零 IP 知识零 fallback 逻辑.
func TOSPresignURL(key string, ttl time.Duration) (string, string) {
	c := tosConfigured()
	if c == nil || len(c.gateways) == 0 {
		return "", ""
	}
	fullKey := c.prefix + key
	expired := time.Now().Unix() + int64(ttl/time.Second)
	// sign-in-query 的 canonical 无 query 段: tos-* 参数全在排除集
	sign := tosV1Sign(c.sk, strings.Join(
		[]string{"GET", "/" + c.bucket + "/" + fullKey, "", fmt.Sprint(expired)}, "\n"))
	cred := fmt.Sprintf("%s/%s/cn-beijing/tos/sig_request", c.ak, time.Now().Format("20060102"))
	q := fmt.Sprintf("tos-algorithm=TOS-HMAC-SHA256&tos-credentials=%s&tos-expiration=%d"+
		"&tos-signame=&tos-signature=%s", tosQueryEscape(cred), expired, sign)
	return fmt.Sprintf("http://%s/%s?%s", c.gateways[0], fullKey, q), c.host
}
