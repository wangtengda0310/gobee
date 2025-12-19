package book_dao

import (
	"reflect"
	"unsafe"
)

type Book struct {
	Title     string `json:"title" csv:"title"`
	Author    string `json:"author" csv:"author"`
	ISBN      string `json:"isbn" csv:"isbn"`
	Publisher string `json:"publisher" csv:"publisher"`
}

type BookGetter interface {
	GetBookById(id string) Book
}

type Author struct {
	firstName  string `csv:"first_name"`
	secondName string `csv:"second_name"`
}

type AuthorGetter interface {
	GetAuthorById(int)
}

type Manager struct {
	books   []*Book   `csvfile:"book.csv"`
	authors []*Author `csvfile:"author.csv"`
}

func (m Manager) CsvStructMapping() map[string]any {
	return (&m).GetCsvStructMapping()
}

// GetCsvStructMapping 自动收集Manager结构体中带有csvfile标签的字段
// 返回map[string]any，其中key是csvfile标签的值，value是指向对应切片字段的指针
func (m *Manager) GetCsvStructMapping() map[string]any {
	result := make(map[string]any)

	// 使用指针来访问私有字段
	managerPtr := unsafe.Pointer(m)
	managerType := reflect.TypeOf(m).Elem()

	// 遍历Manager的所有字段
	for i := 0; i < managerType.NumField(); i++ {
		field := managerType.Field(i)

		// 获取csvfile标签
		csvfileTag := field.Tag.Get("csvfile")
		if csvfileTag != "" {
			// 计算字段的偏移量
			fieldOffset := field.Offset

			// 创建指向字段的指针
			fieldPtr := unsafe.Pointer(uintptr(managerPtr) + fieldOffset)

			// 创建指向切片字段的指针，这样gocsv就可以修改切片内容
			slicePtr := reflect.NewAt(field.Type, fieldPtr)

			// 将指针转换为any类型并存入map
			result[csvfileTag] = slicePtr.Interface()
		}
	}

	return result
}
