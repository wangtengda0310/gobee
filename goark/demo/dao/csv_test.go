package book_dao

import (
	"testing"
)

func TestManager_GetCsvStructMapping(t *testing.T) {
	// 创建一个Manager实例并初始化数据
	manager := Manager{
		books: []*Book{
			{
				Title:     "Go语言编程",
				Author:    "张三",
				ISBN:      "978-7-1234-5678-9",
				Publisher: "技术出版社",
			},
			{
				Title:     "Go并发编程",
				Author:    "李四",
				ISBN:      "978-7-5678-1234-5",
				Publisher: "开发出版社",
			},
		},
		authors: []*Author{
			{
				FirstName:  "张",
				SecondName: "三",
			},
			{
				FirstName:  "李",
				SecondName: "四",
			},
		},
	}

	// 调用GetCsvStructMapping方法
	result := manager.GetCsvStructMapping()

	// 验证结果
	tests := []struct {
		name     string
		expected string
		check    func(any) bool
	}{
		{
			name:     "book.csv映射",
			expected: "book.csv",
			check: func(value any) bool {
				booksPtr, ok := value.(*[]*Book)
				if !ok {
					return false
				}
				books := *booksPtr
				return len(books) == 2 && books[0] != nil && books[0].Title == "Go语言编程"
			},
		},
		{
			name:     "author.csv映射",
			expected: "author.csv",
			check: func(value any) bool {
				authorsPtr, ok := value.(*[]*Author)
				if !ok {
					return false
				}
				authors := *authorsPtr
				return len(authors) == 2 && authors[0] != nil && authors[0].FirstName == "张"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := result[tt.expected]
			if !exists {
				t.Errorf("期望在结果中找到键 '%s'，但不存在", tt.expected)
				return
			}

			if !tt.check(value) {
				t.Errorf("键 '%s' 的值验证失败", tt.expected)
			}
		})
	}
}

func TestManager_GetCsvStructMappingEmptyManager(t *testing.T) {
	// 测试空的Manager实例
	manager := Manager{}

	result := manager.GetCsvStructMapping()

	// 即使Manager为空，也应该返回包含csvfile键的map
	if len(result) != 2 {
		t.Errorf("期望返回2个映射，但返回了 %d 个", len(result))
	}

	// 验证键是否存在
	if _, exists := result["book.csv"]; !exists {
		t.Error("期望在结果中找到 'book.csv' 键")
	}

	if _, exists := result["author.csv"]; !exists {
		t.Error("期望在结果中找到 'author.csv' 键")
	}

	// 验证值为空指针切片
	if booksPtr, ok := result["book.csv"].(*[]*Book); ok && len(*booksPtr) != 0 {
		t.Error("期望 book.csv 的值为空指针切片")
	}

	if authorsPtr, ok := result["author.csv"].(*[]*Author); ok && len(*authorsPtr) != 0 {
		t.Error("期望 author.csv 的值为空指针切片")
	}
}

func TestManager_GetCsvStructMappingPartialData(t *testing.T) {
	// 测试部分数据的Manager实例
	manager := Manager{
		books: []*Book{
			{
				Title:  "只有一本书",
				Author: "作者",
			},
		},
		// authors保持为空
	}

	result := manager.GetCsvStructMapping()

	// 验证books数据（现在是指针切片）
	if booksPtr, ok := result["book.csv"].(*[]*Book); ok {
		books := *booksPtr
		if len(books) != 1 {
			t.Errorf("期望 book.csv 有1本书，但有 %d 本", len(books))
		}
		if books[0] != nil && books[0].Title != "只有一本书" {
			t.Error("book.csv 中的书名不正确")
		}
	} else {
		t.Error("book.csv 的值不是 []*Book 类型")
	}

	// 验证authors为空（现在是指针切片）
	if authorsPtr, ok := result["author.csv"].(*[]*Author); ok {
		authors := *authorsPtr
		if len(authors) != 0 {
			t.Error("期望 author.csv 为空切片")
		}
	} else {
		t.Error("author.csv 的值不是 []*Author 类型")
	}
}