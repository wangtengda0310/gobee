// Package base 提供列级别的通用校验规则
package base

import (
	"os"
	"path/filepath"
	"testing"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/json_rule"

	"github.com/stretchr/testify/assert"
)

// TestResourceCheckRule_ResourceExists 资源存在时校验通过
func TestResourceCheckRule_ResourceExists(t *testing.T) {
	// 创建临时目录和资源文件
	tmpDir := t.TempDir()
	resourcePath := filepath.Join(tmpDir, "Prefabs", "Hero")
	err := os.MkdirAll(resourcePath, 0o755)
	assert.NoError(t, err)
	resourceFile := filepath.Join(resourcePath, "hero_test.prefab")
	err = os.WriteFile(resourceFile, []byte("test"), 0o644)
	assert.NoError(t, err)

	cols := [][]string{
		{"", "", "ResourcePath", "", "Prefabs/Hero/hero_test.prefab"},
	}
	params := map[string]string{
		"clientPath": tmpDir,
	}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 0, 4, params, nil)
	assert.Empty(t, res, "资源存在时不应有错误")
}

// TestResourceCheckRule_ResourceNotExists 资源不存在时校验失败
func TestResourceCheckRule_ResourceNotExists(t *testing.T) {
	tmpDir := t.TempDir()

	cols := [][]string{
		{"", "", "ResourcePath", "", "Prefabs/Hero/not_exist.prefab"},
	}
	params := map[string]string{
		"clientPath": tmpDir,
	}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 0, 4, params, nil)
	assert.Len(t, res, 1, "资源不存在时应有一个错误")
	assert.Contains(t, res[0].Reason, "资源文件不存在")
}

// TestResourceCheckRule_EmptyCell_PathEmptyNoError 路径为空时不报错（allowEmpty=true）
// 需要 ID 列有值确保行不被判定为全空行
func TestResourceCheckRule_EmptyCell_PathEmptyNoError(t *testing.T) {
	tmpDir := t.TempDir()

	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "ResourcePath", "", ""},
	}
	params := map[string]string{
		"clientPath":                  tmpDir,
		string(json_rule.ALLOW_EMPTY): "true",
	}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 1, 4, params, nil)
	assert.Empty(t, res, "路径为空且允许空值时不应有错误")
}

// TestResourceCheckRule_EmptyCell_NotAllowed 路径为空且不允许空值时报错
// 需要 ID 列有值确保行不被判定为全空行
func TestResourceCheckRule_EmptyCell_NotAllowed(t *testing.T) {
	tmpDir := t.TempDir()

	cols := [][]string{
		{"", "", "Id", "", "1"},
		{"", "", "ResourcePath", "", ""},
	}
	params := map[string]string{
		"clientPath":                  tmpDir,
		string(json_rule.ALLOW_EMPTY): "false",
	}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 1, 4, params, nil)
	assert.Len(t, res, 1, "路径为空且不允许空值时应有错误")
	assert.Contains(t, res[0].Reason, "单元格不能为空")
}

// TestResourceCheckRule_WithPrefix 使用 prefix 参数时路径拼接正确
func TestResourceCheckRule_WithPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	// 创建 prefix 下的资源文件
	resourcePath := filepath.Join(tmpDir, "Assets", "Resources")
	err := os.MkdirAll(resourcePath, 0o755)
	assert.NoError(t, err)
	resourceFile := filepath.Join(resourcePath, "texture.png")
	err = os.WriteFile(resourceFile, []byte("test"), 0o644)
	assert.NoError(t, err)

	// cellValue 只有文件名，prefix 提供 Assets/Resources
	cols := [][]string{
		{"", "", "Texture", "", "texture.png"},
	}
	params := map[string]string{
		"clientPath": tmpDir,
		"prefix":     "Assets/Resources",
	}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 0, 4, params, nil)
	assert.Empty(t, res, "prefix 拼接正确时资源应存在")
}

// TestResourceCheckRule_WithPrefix_NotExists prefix 下资源不存在
func TestResourceCheckRule_WithPrefix_NotExists(t *testing.T) {
	tmpDir := t.TempDir()

	cols := [][]string{
		{"", "", "Texture", "", "not_exist.png"},
	}
	params := map[string]string{
		"clientPath": tmpDir,
		"prefix":     "Assets/Resources",
	}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 0, 4, params, nil)
	assert.Len(t, res, 1, "prefix 下资源不存在时应报错")
}

// TestResourceCheckRule_NoClientPath 未配置 clientPath 时不执行检查
func TestResourceCheckRule_NoClientPath(t *testing.T) {
	cols := [][]string{
		{"", "", "ResourcePath", "", "some/path.prefab"},
	}
	params := map[string]string{}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 0, 4, params, nil)
	assert.Empty(t, res, "未配置 clientPath 时不应有错误")
}

// TestResourceCheckRule_AllowComment 允许注释行时跳过检查
// 注释行：第一列（Id列）以 # 开头
func TestResourceCheckRule_AllowComment(t *testing.T) {
	tmpDir := t.TempDir()

	cols := [][]string{
		{"", "", "Id", "", "#comment"},
		{"", "", "ResourcePath", "", "Prefabs/Hero/not_exist.prefab"},
	}
	params := map[string]string{
		"clientPath":                   tmpDir,
		string(json_rule.ALLOW_COMMIT): "true",
	}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 1, 4, params, nil)
	assert.Empty(t, res, "注释行允许时应跳过检查")
}

// TestResourceCheckRule_MultipleRows 多行数据混合场景
func TestResourceCheckRule_MultipleRows(t *testing.T) {
	tmpDir := t.TempDir()
	// 创建一个存在的资源
	resourcePath := filepath.Join(tmpDir, "Audio")
	err := os.MkdirAll(resourcePath, 0o755)
	assert.NoError(t, err)
	existFile := filepath.Join(resourcePath, "bgm_main.wav")
	err = os.WriteFile(existFile, []byte("test"), 0o644)
	assert.NoError(t, err)

	cols := [][]string{
		{"", "", "AudioPath", "", "Audio/bgm_main.wav", "Audio/not_exist.wav", "", "Audio/another_missing.wav"},
	}
	params := map[string]string{
		"clientPath":                  tmpDir,
		string(json_rule.ALLOW_EMPTY): "true",
	}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 0, 4, params, nil)
	// 应有2个错误：not_exist.wav 和 another_missing.wav
	assert.Len(t, res, 2, "应有2个不存在的资源错误")
}

// TestBuildResourcePath 测试路径拼接函数
func TestBuildResourcePath(t *testing.T) {
	tests := []struct {
		name       string
		clientPath string
		prefix     string
		cellValue  string
		want       string
	}{
		{
			name:       "无prefix直接拼接",
			clientPath: "D:/work/client",
			prefix:     "",
			cellValue:  "Prefabs/Hero/test.prefab",
			want:       filepath.Join("D:/work/client", "Prefabs/Hero/test.prefab"),
		},
		{
			name:       "有prefix三层拼接",
			clientPath: "D:/work/client",
			prefix:     "Assets/Resources",
			cellValue:  "texture.png",
			want:       filepath.Join("D:/work/client", "Assets/Resources", "texture.png"),
		},
		{
			name:       "cellValue包含子目录",
			clientPath: "D:/work/client",
			prefix:     "",
			cellValue:  "Master/Card/Audio/Voice/hero.wav",
			want:       filepath.Join("D:/work/client", "Master/Card/Audio/Voice/hero.wav"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildResourcePath(tt.clientPath, tt.prefix, tt.cellValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestResourceCheckRule_ExcelRow 验证 ExcelRow 正确性
func TestResourceCheckRule_ExcelRow(t *testing.T) {
	tmpDir := t.TempDir()

	cols := [][]string{
		{"", "", "ResourcePath", "", "not_exist1.prefab", "not_exist2.prefab", "not_exist3.prefab"},
	}
	params := map[string]string{
		"clientPath": tmpDir,
	}

	bc := new(ResourceCheckRule)
	res := bc.Check("", cols, 0, 4, params, nil)
	assert.Len(t, res, 3)
	// ExcelRow 应为 startRowIdx + i + 1 = 4 + 0 + 1 = 5, 4 + 1 + 1 = 6, 4 + 2 + 1 = 7
	assert.Equal(t, 5, res[0].ExcelRow)
	assert.Equal(t, 6, res[1].ExcelRow)
	assert.Equal(t, 7, res[2].ExcelRow)
}

// TestResourceCheckRule_CommentRow_NotAllowed 不允许注释时注释行报错
func TestResourceCheckRule_CommentRow_NotAllowed(t *testing.T) {
	tmpDir := t.TempDir()

	// 构造数据：第4行是注释行（第0列以#开头）
	cols := [][]string{
		{"", "", "Id", "", "1", "#comment"},
		{"", "", "ResourcePath", "", "test.prefab", "comment_row.prefab"},
	}
	params := map[string]string{
		"clientPath": tmpDir,
	}

	bc := new(ResourceCheckRule)
	// 检查第1列（ResourcePath），startRowIdx=4
	res := bc.Check("", cols, 1, 4, params, nil)
	// 第1行数据正常检查（test.prefab不存在，报错）
	// 第2行是注释行（第一列#开头），不允许注释时报注释错误
	assert.True(t, len(res) >= 1, "不允许注释时应有错误")
}
