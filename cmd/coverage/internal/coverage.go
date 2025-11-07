package internal

// 重构Content字段修改为map[string]interface{} 并修改相关文件
type JSONData struct {
	AtAll   bool                   `json:"atAll"`
	Title   string                 `json:"title"`
	Content map[string]interface{} `json:"content"`
	Secret  string                 `json:"secret"`
	Token   string                 `json:"token"`
}

type Storage interface {
	ReadPreviousResults() ([]string, error)
	WriteResults([]string) error
}
