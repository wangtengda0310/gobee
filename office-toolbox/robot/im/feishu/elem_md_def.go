package feishu

// MDElement 柱状图组件
type MDElement struct {
	Tag       string `json:"tag"`                  // 组件的标签
	Content   string `json:"content,omitempty"`    // 富文本
	TextAlign string `json:"text_align,omitempty"` // 文字对齐方式
	TextSize  string `json:"text_size,omitempty"`  // 文字大小
	Margin    string `json:"margin,omitempty"`     // 组件的外边距
}

func (e *MDElement) isElement() {}
