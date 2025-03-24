package api

import (
	"archive/zip"
	"fmt"
	"github.com/wangtengda/gobee/lvan/exporter/internal"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wangtengda/gobee/lvan/exporter/pkg"
	"github.com/wangtengda/gobee/lvan/exporter/pkg/logger"
)

// HandleBackupRequest 处理备份请求，将指定子目录打包为zip文件供下载
func HandleBackupRequest(w http.ResponseWriter, r *http.Request) {
	// 解析路径中的子目录名
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid request path, usage: /backup/subdir", http.StatusBadRequest)
		return
	}

	// 获取子目录名
	subDir := pathParts[2]
	if subDir == "" {
		http.Error(w, "Subdirectory name is required", http.StatusBadRequest)
		return
	}

	// 获取工作目录
	workDir := internal.WorkDir

	// 构建完整的子目录路径
	subDirPath := filepath.Join(workDir, subDir)

	// 检查子目录是否存在
	if _, err := os.Stat(subDirPath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Subdirectory '%s' does not exist", subDir), http.StatusNotFound)
		return
	}

	// 创建临时文件用于存储zip内容
	tmpFile, err := os.CreateTemp("", "backup-*.zip")
	if err != nil {
		logger.Error("Failed to create temporary file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name()) // 确保临时文件在函数结束时被删除
	defer tmpFile.Close()

	// 创建zip writer
	zipWriter := zip.NewWriter(tmpFile)
	defer zipWriter.Close()

	// 遍历子目录并添加文件到zip
	err = filepath.Walk(subDirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对路径
		relPath, err := filepath.Rel(subDirPath, path)
		if err != nil {
			return err
		}

		// 跳过根目录
		if relPath == "." {
			return nil
		}

		// 创建zip文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// 设置文件名为相对路径
		header.Name = relPath

		// 设置压缩方法
		if !info.IsDir() {
			header.Method = zip.Deflate
		} else {
			// 确保目录路径以/结尾
			header.Name += "/"
		}

		// 创建文件或目录条目
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// 如果是文件，写入文件内容
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logger.Error("Failed to create zip file: %v", err)
		http.Error(w, "Failed to create backup", http.StatusInternalServerError)
		return
	}

	// 关闭zip writer以确保所有数据都被写入
	zipWriter.Close()

	// 将文件指针移到文件开头
	tmpFile.Seek(0, 0)

	// 设置响应头
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s.zip", subDir, time.Now().Format("20060102150405")))

	// 将zip文件内容写入响应
	_, err = io.Copy(w, tmpFile)
	if err != nil {
		logger.Error("Failed to send zip file: %v", err)
		http.Error(w, "Failed to send backup", http.StatusInternalServerError)
		return
	}

	logger.Info("Successfully created backup for subdirectory: %s", subDir)
}

// getWorkDir 获取工作目录
func getWorkDir() string {
	// 这里需要访问main包中的WorkDir变量
	// 由于变量在不同包中，我们需要通过环境变量或其他方式获取
	// 这里假设WorkDir已经通过某种方式设置在pkg包中
	return filepath.Dir(filepath.Dir(pkg.TasksDir))
}
