package _type

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

func TestDir(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "test_dir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试完成后清理临时目录

	// 准备测试数据
	records := []dump.Record{
		{
			"id":   []byte("1"),
			"name": []byte("Alice"),
			"age":  []byte("30"),
		},
		{
			"id":   []byte("2"),
			"name": []byte("Bob"),
			"age":  []byte("25"),
		},
	}

	// 调用被测试的函数
	Dir(tempDir)(records, "id")

	// 验证结果
	for _, record := range records {
		id := string(record["id"])
		recordDir := filepath.Join(tempDir, id)

		// 检查记录目录是否存在
		if _, err := os.Stat(recordDir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist, but it does not", recordDir)
		}

		// 检查每个字段文件是否存在并包含正确的数据
		for column, expectedValue := range record {
			columnFile := filepath.Join(recordDir, column)
			if _, err := os.Stat(columnFile); os.IsNotExist(err) {
				t.Errorf("Expected file %s to exist, but it does not", columnFile)
				continue
			}

			// 读取文件内容并验证
			content, err := os.ReadFile(columnFile)
			if err != nil {
				t.Errorf("Failed to read file %s: %v", columnFile, err)
				continue
			}

			if string(content) != string(expectedValue) {
				t.Errorf("Expected content %s in file %s, but got %s", string(expectedValue), columnFile, string(content))
			}
		}
	}
}

func TestDirWithoutPK(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "test_dir_no_pk")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // 测试完成后清理临时目录

	// 准备测试数据
	records := []dump.Record{
		{
			"name": []byte("Alice"),
			"age":  []byte("30"),
		},
		{
			"name": []byte("Bob"),
			"age":  []byte("25"),
		},
	}

	// 调用被测试的函数（不提供主键参数）
	Dir(tempDir)(records)

	// 验证结果
	for i := range records {
		recordDir := filepath.Join(tempDir, strconv.Itoa(i)) // 索引作为目录名

		// 检查记录目录是否存在
		if _, err := os.Stat(recordDir); os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist, but it does not", recordDir)
		}

		// 检查每个字段文件是否存在并包含正确的数据
		for column, expectedValue := range records[i] {
			columnFile := filepath.Join(recordDir, column)
			if _, err := os.Stat(columnFile); os.IsNotExist(err) {
				t.Errorf("Expected file %s to exist, but it does not", columnFile)
				continue
			}

			// 读取文件内容并验证
			content, err := os.ReadFile(columnFile)
			if err != nil {
				t.Errorf("Failed to read file %s: %v", columnFile, err)
				continue
			}

			if string(content) != string(expectedValue) {
				t.Errorf("Expected content %s in file %s, but got %s", string(expectedValue), columnFile, string(content))
			}
		}
	}
}
