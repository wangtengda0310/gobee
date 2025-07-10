package generator

type TemplateEntry struct {
	Name     string // 驼峰名
	RawName  string // 原始名
	Type     string // Go类型
	XmlType  string // XML类型
	Order    int
	Catelogd string // base/ext
	Title    string
}

type TemplateStruct struct {
	Name        string   // 函数名
	ParamName   string   // 参数结构体名
	Comment     []string // 注释
	Entries     []TemplateEntry
	BaseEntries []TemplateEntry
	ExtEntries  []TemplateEntry
}

type TemplateData struct {
	Structs     []TemplateStruct
	BaseEntries []TemplateEntry
	PackageName string
}
