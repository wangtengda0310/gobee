package gitutil

import (
	"strings"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/common/feishu"
	"github.com/stretchr/testify/assert"
)

// TestDecodeOctalEscape 测试 Git 八进制转义序列解码
// 验证 Git 在 core.quotepath=true 时输出的中文文件名能被正确解码为 UTF-8
func TestDecodeOctalEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "中文文件名-武将归隐表",
			input:    "excel/HeroRetreat_\\346\\255\\246\\345\\260\\206\\345\\275\\222\\351\\232\\220\\350\\241\\250.xlsx",
			expected: "excel/HeroRetreat_武将归隐表.xlsx",
		},
		{
			name:     "纯英文文件名",
			input:    "excel/Hero.xlsx",
			expected: "excel/Hero.xlsx",
		},
		{
			name:     "混合中英文-配置文件",
			input:    "config/\\351\\205\\215\\347\\275\\256\\346\\226\\207\\344\\273\\266.json",
			expected: "config/配置文件.json",
		},
		{
			name:     "多个中文-文件名",
			input:    "\\346\\226\\207\\344\\273\\266\\345\\220\\215.txt",
			expected: "文件名.txt",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "无效八进制序列-保持原样",
			input:    "file\\999.txt",
			expected: "file\\999.txt",
		},
		{
			name:     "八进制紧挨普通字符",
			input:    "file\\346\\226\\207test.txt",
			expected: "file文test.txt",
		},
		{
			name:     "充值表-验证流水线乱码场景",
			input:    "excel/Recharge_\\345\\205\\205\\345\\200\\274\\350\\241\\250.xlsx",
			expected: "excel/Recharge_充值表.xlsx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, feishu.DecodeOctalEscape(tt.input))
		})
	}
}

// TestDecodeGitPath_QuotedPath 测试 git core.quotepath=true 时输出带引号的路径解码
// 场景：Docker Alpine 容器中 git 默认 quotepath=true，
// 输出格式为 "excel/Survey_\350\205...xlsx"（带双引号包裹）
// 解码后应去除引号，确保 .xlsx 后缀匹配能正常工作
func TestDecodeGitPath_QuotedPath(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expected         string
		expectXlsxSuffix bool // 解码后是否应以 .xlsx 结尾
	}{
		{
			name:             "带引号的中文名xlsx-腾讯问卷表",
			input:            "\"excel/Survey_\\350\\205\\276\\350\\256\\257\\351\\227\\256\\345\\215\\267\\350\\241\\250.xlsx\"",
			expected:         "excel/Survey_腾讯问卷表.xlsx",
			expectXlsxSuffix: true,
		},
		{
			name:             "带引号的中文名xlsx-武将归隐表",
			input:            "\"excel/HeroRetreat_\\346\\255\\246\\345\\260\\206\\345\\275\\222\\351\\232\\220\\350\\241\\250.xlsx\"",
			expected:         "excel/HeroRetreat_武将归隐表.xlsx",
			expectXlsxSuffix: true,
		},
		{
			name:             "不带引号的纯英文路径",
			input:            "excel/Hero.xlsx",
			expected:         "excel/Hero.xlsx",
			expectXlsxSuffix: true,
		},
		{
			name:             "带引号的非xlsx文件",
			input:            "\"excel/\\351\\205\\215\\347\\275\\256.json\"",
			expected:         "excel/配置.json",
			expectXlsxSuffix: false,
		},
		{
			name:             "空字符串",
			input:            "",
			expected:         "",
			expectXlsxSuffix: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeGitPath(tt.input)
			assert.Equal(t, tt.expected, result)
			if tt.expectXlsxSuffix {
				assert.True(t, strings.HasSuffix(strings.ToLower(result), ".xlsx"),
					"解码后路径应以 .xlsx 结尾，实际: %q", result)
			}
		})
	}
}

// BenchmarkDecodeOctalEscape 性能测试
func BenchmarkDecodeOctalEscape(b *testing.B) {
	input := "excel/HeroRetreat_\\346\\255\\246\\345\\260\\206\\345\\275\\222\\351\\232\\220\\350\\241\\250.xlsx"
	for i := 0; i < b.N; i++ {
		feishu.DecodeOctalEscape(input)
	}
}
