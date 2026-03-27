package feishu

// PieElement 饼图组件
type PieElement struct {
	Tag        string     `json:"tag"`                   // 组件的标签
	ChartSpec  *ChartSpec `json:"chart_spec,omitempty"`  // 图表规格
	Preview    bool       `json:"preview,omitempty"`     // 是否预览
	ColorTheme string     `json:"color_theme,omitempty"` // 颜色主题
	Height     string     `json:"height,omitempty"`      // 高度
	Margin     string     `json:"margin,omitempty"`      // 组件的外边距
}

func (e *PieElement) isElement() {}

// ChartSpec 图表规格
type ChartSpec struct {
	Type          string        `json:"type"`                    // 图表类型：pie, bar, line 等
	Title         *ChartTitle   `json:"title,omitempty"`         // 图表标题
	Data          *ChartData    `json:"data,omitempty"`          // 图表数据
	ValueField    string        `json:"valueField,omitempty"`    // 值字段
	CategoryField string        `json:"categoryField,omitempty"` // 分类字段
	OuterRadius   float64       `json:"outerRadius,omitempty"`   // 外半径
	Legends       *ChartLegends `json:"legends,omitempty"`       // 图例配置
	Padding       *ChartPadding `json:"padding,omitempty"`       // 内边距
	Label         *ChartLabel   `json:"label,omitempty"`         // 标签配置
}

// ChartTitle 图表标题
type ChartTitle struct {
	Text string `json:"text,omitempty"` // 标题文本
}

// ChartData 图表数据
type ChartData struct {
	Values []*ChartDataValue `json:"values,omitempty"` // 数据值
}

// ChartDataValue 图表数据值
type ChartDataValue struct {
	Type  string `json:"type,omitempty"`  // 类型/分类
	Value string `json:"value,omitempty"` // 值
}

// ChartLegends 图表图例
type ChartLegends struct {
	Visible bool   `json:"visible,omitempty"` // 是否可见
	Orient  string `json:"orient,omitempty"`  // 方向：left, right, top, bottom
}

// ChartPadding 图表内边距
type ChartPadding struct {
	Left   int `json:"left,omitempty"`   // 左内边距
	Top    int `json:"top,omitempty"`    // 上内边距
	Bottom int `json:"bottom,omitempty"` // 下内边距
	Right  int `json:"right,omitempty"`  // 右内边距
}

// ChartLabel 图表标签
type ChartLabel struct {
	Visible bool `json:"visible,omitempty"` // 是否可见
}
