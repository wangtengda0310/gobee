package load

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestZip(t *testing.T) {
	// 创建临时 ZIP 文件用于测试
	tempFile, err := os.CreateTemp("", "test_load_zip_*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFileName) // 测试完成后清理临时文件

	// 创建测试数据
	records := []map[string][]byte{
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

	// 创建 ZIP 文件并写入测试数据
	createTestZip(tempFileName, records)

	// 调用被测试的函数
	loadedRecords := Zip(tempFileName)

	// 验证结果
	if len(loadedRecords) != len(records) {
		t.Errorf("Expected %d records, but got %d", len(records), len(loadedRecords))
	}

	// 验证每个记录的内容
	for i, expectedRecord := range records {
		if i >= len(loadedRecords) {
			t.Errorf("Expected record %d not found", i)
			continue
		}

		loadedRecord := loadedRecords[i]
		for fieldName, expectedValue := range expectedRecord {
			loadedValue, exists := loadedRecord[fieldName]
			if !exists {
				t.Errorf("Expected field %s not found in record %d", fieldName, i)
				continue
			}

			if string(loadedValue) != string(expectedValue) {
				t.Errorf("Expected value %s for field %s in record %d, but got %s",
					string(expectedValue), fieldName, i, string(loadedValue))
			}
		}
	}
}

func TestIsZipFileExist(t *testing.T) {
	// 测试存在的 ZIP 文件
	tempFile, err := os.CreateTemp("", "test_exist_zip_*.zip")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFileName) // 测试完成后清理临时文件

	// 创建一个有效的 ZIP 文件
	records := []map[string][]byte{{"test": []byte("data")}}
	createTestZip(tempFileName, records)

	// 验证文件存在
	if !IsZipFileExist(tempFileName) {
		t.Error("Expected ZIP file to exist, but IsZipFileExist returned false")
	}

	// 测试不存在的 ZIP 文件
	nonExistentFile := filepath.Join(os.TempDir(), "non_existent_zip_12345.zip")
	if IsZipFileExist(nonExistentFile) {
		t.Error("Expected non-existent ZIP file to return false, but IsZipFileExist returned true")
	}
}

// createTestZip 是一个辅助函数，用于创建测试用的 ZIP 文件
func createTestZip(filename string, records []map[string][]byte) {
	// 创建 ZIP 文件
	zipFile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer zipFile.Close()

	// 创建 ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 为每条记录创建文件夹和文件
	for i, record := range records {
		// 使用索引作为目录名
		recordDir := strconv.Itoa(i)

		// 为每个字段创建文件
		for column, valueByte := range record {
			// 创建文件路径
			filePath := filepath.Join(recordDir, column)

			// 创建 ZIP 中的文件
			f, err := zipWriter.Create(filePath)
			if err != nil {
				panic(err)
			}

			// 写入数据
			_, err = f.Write(valueByte)
			if err != nil {
				panic(err)
			}
		}
	}
}
