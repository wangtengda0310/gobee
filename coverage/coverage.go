package main

type JSONData struct {
	AtAll   bool          `json:"atAll"`
	Title   string        `json:"title"`
	Content *CoverageData `json:"content"`
	Secret  string        `json:"secret"`
	Token   string        `json:"token"`
}
type CoverageData struct {
	TotalCoverage        float64  `json:"总覆盖率"`
	Total100PercentFiles int      `json:"覆盖率达到100%的文件数"`
	New100PercentFiles   []string `json:"新达到100%覆盖率的文件"`
	NonTestFiles         int      `json:"缺少测试的文件数"`
	Non100PercentFiles   int      `json:"未达到100%覆盖率的文件数"`
}

type Storage interface {
	ReadPreviousResults() ([]string, error)
	WriteResults([]string) error
}
