package _type

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/wangtengda0310/gobee/lvan/pkg/dump"
)

func TestZip(t *testing.T) {
	// 创建临时文件用于测试
	tempFile, err := os.CreateTemp("", "test_zip_*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFileName) // 测试完成后清理临时文件

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
	Zip(tempFileName)(records, "id")

	// 验证结果
	// 打开 ZIP 文件
	reader, err := zip.OpenReader(tempFileName)
	if err != nil {
		t.Fatalf("Failed to open zip file: %v", err)
	}
	defer reader.Close()

	// 检查 ZIP 文件中的文件数量
	expectedFileCount := len(records) * len(records[0]) // 每条记录有3个字段
	if len(reader.File) != expectedFileCount {
		t.Errorf("Expected %d files in zip, but got %d", expectedFileCount, len(reader.File))
	}

	// 验证每个记录和字段
	for _, record := range records {
		id := string(record["id"])
		for column, expectedValue := range record {
			// 构造 ZIP 中的文件路径
			filePath := filepath.Join(id, column)

			// 查找对应的文件
			var foundFile *zip.File
			for _, file := range reader.File {
				if file.Name == filePath {
					foundFile = file
					break
				}
			}

			if foundFile == nil {
				t.Errorf("Expected file %s in zip, but not found", filePath)
				continue
			}

			// 读取文件内容并验证
			rc, err := foundFile.Open()
			if err != nil {
				t.Errorf("Failed to open file %s in zip: %v", filePath, err)
				continue
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read file %s in zip: %v", filePath, err)
				continue
			}

			if string(content) != string(expectedValue) {
				t.Errorf("Expected content %s in file %s, but got %s", string(expectedValue), filePath, string(content))
			}
		}
	}
}

func TestZipWithoutPK(t *testing.T) {
	// 创建临时文件用于测试
	tempFile, err := os.CreateTemp("", "test_zip_no_pk_*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFileName) // 测试完成后清理临时文件

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
	Zip(tempFileName)(records)

	// 验证结果
	// 打开 ZIP 文件
	reader, err := zip.OpenReader(tempFileName)
	if err != nil {
		t.Fatalf("Failed to open zip file: %v", err)
	}
	defer reader.Close()

	// 检查 ZIP 文件中的文件数量
	expectedFileCount := len(records) * len(records[0]) // 每条记录有2个字段
	if len(reader.File) != expectedFileCount {
		t.Errorf("Expected %d files in zip, but got %d", expectedFileCount, len(reader.File))
	}

	// 验证每个记录和字段
	for i, record := range records {
		index := i
		for column, expectedValue := range record {
			// 构造 ZIP 中的文件路径
			filePath := filepath.Join(strconv.Itoa(index), column)

			// 查找对应的文件
			var foundFile *zip.File
			for _, file := range reader.File {
				if file.Name == filePath {
					foundFile = file
					break
				}
			}

			if foundFile == nil {
				t.Errorf("Expected file %s in zip, but not found", filePath)
				continue
			}

			// 读取文件内容并验证
			rc, err := foundFile.Open()
			if err != nil {
				t.Errorf("Failed to open file %s in zip: %v", filePath, err)
				continue
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Errorf("Failed to read file %s in zip: %v", filePath, err)
				continue
			}

			if string(content) != string(expectedValue) {
				t.Errorf("Expected content %s in file %s, but got %s", string(expectedValue), filePath, string(content))
			}
		}
	}
}
