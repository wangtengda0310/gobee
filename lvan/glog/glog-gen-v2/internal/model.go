package internal

type Entry struct {
	Name      string // 字段名
	Type      string // go类型
	XMLType   string // xml原始类型
	Order     int    // 顺序
	Title     string // 注释
	Desc      string // 详细注释
	Catalog   string // base/ext
	Required  bool   // 是否必填
	ExtType   string // 扩展类型（可选）
	ExtID     string // 扩展ID（可选）
	ExtIDDict string // 扩展字典（可选）
}

type Struct struct {
	Name     string // struct名
	FuncName string // 日志函数名
	Version  string
	Desc     string
	Obj      string
	Source   string
	Code     string
	Level    string
	IsGlog   bool
	Type     string
	Trigger  string
	Use      string
	Entries  []Entry
}

type TypeMapping struct {
	XMLType string
	GoType  string
}

type TemplateData struct {
	Structs     []Struct
	BaseEntries []Entry
	TypeMap     map[string]string
}
