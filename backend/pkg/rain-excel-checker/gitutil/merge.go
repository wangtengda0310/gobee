// Package gitutil 提供 git 仓库操作的工具函数
package gitutil

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// MergeChildCommit merge 场景中一个子 commit 的信息
type MergeChildCommit struct {
	Hash      string // commit hash（完整40位）
	Author    string // 提交作者
	Email     string // 作者邮箱
	Timestamp int64  // 提交时间戳（committer date，Unix秒）
	Branch    int    // 来自哪个分支（0=主分支parent1, 1=被合并分支parent2）
}

// MergeChildCommits merge commit 的子提交信息
type MergeChildCommits struct {
	SortedCommits  []MergeChildCommit // 按时间正序的所有子commit（最早的在前）
	Parent1Commits []string           // 主分支的commit hash列表（原始拓扑序）
	Parent2Commits []string           // 被合并分支的commit hash列表（原始拓扑序）
}

// GetMergeChildCommits 获取 merge commit 两个分支的子 commit 信息
// 返回按时间排序的统一列表，以及按分支分组的原始列表（供通知使用）
//
// 处理流程：
//  1. 获取 merge commit 的两个父提交
//  2. 计算两个分支的 merge base
//  3. 分别获取两个分支从 merge base 到各自父提交之间的 commit
//  4. 批量获取所有 commit 的 author 和 timestamp
//  5. 构建统一列表并按时间排序
func GetMergeChildCommits(repoPath, mergeHash string) (*MergeChildCommits, error) {
	// 1. 获取两个父提交
	parents, err := GetMergeParentCommits(repoPath, mergeHash)
	if err != nil {
		return nil, fmt.Errorf("获取 merge commit 父提交失败: %w", err)
	}

	// 2. 计算 merge base
	mergeBase, err := GetMergeBase(repoPath, parents[0], parents[1])
	if err != nil {
		return nil, fmt.Errorf("获取 merge base 失败: %w", err)
	}

	// 3. 分别获取两个分支的 commit（按拓扑序，即时间正序）
	parent1Commits, err := GetCommitsBetween(repoPath, mergeBase, parents[0])
	if err != nil {
		return nil, fmt.Errorf("获取主分支 commit 列表失败: %w", err)
	}

	parent2Commits, err := GetCommitsBetween(repoPath, mergeBase, parents[1])
	if err != nil {
		return nil, fmt.Errorf("获取被合并分支 commit 列表失败: %w", err)
	}

	// 4. 批量获取所有 commit 的 author 和 timestamp
	allHashes := make([]string, 0, len(parent1Commits)+len(parent2Commits))
	allHashes = append(allHashes, parent1Commits...)
	allHashes = append(allHashes, parent2Commits...)

	authors, timestamps, emails, err := batchGetCommitInfo(repoPath, allHashes)
	if err != nil {
		return nil, fmt.Errorf("批量获取 commit 信息失败: %w", err)
	}

	// 5. 构建 MergeChildCommit 列表并标记分支归属
	var sortedCommits []MergeChildCommit
	for _, hash := range parent1Commits {
		sortedCommits = append(sortedCommits, MergeChildCommit{
			Hash:      hash,
			Author:    authors[hash],
			Email:     emails[hash],
			Timestamp: timestamps[hash],
			Branch:    0,
		})
	}
	for _, hash := range parent2Commits {
		sortedCommits = append(sortedCommits, MergeChildCommit{
			Hash:      hash,
			Author:    authors[hash],
			Email:     emails[hash],
			Timestamp: timestamps[hash],
			Branch:    1,
		})
	}

	// 6. 按时间排序（时间早的在前）
	sort.SliceStable(sortedCommits, func(i, j int) bool {
		return sortedCommits[i].Timestamp < sortedCommits[j].Timestamp
	})

	return &MergeChildCommits{
		SortedCommits:  sortedCommits,
		Parent1Commits: parent1Commits,
		Parent2Commits: parent2Commits,
	}, nil
}

// batchGetCommitInfo 批量获取多个 commit 的作者、时间戳和邮箱
// 使用一次 git log 命令获取所有信息，避免逐个调用
// 返回三个 map：authors（hash -> 作者名）、timestamps（hash -> Unix时间戳）、emails（hash -> 邮箱）
func batchGetCommitInfo(repoPath string, commits []string) (map[string]string, map[string]int64, map[string]string, error) {
	if len(commits) == 0 {
		return make(map[string]string), make(map[string]int64), make(map[string]string), nil
	}

	// 构造 git log 参数：输出 hash|timestamp|author|email
	// 使用 %H（完整hash）、%ct（committer date Unix时间戳）、%an（作者名）、%ae（作者邮箱）
	args := []string{
		"log",
		"--pretty=format:%H|%ct|%an|%ae",
		"--no-walk",
	}
	args = append(args, commits...)

	output, err := runGit(repoPath, args...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("执行 git log 批量查询失败: %w", err)
	}

	authors, timestamps, emails := parseBatchLogOutput(output, len(commits))
	return authors, timestamps, emails, nil
}

// parseBatchLogOutput 解析 git log --pretty=format:"%H|%ct|%an|%ae" 的输出
// 提取为独立函数便于单元测试覆盖格式解析的正确性
// 参数 expectedCount 用于预分配 map 容量，传入 0 即可
// 向后兼容：如果一行只有三个字段（hash|timestamp|author），email 为空字符串
func parseBatchLogOutput(output string, expectedCount int) (map[string]string, map[string]int64, map[string]string) {
	authors := make(map[string]string, expectedCount)
	timestamps := make(map[string]int64, expectedCount)
	emails := make(map[string]string, expectedCount)

	if output == "" {
		return authors, timestamps, emails
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 格式：hash|timestamp|author|email
		// email 可能包含特殊字符（如 + 和 .），但不包含 |，所以按 | 分割
		// 向后兼容：如果只有三个字段，email 为空
		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 3 {
			continue
		}

		hash := parts[0]
		timestampStr := parts[1]
		author := parts[2]
		email := ""
		if len(parts) == 4 {
			email = parts[3]
		}

		ts, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			// 时间戳解析失败，跳过该行
			continue
		}

		authors[hash] = author
		timestamps[hash] = ts
		emails[hash] = email
	}

	return authors, timestamps, emails
}
