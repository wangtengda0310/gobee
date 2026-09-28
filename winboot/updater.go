package main

// updater.go: 自动更新 (参考 PyQt 实现移植).
//
// 部署侧两个静态文件: <UPDATE_BASE>/winboot-manifest.json {"version","url","sha256","size"}
//                    <UPDATE_BASE>/winboot.exe (url 指向它, 可相对)
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
}

func updateBaseURL() string {
	if v := os.Getenv("WINBOOT_UPDATE_BASE"); v != "" {
		return v
	}
	return updateBase
}

// CheckUpdate 有新版返回 manifest 信息, 无新版/出错返回 nil (静默).
func CheckUpdate() *updateInfo {
	base := updateBaseURL()
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
func DownloadUpdate(info *updateInfo, log func(string)) (string, error) {
	base := updateBaseURL()
	url := info.URL
	if url != "" && !strings.Contains(url, "://") {
		url = base + "/" + url
	}
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	dest := filepath.Join(os.TempDir(), fmt.Sprintf("winboot_update_%s.exe", info.Version))
	f, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	var total int64
	nextLog := int64(50 << 20)
	buf := make([]byte, 1 << 20)
	for {
		n, rerr := resp.Body.Read(buf)
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
