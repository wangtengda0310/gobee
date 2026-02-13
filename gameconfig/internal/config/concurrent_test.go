package config

import (
	"sync"
	"testing"
	"time"
)

// TestConcurrentSetAndGetMockData 测试并发设置和读取 Mock 数据
func TestConcurrentSetAndGetMockData(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode: ModeMemory,
	})

	// 准备三组不同的 mock 数据
	mockData1 := [][]string{{"id", "name"}, {"1", "a"}}
	mockData2 := [][]string{{"id", "name"}, {"2", "b"}}
	mockData3 := [][]string{{"id", "name"}, {"3", "c"}}

	var wg sync.WaitGroup
	numGoroutines := 6

	// 并发操作序列：先设置，后读取
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			// 设置数据
			switch index % 3 {
			case 0:
				loader.SetMockData(mockData1)
			case 1:
				loader.SetMockData(mockData2)
			case 2:
				loader.SetMockData(mockData3)
			case 3:
				loader.SetMockData(mockData1)
			case 4:
				loader.SetMockData(mockData2)
			case 5:
				loader.SetMockData(mockData3)
			}

			// 给更多时间让数据被设置完成
			time.Sleep(50 * time.Millisecond)

			// 读取验证
			items, err := loader.Load()
			if err != nil {
				t.Errorf("Load() error = %v", err)
			}
			if len(items) != 1 {
				t.Errorf("期望 1 条数据，得到 %d", len(items))
			}
		}(i)
	}

	wg.Wait()

	// 最终验证：数据应该是某个确定的状态
	finalData := loader.GetMockData()
	t.Logf("最终 mock 数据: %+v", finalData)

	// 验证数据是三个数据集之一
	if len(finalData) != 1 && len(finalData) != 2 {
		t.Errorf("期望 1 或 2 行数据，得到 %d 行", len(finalData))
	}

	// 验证数据完整性
	expectedData1 := [][]string{{"id", "name"}, {"3", "c"}}
	expectedData2 := [][]string{{"id", "name"}, {"2", "b"}}
	expectedData3 := [][]string{{"id", "name"}, {"1", "a"}}

	dataEqual := false
	if equalSlice(finalData, expectedData1) {
		dataEqual = true
	}
	if equalSlice(finalData, expectedData2) {
		dataEqual = true
	}
	if equalSlice(finalData, expectedData3) {
		dataEqual = true
	}

	if !dataEqual {
		t.Errorf("最终数据与任何期望数据集不匹配: got %+v", finalData)
	}
}

// equalSlice 检查两个二维字符串切片是否相等
func equalSlice(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		match := false
		for j := range b {
			if len(a[i]) != len(b[j]) {
				return false
			}
			if len(a[i]) == 0 {
				return false
			}
			for k := range a[i] {
				if a[i][k] != b[j][k] {
					return false
				}
			}
			match = true
		}
		if match {
			return true
		}
	}
	return false
}

// TestConcurrentReload 测试并发重新加载
func TestConcurrentReload(t *testing.T) {
	type TestItem struct {
		ID   int    `excel:"id"`
		Name string `excel:"name"`
	}

	data := [][]string{
		{"id", "name"},
		{"1", "item1"},
		{"2", "item2"},
	}

	loader := NewLoader[TestItem]("", "test", LoadOptions{
		Mode:     ModeMemory,
		MockData: data,
	})

	var wg sync.WaitGroup
	numGoroutines := 5

	// 先读取初始数据
	initialItems, err := loader.Load()
	if err != nil {
		t.Fatalf("Initial Load() error = %v", err)
	}

	// 并发重新加载
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := loader.Reload()
			if err != nil {
				t.Errorf("Reload() error = %v", err)
			}
			// 验证数据一致性
			if len(items) != len(initialItems) {
				t.Errorf("Reload() 数据长度不一致: %d vs %d",
					len(items), len(initialItems))
			}
		}()
	}

	wg.Wait()
}
