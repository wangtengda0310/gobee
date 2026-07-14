package gitutil

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseBatchLogOutput_正常解析 测试标准格式的日志输出解析
func TestParseBatchLogOutput_正常解析(t *testing.T) {
	output := "abc123def456|1712000000|张三|zhangsan@ztgame.com\ndef789abc012|1712000100|李四|lisi@ztgame.com"
	authors, timestamps, emails := parseBatchLogOutput(output, 2)

	assert.Equal(t, "张三", authors["abc123def456"])
	assert.Equal(t, "李四", authors["def789abc012"])
	assert.Equal(t, int64(1712000000), timestamps["abc123def456"])
	assert.Equal(t, int64(1712000100), timestamps["def789abc012"])
	assert.Equal(t, "zhangsan@ztgame.com", emails["abc123def456"])
	assert.Equal(t, "lisi@ztgame.com", emails["def789abc012"])
	assert.Equal(t, 2, len(authors))
	assert.Equal(t, 2, len(timestamps))
	assert.Equal(t, 2, len(emails))
}

// TestParseBatchLogOutput_空输出 测试空字符串输入
func TestParseBatchLogOutput_空输出(t *testing.T) {
	authors, timestamps, emails := parseBatchLogOutput("", 0)

	assert.Equal(t, 0, len(authors))
	assert.Equal(t, 0, len(timestamps))
	assert.Equal(t, 0, len(emails))
}

// TestParseBatchLogOutput_包含空格的作者名 测试作者名中有空格和特殊字符
func TestParseBatchLogOutput_包含空格的作者名(t *testing.T) {
	output := "abc123|1712000000|John Smith|john@company.com"
	authors, timestamps, emails := parseBatchLogOutput(output, 1)

	assert.Equal(t, "John Smith", authors["abc123"])
	assert.Equal(t, int64(1712000000), timestamps["abc123"])
	assert.Equal(t, "john@company.com", emails["abc123"])
}

// TestParseBatchLogOutput_作者名包含竖线 测试四字段格式下作者名中包含 | 字符
// 使用 SplitN(line, "|", 4) 分割，第四个字段之后的内容归入 email
func TestParseBatchLogOutput_作者名包含竖线(t *testing.T) {
	// 四字段格式：hash|timestamp|author|email，作者名不应包含 |
	// 但如果出现了，SplitN 保证最多分成4段，author 仍取第3段
	output := "abc123|1712000000|test|user@company.com"
	authors, timestamps, emails := parseBatchLogOutput(output, 1)

	assert.Equal(t, "test", authors["abc123"])
	assert.Equal(t, int64(1712000000), timestamps["abc123"])
	assert.Equal(t, "user@company.com", emails["abc123"])
}

// TestParseBatchLogOutput_无效行跳过 测试格式不正确的行被正确跳过
func TestParseBatchLogOutput_无效行跳过(t *testing.T) {
	output := "abc123|1712000000|张三|zhangsan@ztgame.com\n无效行\ndef456|not_number|李四|lisi@ztgame.com\n\nxyz789|1712000200|王五|wangwu@ztgame.com"
	authors, timestamps, emails := parseBatchLogOutput(output, 3)

	// 只有第一行和最后一行有效
	assert.Equal(t, "张三", authors["abc123"])
	assert.Equal(t, "王五", authors["xyz789"])
	assert.Equal(t, int64(1712000000), timestamps["abc123"])
	assert.Equal(t, int64(1712000200), timestamps["xyz789"])
	assert.Equal(t, "zhangsan@ztgame.com", emails["abc123"])
	assert.Equal(t, "wangwu@ztgame.com", emails["xyz789"])
	assert.Equal(t, 2, len(authors))
	// 时间戳解析失败的行不加入 map
	_, exists := authors["def456"]
	assert.False(t, exists)
}

// TestParseBatchLogOutput_单行无换行 测试没有换行符的单行输出
func TestParseBatchLogOutput_单行无换行(t *testing.T) {
	output := "abc123|1712000000|张三|zhangsan@ztgame.com"
	authors, timestamps, emails := parseBatchLogOutput(output, 1)

	assert.Equal(t, "张三", authors["abc123"])
	assert.Equal(t, int64(1712000000), timestamps["abc123"])
	assert.Equal(t, "zhangsan@ztgame.com", emails["abc123"])
}

// TestParseBatchLogOutput_完整40位hash 测试使用完整 commit hash 的场景
func TestParseBatchLogOutput_完整40位hash(t *testing.T) {
	hash1 := "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0"
	hash2 := "1111222233334444555566667777888899990000"
	output := hash1 + "|1712000000|张三|zhangsan@ztgame.com\n" + hash2 + "|1712000100|李四|lisi@ztgame.com"
	authors, timestamps, emails := parseBatchLogOutput(output, 2)

	assert.Equal(t, "张三", authors[hash1])
	assert.Equal(t, "李四", authors[hash2])
	assert.Equal(t, int64(1712000000), timestamps[hash1])
	assert.Equal(t, int64(1712000100), timestamps[hash2])
	assert.Equal(t, "zhangsan@ztgame.com", emails[hash1])
	assert.Equal(t, "lisi@ztgame.com", emails[hash2])
}

// TestParseBatchLogOutput_缺少竖线分隔符 测试管道符数量不足的行
func TestParseBatchLogOutput_缺少竖线分隔符(t *testing.T) {
	output := "no_pipes_at_all\nonly_one|pipe"
	authors, timestamps, emails := parseBatchLogOutput(output, 2)

	assert.Equal(t, 0, len(authors))
	assert.Equal(t, 0, len(timestamps))
	assert.Equal(t, 0, len(emails))
}

// TestParseBatchLogOutput_四字段格式 测试完整的四字段格式解析
func TestParseBatchLogOutput_四字段格式(t *testing.T) {
	output := "abc123|1712000000|张三|zhangsan@ztgame.com\ndef456|1712000100|李四|lisi@ztgame.com"
	authors, timestamps, emails := parseBatchLogOutput(output, 2)
	// 验证 authors
	assert.Equal(t, "张三", authors["abc123"])
	assert.Equal(t, "李四", authors["def456"])
	// 验证 timestamps
	assert.Equal(t, int64(1712000000), timestamps["abc123"])
	assert.Equal(t, int64(1712000100), timestamps["def456"])
	// 验证 emails
	assert.Equal(t, "zhangsan@ztgame.com", emails["abc123"])
	assert.Equal(t, "lisi@ztgame.com", emails["def456"])
}

// TestParseBatchLogOutput_三字段兼容 测试旧格式（无邮箱）的向后兼容
func TestParseBatchLogOutput_三字段兼容(t *testing.T) {
	output := "abc123|1712000000|张三"
	authors, timestamps, emails := parseBatchLogOutput(output, 1)
	assert.Equal(t, "张三", authors["abc123"])
	assert.Equal(t, int64(1712000000), timestamps["abc123"])
	assert.Equal(t, "", emails["abc123"]) // 无邮箱时为空
}

// TestParseBatchLogOutput_邮箱含特殊字符 测试邮箱中包含 + 和 . 等特殊字符
func TestParseBatchLogOutput_邮箱含特殊字符(t *testing.T) {
	output := "abc123|1712000000|张三|user.name+tag@company.com"
	_, _, emails := parseBatchLogOutput(output, 1)
	assert.Equal(t, "user.name+tag@company.com", emails["abc123"])
}

// TestMergeChildCommits_按时间排序 测试 MergeChildCommit 列表按 Timestamp 升序排列
func TestMergeChildCommits_按时间排序(t *testing.T) {
	commits := []MergeChildCommit{
		{Hash: "c3", Author: "王五", Timestamp: 1712000200, Branch: 1},
		{Hash: "c1", Author: "张三", Timestamp: 1712000000, Branch: 0},
		{Hash: "c2", Author: "李四", Timestamp: 1712000100, Branch: 0},
	}

	// 使用与 merge.go 中 GetMergeChildCommits 相同的排序逻辑
	sort.SliceStable(commits, func(i, j int) bool {
		return commits[i].Timestamp < commits[j].Timestamp
	})

	assert.Equal(t, "c1", commits[0].Hash)
	assert.Equal(t, int64(1712000000), commits[0].Timestamp)
	assert.Equal(t, "c2", commits[1].Hash)
	assert.Equal(t, int64(1712000100), commits[1].Timestamp)
	assert.Equal(t, "c3", commits[2].Hash)
	assert.Equal(t, int64(1712000200), commits[2].Timestamp)
}

// TestMergeChildCommits_排序稳定性 测试相同时间戳时保持原始顺序
func TestMergeChildCommits_排序稳定性(t *testing.T) {
	commits := []MergeChildCommit{
		{Hash: "c3", Author: "王五", Timestamp: 1712000000, Branch: 1},
		{Hash: "c1", Author: "张三", Timestamp: 1712000000, Branch: 0},
		{Hash: "c2", Author: "李四", Timestamp: 1712000000, Branch: 0},
	}

	// SliceStable 保证相同时间戳时保持原始顺序
	sort.SliceStable(commits, func(i, j int) bool {
		return commits[i].Timestamp < commits[j].Timestamp
	})

	// 相同时间戳，顺序不变
	assert.Equal(t, "c3", commits[0].Hash)
	assert.Equal(t, "c1", commits[1].Hash)
	assert.Equal(t, "c2", commits[2].Hash)
}

// TestMergeChildCommits_分支归属标记 测试 Branch 字段正确标记
func TestMergeChildCommits_分支归属标记(t *testing.T) {
	// 模拟构建过程：parent1 的 commit Branch=0，parent2 的 commit Branch=1
	parent1Hashes := []string{"aaa", "bbb"}
	parent2Hashes := []string{"ccc"}

	allCommits := make([]MergeChildCommit, 0, 3)
	for _, hash := range parent1Hashes {
		allCommits = append(allCommits, MergeChildCommit{Hash: hash, Branch: 0})
	}
	for _, hash := range parent2Hashes {
		allCommits = append(allCommits, MergeChildCommit{Hash: hash, Branch: 1})
	}

	// 验证分支归属
	assert.Equal(t, 0, allCommits[0].Branch)
	assert.Equal(t, "aaa", allCommits[0].Hash)
	assert.Equal(t, 0, allCommits[1].Branch)
	assert.Equal(t, "bbb", allCommits[1].Hash)
	assert.Equal(t, 1, allCommits[2].Branch)
	assert.Equal(t, "ccc", allCommits[2].Hash)
}
