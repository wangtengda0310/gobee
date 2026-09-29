package main

// updater.go: 自动更新 (参考 PyQt 实现移植).
//
// 两条清单渠道 (优先级见 hub.updateWorker): TOS 对象存储 (tos.go, env 配置启用,
// 清单须带本栈 flavor) 与部署中心 HTTP (下方 UPDATE_BASE).
// 部署侧两个静态文件: <UPDATE_BASE>/winboot-manifest.json {"version","url","sha256","size"}
//                    <UPDATE_BASE>/winboot.exe (url 指向它, 可相对; TOS 渠道则为 "tos:<file>" 伪协议)
// 热替换: 运行中的 exe 不能覆写但可以改名 -> 自己改名 .old 让路 -> 新文件就位
//        -> 拉起新进程 -> 旧进程退出; 新进程启动时清 .old.
// 注意: UPDATE_BASE 留空 = 关闭 (默认关; 部署自己的更新服务后填写).
// 环境变量 WINBOOT_UPDATE_BASE 可临时覆盖.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const updateBase = "" // 部署地址; 发布前设置
const manifestName = "winboot-manifest.json"

type updateInfo struct {
	Version string `json:"version"`
	URL     string `json:"url"`
	SHA256  string `json:"sha256"`
	Size    int64  `json:"size"`
	Flavor  string `json:"flavor,omitempty"` // 栈身份 (TOS 清单采纳的门禁; 部署中心清单不校验)
}

func updateBaseURL() string {
	if v := os.Getenv("WINBOOT_UPDATE_BASE"); v != "" {
		return v
	}
	return updateBase
}

// ConfiguredBase 本机配置的更新中心地址 (环境变量 > 代码内配置; 空串 = 未配置).
func ConfiguredBase() string { return updateBaseURL() }

// CheckUpdate 有新版返回 manifest 信息, 无新版/出错返回 nil (静默).
// base 为空时用本机配置 (env > 代码内).
func CheckUpdate(base string) *updateInfo {
	if base == "" {
		base = updateBaseURL()
	}
	if base == "" {
		return nil
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(base + "/" + manifestName)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var info updateInfo
	if json.NewDecoder(resp.Body).Decode(&info) != nil {
		return nil
	}
	if VersionGt(info.Version, TOOL_VERSION) {
		return &info
	}
	return nil
}

// DownloadUpdate 按清单下载并校验 sha256/size; 成功返回本地临时路径.
// "tos:<file>" 伪协议走 TOS 渠道 (对象位于发布前缀下, bust 用内容 sha 前 16 位
// 破 CDN 缓存), 其余按部署中心 HTTP 下载.
func DownloadUpdate(info *updateInfo, log func(string)) (string, error) {
	dest := filepath.Join(os.TempDir(), fmt.Sprintf("winboot_update_%s.exe", info.Version))
	var body io.ReadCloser
	if file := strings.TrimPrefix(info.URL, "tos:"); strings.HasPrefix(info.URL, "tos:") && file != "" {
		c := tosConfigured()
		if c == nil {
			return "", fmt.Errorf("TOS 渠道未配置")
		}
		bust := ""
		if len(info.SHA256) >= 16 {
			bust = strings.ToLower(info.SHA256[:16])
		}
		resp, err := tosOpen(c, &http.Client{Timeout: 10 * time.Minute}, c.prefix+file, bust)
		if err != nil {
			return "", err
		}
		body = resp.Body
	} else {
		base := updateBaseURL()
		url := info.URL
		if url != "" && !strings.Contains(url, "://") {
			url = base + "/" + url
		}
		resp, err := (&http.Client{Timeout: 10 * time.Minute}).Get(url)
		if err != nil {
			return "", err
		}
		body = resp.Body
	}
	defer body.Close()
	return saveVerified(body, dest, info, log)
}

// saveVerified 流式落盘 + 进度日志 + sha256/size 校验 (两个渠道共用).
func saveVerified(body io.Reader, dest string, info *updateInfo, log func(string)) (string, error) {
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	var total int64
	nextLog := int64(50 << 20)
	buf := make([]byte, 1<<20)
	for {
		n, rerr := body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				f.Close()
				return "", werr
			}
			h.Write(buf[:n])
			total += int64(n)
			if total >= nextLog {
				log(fmt.Sprintf("新版本下载中... %dMB", total>>20))
				nextLog += 50 << 20
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			f.Close()
			return "", rerr
		}
	}
	f.Close()
	if info.Size > 0 && total != info.Size {
		os.Remove(dest)
		return "", fmt.Errorf("尺寸不符: 期望 %d, 实得 %d", info.Size, total)
	}
	if info.SHA256 != "" && hex.EncodeToString(h.Sum(nil)) != info.SHA256 {
		os.Remove(dest)
		return "", fmt.Errorf("sha256 校验失败")
	}
	return dest, nil
}

// ApplyUpdate 热替换 (改名让路 + 就位失败回滚); 成功 true.
func ApplyUpdate(newExe string) bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	old := exe + ".old"
	if _, err := os.Stat(old); err == nil {
		if err := os.Remove(old); err != nil {
			return false
		}
	}
	if err := os.Rename(exe, old); err != nil {
		return false
	}
	if err := os.Rename(newExe, exe); err != nil {
		_ = os.Rename(old, exe) // 回滚: 绝不留空位
		return false
	}
	return true
}

// RestartSelf 拉起新进程 (调用方随后退出).
func RestartSelf() {
	if exe, err := os.Executable(); err == nil {
		cmd := exec.Command(exe)
		cmd.SysProcAttr = hideWindow()
		_ = cmd.Start()
	}
}

// CleanupOld 启动时清理上次更新留下的 .old (占用删不掉留到下次).
func CleanupOld() {
	if exe, err := os.Executable(); err == nil {
		_ = os.Remove(exe + ".old")
	}
}
