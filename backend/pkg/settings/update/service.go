// Package update 提供 Android APK 自动更新服务(非 Play 分发)。
//
// 流程:前端 getAppVersionCode(Java 桥读 build.gradle versionCode) → CheckUpdate
// 对比服务端 latest.json versionCode → 有新版则 DownloadApk 下载到 files/updates/
// (边下边算 SHA256 + Event 推进度) → 前端调 Java installApk 触发系统安装器。
//
// 更新源:https://itsnot.fun/rain-qa-func/latest.json(自建 HTTP,DigiCert 正式证书,
// Android 默认信任,无需 cleartext 配置)。
// Wails v3 自带 updater 桌面专用(binary swap 机制)不适用 Android,详见 docs/Android-自动更新.md。
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/appconfig"
)

// UpdateJSONURL 服务端版本元数据 URL(自建 HTTP 更新源,itsnot.fun nginx 托管)。
const UpdateJSONURL = "https://itsnot.fun/rain-qa-func/latest.json"

// UpdateInfo 服务端 latest.json 的版本元数据。
type UpdateInfo struct {
	VersionCode  int    `json:"versionCode"`  // 整数版本号(递增,对比用,源自 build.gradle)
	VersionName  string `json:"versionName"`  // 展示名(如 1.0.1)
	ApkUrl       string `json:"apkUrl"`       // APK 下载 URL
	Sha256       string `json:"sha256"`       // APK SHA256(空则跳过校验)
	ReleaseNotes string `json:"releaseNotes"` // 更新说明(前端展示)
}

// UpdateService Android APK 自动更新服务。
type UpdateService struct {
	app *application.App
}

// New 创建更新服务。需调用 InitWithApp 注入 app 后才能 Event 推进度。
func New() *UpdateService { return &UpdateService{} }

// InitWithApp 注入 app 实例(供 Event.Emit 推送下载进度)。
func (s *UpdateService) InitWithApp(app *application.App) { s.app = app }

// CheckUpdate 查询服务端版本,对比 currentVersionCode。
// currentVersionCode 由前端 getAppVersionCode()(Java 桥读 build.gradle)传入。
// 返回:有新版 → *UpdateInfo;无新版或当前已最新 → nil。
// @frontend
func (s *UpdateService) CheckUpdate(currentVersionCode int) (*UpdateInfo, error) {
	log.Printf("[update] CheckUpdate 进入, currentVersionCode=%d", currentVersionCode)
	resp, err := http.Get(UpdateJSONURL)
	if err != nil {
		log.Printf("[update] CheckUpdate http.Get 失败: %v", err)
		return nil, fmt.Errorf("检查更新失败(网络): %v", err)
	}
	defer resp.Body.Close()
	log.Printf("[update] CheckUpdate http.Get 成功, status=%d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务端返回状态 %d", resp.StatusCode)
	}
	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		log.Printf("[update] CheckUpdate 解析失败: %v", err)
		return nil, fmt.Errorf("解析版本信息失败: %v", err)
	}
	log.Printf("[update] CheckUpdate current=%d server versionCode=%d versionName=%s → 有新版=%v",
		currentVersionCode, info.VersionCode, info.VersionName, info.VersionCode > currentVersionCode)
	if info.VersionCode > currentVersionCode {
		return &info, nil
	}
	return nil, nil // 无新版
}

// DownloadApk 下载 info.ApkUrl 到 app 私有 files/updates/,边下边算 SHA256,
// 通过 updateProgress 事件推送百分比进度,下载完校验 SHA256(若 info.Sha256 非空)。
// 返回本地 APK 绝对路径(供前端调 Java installApk)。
// @frontend
func (s *UpdateService) DownloadApk(info UpdateInfo) (string, error) {
	if info.ApkUrl == "" {
		return "", fmt.Errorf("apkUrl 为空")
	}
	// files/updates/ 目录(Android: /data/data/<pkg>/files/updates/;桌面: CWD/updates/)
	updatesDir := filepath.Join(filepath.Dir(appconfig.FilePath()), "updates")
	if err := os.MkdirAll(updatesDir, 0755); err != nil {
		return "", fmt.Errorf("创建更新目录失败: %v", err)
	}
	apkPath := filepath.Join(updatesDir, fmt.Sprintf("rain-qa-func-%d.apk", info.VersionCode))
	partPath := apkPath + ".part"

	// 系统干净:清理旧 versionCode 的 APK + .part(仅留当前版本,卸载随 files/ 清)
	cleanOldUpdates(updatesDir, info.VersionCode)
	s.emitProgress(0) // 下载启动信号(前端显示 0%,避免"无反应")
	log.Printf("[update] DownloadApk 开始: %s", info.ApkUrl)

	// HEAD 获取总大小(判断续传点 + 卡死恢复)
	total, _, err := headApk(info.ApkUrl)
	if err != nil {
		return "", fmt.Errorf("查询文件信息失败: %v", err)
	}
	if total <= 0 {
		return "", fmt.Errorf("无法获取文件大小(Content-Length=%d)", total)
	}

	// 断点续传起点:查 .part 已下大小(超过总大小则旧文件,重头)
	offset := int64(0)
	if fi, e := os.Stat(partPath); e == nil {
		offset = fi.Size()
		if offset > total {
			os.Remove(partPath)
			offset = 0
		}
		if offset > 0 {
			log.Printf("[update] 断点续传: 已下 %d/%d (%.1f%%)", offset, total, float64(offset)*100/float64(total))
		}
	}

	// 分块下载(2MB/块)+ 每块超时(30s)+ 失败重试(5次退避)
	// 解原全量 http.Get 三大问题:无超时(卡死)/无续传(中断重头)/无重试(弱网失败)
	client := &http.Client{}
	for offset < total {
		end := min(offset+apkDownloadChunkSize-1, total-1)
		var chunkErr error
		for retry := 0; retry <= apkDownloadMaxRetry; retry++ {
			chunkErr = s.downloadChunk(client, info.ApkUrl, partPath, offset, end)
			if chunkErr == nil {
				break
			}
			if retry == apkDownloadMaxRetry {
				return "", fmt.Errorf("下载失败(offset=%d,重试 %d 次): %v", offset, retry, chunkErr)
			}
			backoff := time.Duration(retry+1) * time.Second
			log.Printf("[update] 块下载重试 %d (offset=%d,%v),%v 后重试", retry, offset, chunkErr, backoff)
			time.Sleep(backoff)
		}
		offset = end + 1
		s.emitProgress(int(offset * 100 / total))
	}
	s.emitProgress(100)

	// SHA256 校验(整体重算 .part;防劫持/防损坏)
	if info.Sha256 != "" {
		got, e := sha256File(partPath)
		if e != nil {
			return "", fmt.Errorf("SHA256 计算失败: %v", e)
		}
		if got != info.Sha256 {
			os.Remove(partPath)
			return "", fmt.Errorf("SHA256 校验失败: 期望 %s 实际 %s", info.Sha256, got)
		}
	}

	// .part → .apk
	if err := os.Rename(partPath, apkPath); err != nil {
		return "", fmt.Errorf("重命名失败: %v", err)
	}
	log.Printf("[update] 下载完成: %s (%d bytes)", apkPath, total)
	return apkPath, nil
}

// 下载优化常量(分块+超时+重试,解弱网/卡死/中断)。
const (
	// apkDownloadChunkSize 分块下载块大小(2MB)。弱网下小块更易重试成功,平衡请求数与容错。
	apkDownloadChunkSize = 2 * 1024 * 1024
	// apkDownloadChunkTimeout 单块下载超时(30s)。无数据 30s 判卡死,触发重试(断点续传)。
	apkDownloadChunkTimeout = 30 * time.Second
	// apkDownloadMaxRetry 单块最大重试次数(失败后线性退避重试,超过则整体失败)。
	apkDownloadMaxRetry = 5
)

// downloadChunk 下载单个块 [offset,end] 写入 partPath 对应位置。带 context 超时(防卡死)。
func (s *UpdateService) downloadChunk(client *http.Client, url, partPath string, offset, end int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), apkDownloadChunkTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, end))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 续传期望 206 Partial Content;首块(offset=0,end=total-1)服务器可能返回 200
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("块下载返回状态 %d", resp.StatusCode)
	}

	// 写 .part 的 offset 位置(分块顺序写,offset 单调增,Seek 覆盖写无残留)
	out, err := os.OpenFile(partPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := out.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入中断(offset=%d): %v", offset, err)
	}
	// 块不完整(服务器返回少于请求范围)→ 视为失败,外层重试该块
	if written < end-offset+1 {
		return fmt.Errorf("块不完整(期望 %d 实际 %d)", end-offset+1, written)
	}
	return nil
}

// headApk HEAD 请求获取 Content-Length(判断总大小 + 续传点校验)。带 15s 超时(防卡死)。
func headApk(url string) (total int64, etag string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, "", fmt.Errorf("HEAD 返回状态 %d", resp.StatusCode)
	}
	log.Printf("[update] headApk: total=%d etag=%s", resp.ContentLength, resp.Header.Get("Etag"))
	return resp.ContentLength, resp.Header.Get("Etag"), nil
}

// cleanOldUpdates 清理 updates/ 目录下非当前 versionCode 的 APK + .part(系统干净,避免长期累积)。
func cleanOldUpdates(dir string, keepVersionCode int) {
	keep := fmt.Sprintf("rain-qa-func-%d.apk", keepVersionCode)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		name := e.Name()
		if name == keep || name == keep+".part" {
			continue
		}
		// 删其他版本 APK + 残留 .part
		if filepath.Ext(name) == ".apk" || strings.HasSuffix(name, ".part") {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

// CleanInstalledApks 清理 updates/ 下所有 .apk(已安装版本),保留 .part(断点续传未完成)。
// 系统干净:应用启动时调。上次下载的 APK 若已安装(能启动=装成功),可清;
// 未完成的 .part 保留以便下次续传。方案 🅱️(安装失败可重试,启动时清已装的)。
func CleanInstalledApks() {
	updatesDir := filepath.Join(filepath.Dir(appconfig.FilePath()), "updates")
	entries, err := os.ReadDir(updatesDir)
	if err != nil {
		return // 目录不存在(首次启动)无碍
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".apk" {
			os.Remove(filepath.Join(updatesDir, e.Name()))
		}
	}
}

// sha256File 计算文件 SHA256(断点续传完成后整体重算,简单可靠,避免 hash 状态序列化)。
func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// emitProgress 推送下载进度到前端(updateProgress 事件,前端监听显示进度条)。
// 用 map[string]any 而非结构体(Wails v3 Event.Emit 序列化自定义结构体有问题)。
func (s *UpdateService) emitProgress(pct int) {
	if s.app != nil {
		s.app.Event.Emit("updateProgress", map[string]any{"percent": pct})
	}
}
