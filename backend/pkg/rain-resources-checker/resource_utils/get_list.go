package resource_utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/resource_utils/gitlab"
)

type FileInfo struct {
	Name string
	Path string
}

func GetFilesLocal(dir, ext string) (filesInfo []FileInfo, err error) {
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	files, err := filepath.Glob(fmt.Sprintf("%s/*%s", dir, ext))
	if err != nil {
		return nil, err
	}

	filesInfo = make([]FileInfo, len(files))
	for i, file := range files {
		filesInfo[i].Path = file
		filesInfo[i].Name = filepath.Base(file)
	}

	return filesInfo, nil
}

// GetFilesGitlab
// branch "v0.0.8-pre-release"
// project "Xcards/client"
// projectPath "Master/Card/Audio/Voice"
// accessToken "jMD7RXcbMLosvHoPZTbS"
// ext 后缀过滤
func GetFilesGitlab(branch, project, projectPath, accessToken, ext string) (filesInfo []gitlab.FileInfo, err error) {
	// 初始化检查器
	checker := gitlab.NewGitLabResourceChecker(
		"https://git.devcloud.ztgame.com",
		project,
		accessToken,
	)

	fmt.Printf("正在获取 %s 分支的Voice目录...\n", branch)

	filesInfo, err = checker.ListAllVoiceFilesWithLinkHeader(branch, projectPath, ext)
	if err != nil {
		fmt.Printf("获取文件列表失败: %v\n", err)
		os.Exit(1)
	}

	return filesInfo, nil
}
