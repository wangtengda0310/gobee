// Package gitutil 提供 git 仓库操作的工具函数
package gitutil

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
)

// runGit 在指定目录执行 git 命令并返回输出
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ==================== 基础查询 ====================

// GetRepoRoot 获取 git 仓库根目录
func GetRepoRoot(path string) (string, error) {
	return runGit(path, "rev-parse", "--show-toplevel")
}

// decodeGitPath 解码 git 输出的文件路径
// 处理两个问题：
//  1. git core.quotepath=true 时中文路径会变成八进制转义（如 \350\205...）
//  2. quotepath=true 时整个路径会被双引号包裹（如 "excel/XXX.xlsx"）
//
// 如果只做八进制解码不去引号，路径会变成 "excel/腾讯.xlsx"（带引号），
// 导致 strings.HasSuffix(".xlsx") 匹配失败
func decodeGitPath(path string) string {
	decoded := feishu.DecodeOctalEscape(path)
	// 去除 quotepath=true 时 git 添加的首尾双引号
	decoded = strings.Trim(decoded, "\"")
	return decoded
}

// ListXlsxAtCommit 获取指定 commit 中所有 xlsx 文件的相对路径列表
// 用于 merge 遍历场景，补充 sheetMap 中缺失的关联表
// 注意：Git 在 core.quotepath=true 时会对中文文件名做八进制转义并加引号包裹，
// 本函数通过 decodeGitPath 自动处理八进制解码和去引号
func ListXlsxAtCommit(repoPath, commitHash string) ([]string, error) {
	output, err := runGit(repoPath, "ls-tree", "-r", "--name-only", commitHash)
	if err != nil {
		return nil, err
	}
	var xlsxFiles []string
	for line := range strings.SplitSeq(output, "\n") {
		decoded := decodeGitPath(line)
		if decoded != "" && strings.HasSuffix(strings.ToLower(decoded), ".xlsx") {
			xlsxFiles = append(xlsxFiles, decoded)
		}
	}
	return xlsxFiles, nil
}

// GetHeadHash 获取当前 HEAD commit hash
func GetHeadHash(repoPath string) (string, error) {
	return runGit(repoPath, "rev-parse", "HEAD")
}

// GetCurrentBranch 获取当前分支名
func GetCurrentBranch(repoPath string) (string, error) {
	return runGit(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
}

// ==================== Commit 信息查询 ====================

// GetCommitAuthor 获取指定 commit 的作者
func GetCommitAuthor(repoPath, commitHash string) (string, error) {
	return runGit(repoPath, "log", "-1", "--pretty=%an", commitHash)
}

// GetCommitDate 获取指定 commit 的提交时间
// commitHash 为空时获取 HEAD 的提交时间
func GetCommitDate(repoPath, commitHash string) (time.Time, error) {
	if commitHash == "" {
		commitHash = "HEAD"
	}
	output, err := runGit(repoPath, "log", "-1", "--pretty=%ci", commitHash)
	if err != nil {
		return time.Time{}, fmt.Errorf("获取 commit %s 提交时间失败: %w", commitHash, err)
	}
	// git %ci 输出格式: "2026-04-13 16:05:45 +0800"
	t, err := time.Parse("2006-01-02 15:04:05 -0700", output)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析 commit 时间失败: %w", err)
	}
	return t, nil
}

// GetCommitMessage 获取指定 commit 的提交消息
func GetCommitMessage(repoPath, commitHash string) (string, error) {
	return runGit(repoPath, "log", "-1", "--pretty=%s", commitHash)
}

// GetParentCommit 获取指定 commit 的第一个父 commit hash
// 用于获取某个 commit 的前一版本，作为 diff 对比基准
func GetParentCommit(repoPath, commitHash string) (string, error) {
	return runGit(repoPath, "rev-parse", commitHash+"^")
}

// GetCommitAuthorEmail 获取指定 commit 的作者邮箱
func GetCommitAuthorEmail(repoPath, commitHash string) (string, error) {
	return runGit(repoPath, "log", "-1", "--pretty=%ae", commitHash)
}

// GetCommitInfo 获取指定 commit 的完整信息（hash、message、author）
// commitHash 为空时获取 HEAD 的信息
func GetCommitInfo(repoPath, commitHash string) (hash, message, author string, err error) {
	if commitHash == "" {
		commitHash = "HEAD"
	}

	hash, err = GetHeadHash(repoPath)
	if err != nil {
		return "", "", "", fmt.Errorf("获取 commit hash 失败: %w", err)
	}

	// commitHash 可能是 "HEAD"、"HEAD~1" 等符号引用，但 GetCommitMessage/GetCommitAuthor
	// 底层 git log 命令都支持符号引用，无需额外转换
	message, err = GetCommitMessage(repoPath, commitHash)
	if err != nil {
		return "", "", "", fmt.Errorf("获取 commit message 失败: %w", err)
	}

	author, err = GetCommitAuthor(repoPath, commitHash)
	if err != nil {
		return "", "", "", fmt.Errorf("获取作者失败: %w", err)
	}

	return hash, message, author, nil
}

// ==================== Merge Commit 检测 ====================

// IsMergeCommit 判断指定 commit 是否为 merge commit（父提交数 > 1）
func IsMergeCommit(repoPath, commitHash string) (bool, error) {
	// git cat-file -p 获取 commit 对象内容，解析 parent 行数量
	output, err := runGit(repoPath, "cat-file", "-p", commitHash)
	if err != nil {
		return false, fmt.Errorf("获取 commit 信息失败: %w", err)
	}

	parentCount := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "parent ") {
			parentCount++
		}
		// parent 行在 tree 行之后，一旦遇到空行或其他行就停止计数
		if line == "" {
			break
		}
	}

	return parentCount > 1, nil
}

// GetMergeParentCommits 获取 merge commit 的父提交列表
// 返回的列表中第一个是主分支父提交，第二个是被合并的分支父提交
func GetMergeParentCommits(repoPath, commitHash string) ([]string, error) {
	output, err := runGit(repoPath, "cat-file", "-p", commitHash)
	if err != nil {
		return nil, fmt.Errorf("获取 commit 信息失败: %w", err)
	}

	var parents []string
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "parent ") {
			parents = append(parents, strings.TrimPrefix(line, "parent "))
		}
		if line == "" {
			break
		}
	}

	if len(parents) < 2 {
		return nil, fmt.Errorf("commit %s 不是 merge commit（只有 %d 个父提交）", commitHash, len(parents))
	}

	return parents, nil
}

// GetMergeBase 获取两个 commit 的最近公共祖先（merge base）
func GetMergeBase(repoPath, commit1, commit2 string) (string, error) {
	return runGit(repoPath, "merge-base", commit1, commit2)
}

// GetCommitsBetween 获取两个 commit 之间的所有 commit hash（按时间正序）
// 返回 baseCommit 之后到 headCommit 之间的所有 commit（不包含 baseCommit，包含 headCommit）
func GetCommitsBetween(repoPath, baseCommit, headCommit string) ([]string, error) {
	output, err := runGit(repoPath, "rev-list", "--reverse", baseCommit+".."+headCommit)
	if err != nil {
		return nil, fmt.Errorf("获取 commit 列表失败: %w", err)
	}

	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// ==================== 变更文件查询 ====================

// GetCommitDiffFiles 获取指定 commit 的变更文件列表（相对于仓库根目录的路径）
// 对于 merge commit 会返回与第一个父提交的差异
// 注意：Git 在 core.quotepath=true（Docker Alpine 默认值）时会输出带引号的八进制转义路径
// （如 "excel/Survey_\350\205...xlsx"），本函数通过 decodeGitPath 自动解码为 UTF-8 并去除引号
func GetCommitDiffFiles(repoPath, commitHash string) ([]string, error) {
	output, err := runGit(repoPath, "diff-tree", "--no-commit-id", "-r", "--name-only", commitHash)
	if err != nil {
		return nil, fmt.Errorf("获取 commit 变更文件失败: %w", err)
	}

	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(output, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		decoded := decodeGitPath(line)
		result = append(result, decoded)
	}

	return result, nil
}

// GetFileAtCommit 获取指定 commit 的文件内容
func GetFileAtCommit(repoPath, commitHash, filePath string) ([]byte, error) {
	// 获取 git 仓库根目录
	gitRoot, err := GetRepoRoot(repoPath)
	if err != nil {
		return nil, fmt.Errorf("获取 git 仓库失败: %w", err)
	}

	// 计算文件相对于仓库根目录的相对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取绝对路径失败: %w", err)
	}

	relPath, err := filepath.Rel(gitRoot, absPath)
	if err != nil {
		return nil, fmt.Errorf("计算相对路径失败: %w", err)
	}

	// Git 命令需要正斜杠路径
	gitPath := strings.ReplaceAll(relPath, "\\", "/")

	// git show 不通过 runGit 调用，因为需要返回 []byte
	cmd := exec.Command("git", "show", commitHash+":"+gitPath)
	cmd.Dir = gitRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取文件内容失败: %w", err)
	}

	return output, nil
}
