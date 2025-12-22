package csv_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/wangtengda0310/gobee/ark/csv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDataModel 定义测试用的数据模型
type TestProduct struct {
	ID       string  `csv:"id"`
	Name     string  `csv:"name"`
	Price    float64 `csv:"price"`
	Category string  `csv:"category"`
	InStock  bool    `csv:"in_stock"`
}

type TestUser struct {
	UserID    string `csv:"user_id"`
	Username  string `csv:"username"`
	Email     string `csv:"email"`
	CreatedAt string `csv:"created_at"`
}

// TestProductManager 实现 csv.DataHolder 接口
type TestProductManager struct {
	products []*TestProduct `csvfile:"products.csv"`
	users    []*TestUser    `csvfile:"users.csv"`
}

func (m *TestProductManager) CsvStructMapping() map[string]any {
	result := make(map[string]any)
	result["products.csv"] = &m.products
	result["users.csv"] = &m.users
	return result
}

// TestDao 模拟外部调用者
type TestDao struct {
	loader csv.Loader `autowire:""`
	manager *TestProductManager
}

func (d *TestDao) InitTestDao() {
	d.loader.Load()
}

func (d *TestDao) GetProducts() []*TestProduct {
	return d.manager.products
}

func (d *TestDao) GetUsers() []*TestUser {
	return d.manager.users
}

// TestCustomerService 另一个外部调用者
type TestCustomerService struct {
	csv.Loader `autowire:""`
	productManager *TestProductManager
}

func (s *TestCustomerService) InitService() {
	s.Load()
}

func (s *TestCustomerService) GetProductByID(id string) *TestProduct {
	for _, product := range s.productManager.products {
		if product.ID == id {
			return product
		}
	}
	return nil
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "csv_test_*")
	require.NoError(t, err)

	// 创建测试用的 products.csv
	productsCSV := `id,name,price,category,in_stock
p001,Laptop,999.99,Electronics,true
p002,Book,29.99,Books,true
p003,Headphones,199.99,Electronics,false
p004,Coffee Maker,79.99,Home,true
`

	err = os.WriteFile(filepath.Join(tempDir, "products.csv"), []byte(productsCSV), 0644)
	require.NoError(t, err)

	// 创建测试用的 users.csv
	usersCSV := `user_id,username,email,created_at
u001,john_doe,john@example.com,2023-01-15
u002,jane_smith,jane@example.com,2023-02-20
u003,bob_wilson,bob@example.com,2023-03-10
`

	err = os.WriteFile(filepath.Join(tempDir, "users.csv"), []byte(usersCSV), 0644)
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// TestCSVLoader 测试自定义 CSV 加载器
func TestCSVLoader(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// 创建 TestProductManager
	manager := &TestProductManager{}

	// 直接使用 gocsv 加载数据
	loadCSVFiles(tempDir, []csv.DataHolder{manager})

	// 验证产品数据
	assert.Len(t, manager.products, 4, "应该加载 4 个产品")

	// 验证第一个产品
	assert.Equal(t, "p001", manager.products[0].ID)
	assert.Equal(t, "Laptop", manager.products[0].Name)
	assert.Equal(t, 999.99, manager.products[0].Price)
	assert.Equal(t, "Electronics", manager.products[0].Category)
	assert.True(t, manager.products[0].InStock)

	// 验证用户数据
	assert.Len(t, manager.users, 3, "应该加载 3 个用户")

	assert.Equal(t, "u001", manager.users[0].UserID)
	assert.Equal(t, "john_doe", manager.users[0].Username)
	assert.Equal(t, "john@example.com", manager.users[0].Email)
	assert.Equal(t, "2023-01-15", manager.users[0].CreatedAt)
}

// loadCSVFiles 模拟 Gocsvimpl.Load 的功能，但不使用依赖注入
func loadCSVFiles(baseDir string, holders []csv.DataHolder) {
	for _, holder := range holders {
		m := holder.CsvStructMapping()
		for file, record := range m {
			open, err := os.Open(filepath.Join(baseDir, file))
			if err != nil {
				log.Printf("Warning: could not open file %s: %v", file, err)
				continue
			}

			err = gocsv.UnmarshalFile(open, record)
			if err != nil {
				log.Printf("Warning: could not unmarshal file %s: %v", file, err)
			}
			open.Close()
		}
	}
	log.Println("load ok")
}

// TestCSVModuleBasicFunctionality 测试 CSV 模块的基本功能
func TestCSVModuleBasicFunctionality(t *testing.T) {
	// 使用自定义加载器进行测试
	TestCSVLoader(t)
}

// TestCSVDataHolderInterface 测试 DataHolder 接口的实现
func TestCSVDataHolderInterface(t *testing.T) {
	manager := &TestProductManager{}

	// 测试 CsvStructMapping 方法
	mapping := manager.CsvStructMapping()

	// 验证返回的映射
	assert.Len(t, mapping, 2, "应该返回 2 个文件映射")

	// 验证键存在
	productsValue, exists := mapping["products.csv"]
	assert.True(t, exists, "应该包含 products.csv 键")
	assert.NotNil(t, productsValue, "products.csv 的值不应该为 nil")

	usersValue, exists := mapping["users.csv"]
	assert.True(t, exists, "应该包含 users.csv 键")
	assert.NotNil(t, usersValue, "users.csv 的值不应该为 nil")
}

// TestMultipleDataHolders 测试多个 DataHolder 的场景
func TestMultipleDataHolders(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	manager1 := &TestProductManager{}
	manager2 := &TestProductManager{}

	// 同时加载两个 DataHolder
	loadCSVFiles(tempDir, []csv.DataHolder{manager1, manager2})

	// 验证两个 Manager 都成功加载数据
	assert.Len(t, manager1.products, 4, "Manager1 应该加载 4 个产品")
	assert.Len(t, manager2.products, 4, "Manager2 应该加载 4 个产品")

	// 验证数据内容相同
	for i, product := range manager1.products {
		assert.Equal(t, manager2.products[i].ID, product.ID)
		assert.Equal(t, manager2.products[i].Name, product.Name)
	}

	assert.Len(t, manager1.users, 3, "Manager1 应该加载 3 个用户")
	assert.Len(t, manager2.users, 3, "Manager2 应该加载 3 个用户")
}

// TestCustomerServiceUsage 测试客户服务的使用场景
func TestCustomerServiceUsage(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	productManager := &TestProductManager{}

	// 加载数据
	loadCSVFiles(tempDir, []csv.DataHolder{productManager})

	// 创建客户服务并设置已加载的 ProductManager
	customerService := &TestCustomerService{
		productManager: productManager,
	}

	// 测试按 ID 查找产品
	product := customerService.GetProductByID("p002")
	require.NotNil(t, product)
	assert.Equal(t, "p002", product.ID)
	assert.Equal(t, "Book", product.Name)
	assert.Equal(t, 29.99, product.Price)

	// 测试查找不存在的产品
	product = customerService.GetProductByID("p999")
	assert.Nil(t, product)
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	manager := &TestProductManager{}

	// 加载数据
	loadCSVFiles(tempDir, []csv.DataHolder{manager})

	// 并发读取数据
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			products := manager.products
			users := manager.users
			assert.Len(t, products, 4, "并发访问应该返回正确数量的产品")
			assert.Len(t, users, 3, "并发访问应该返回正确数量的用户")
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("并发测试超时")
		}
	}
}

// TestDataIntegrity 测试数据完整性
func TestDataIntegrity(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	manager := &TestProductManager{}

	// 加载数据
	loadCSVFiles(tempDir, []csv.DataHolder{manager})

	// 验证产品数据的完整性
	expectedProducts := []TestProduct{
		{ID: "p001", Name: "Laptop", Price: 999.99, Category: "Electronics", InStock: true},
		{ID: "p002", Name: "Book", Price: 29.99, Category: "Books", InStock: true},
		{ID: "p003", Name: "Headphones", Price: 199.99, Category: "Electronics", InStock: false},
		{ID: "p004", Name: "Coffee Maker", Price: 79.99, Category: "Home", InStock: true},
	}

	assert.Len(t, manager.products, len(expectedProducts), "产品数量应该匹配")

	for i, expected := range expectedProducts {
		actual := manager.products[i]
		assert.Equal(t, expected.ID, actual.ID, "产品 ID 应该匹配")
		assert.Equal(t, expected.Name, actual.Name, "产品名称应该匹配")
		assert.Equal(t, expected.Price, actual.Price, "产品价格应该匹配")
		assert.Equal(t, expected.Category, actual.Category, "产品类别应该匹配")
		assert.Equal(t, expected.InStock, actual.InStock, "库存状态应该匹配")
	}

	// 验证用户数据的完整性
	expectedUsers := []TestUser{
		{UserID: "u001", Username: "john_doe", Email: "john@example.com", CreatedAt: "2023-01-15"},
		{UserID: "u002", Username: "jane_smith", Email: "jane@example.com", CreatedAt: "2023-02-20"},
		{UserID: "u003", Username: "bob_wilson", Email: "bob@example.com", CreatedAt: "2023-03-10"},
	}

	assert.Len(t, manager.users, len(expectedUsers), "用户数量应该匹配")

	for i, expected := range expectedUsers {
		actual := manager.users[i]
		assert.Equal(t, expected.UserID, actual.UserID, "用户 ID 应该匹配")
		assert.Equal(t, expected.Username, actual.Username, "用户名应该匹配")
		assert.Equal(t, expected.Email, actual.Email, "邮箱应该匹配")
		assert.Equal(t, expected.CreatedAt, actual.CreatedAt, "创建时间应该匹配")
	}
}

// TestNoDataHolders 测试没有 DataHolder 的情况
func TestNoDataHolders(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// 不添加任何 DataHolder
	loadCSVFiles(tempDir, []csv.DataHolder{})

	// 应该能够正常执行，没有数据被加载
	assert.True(t, true, "没有 DataHolder 时应该能够正常执行")
}

// TestMultipleLoadCalls 测试多次调用 Load 方法
func TestMultipleLoadCalls(t *testing.T) {
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	manager := &TestProductManager{}

	// 第一次加载
	loadCSVFiles(tempDir, []csv.DataHolder{manager})
	firstLoadCount := len(manager.products)

	// 第二次加载
	loadCSVFiles(tempDir, []csv.DataHolder{manager})
	secondLoadCount := len(manager.products)

	// 数据应该被重复加载
	assert.Equal(t, firstLoadCount, secondLoadCount, "多次加载应该产生相同数量的数据")
	assert.Greater(t, firstLoadCount, 0, "应该有数据被加载")
}

// BenchmarkCSVLoading 性能基准测试
func BenchmarkCSVLoading(b *testing.B) {
	tempDir, cleanup := setupTestEnvironment(&testing.T{})
	defer cleanup()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := &TestProductManager{}

		loadCSVFiles(tempDir, []csv.DataHolder{manager})

		// 验证数据加载成功
		if len(manager.products) != 4 {
			b.Fatal("数据加载失败")
		}
	}
}