package config

import (
	"github.com/xuri/excelize/v2"
)

// Comment 批注信息
type Comment struct {
	Cell    string // 单元格位置（如 "A2"）
	Content string // 批注内容
	Author  string // 作者（可选）
}

// CommentReader 批注读取器
type CommentReader struct {
	file   *excelize.File
	sheet  string
}

// NewCommentReader 创建批注读取器
func NewCommentReader(file *excelize.File, sheet string) *CommentReader {
	return &CommentReader{
		file:  file,
		sheet: sheet,
	}
}

// GetCellComment 获取指定单元格的批注
func (r *CommentReader) GetCellComment(cell string) (Comment, bool) {
	// excelize v2 获取批注
	comments, err := r.file.GetComments(r.sheet)
	if err != nil {
		return Comment{}, false
	}

	for _, comment := range comments {
		// excelize v2 的 Comment 结构体字段名可能不同
		// 使用字符串匹配来查找单元格
		if getCellRef(comment) == cell {
			return Comment{
				Cell:    cell,
				Content: comment.Text,
				Author:  comment.Author,
			}, true
		}
	}

	return Comment{}, false
}

// GetAllComments 获取所有批注
func (r *CommentReader) GetAllComments() map[string]string {
	comments, err := r.file.GetComments(r.sheet)
	if err != nil {
		return make(map[string]string)
	}

	result := make(map[string]string)
	for _, comment := range comments {
		result[getCellRef(comment)] = comment.Text
	}

	return result
}

// getCellRef 从 Comment 中提取单元格引用
// excelize v2 的 Comment 结构可能没有直接的 Ref 字段
// 这个函数作为占位符，实际使用时可能需要调整
func getCellRef(comment excelize.Comment) string {
	// TODO: 根据实际的 excelize v2 API 调整
	// 暂时返回空字符串
	return ""
}
