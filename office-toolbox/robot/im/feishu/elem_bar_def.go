package feishu

// BarElement 柱状图组件
type BarElement struct {
	Tag        string        `json:"tag"`                   // 组件的标签
	ChartSpec  *BarChartSpec `json:"chart_spec,omitempty"`  // 柱状图规格
	Preview    bool          `json:"preview,omitempty"`     // 是否预览
	ColorTheme string        `json:"color_theme,omitempty"` // 颜色主题
	Height     string        `json:"height,omitempty"`      // 高度
	Margin     string        `json:"margin,omitempty"`      // 组件的外边距
}

func (e *BarElement) isElement() {}

// BarChartSpec 柱状图规格
type BarChartSpec struct {
	Type        string        `json:"type"`                  // 图表类型：bar
	Title       *ChartTitle   `json:"title,omitempty"`       // 图表标题
	Data        *BarChartData `json:"data,omitempty"`        // 图表数据
	XField      []string      `json:"xField,omitempty"`      // X轴字段（数组）
	YField      string        `json:"yField,omitempty"`      // Y轴字段
	SeriesField string        `json:"seriesField,omitempty"` // 系列字段
	Legends     *ChartLegends `json:"legends,omitempty"`     // 图例配置
	Padding     *ChartPadding `json:"padding,omitempty"`     // 内边距
	Label       *ChartLabel   `json:"label,omitempty"`       // 标签配置
}

// BarChartData 柱状图数据
type BarChartData struct {
	Values []*BarDataValue `json:"values,omitempty"` // 数据值
}

// BarDataValue 柱状图数据值（扩展字段）
type BarDataValue struct {
	Type  string      `json:"type,omitempty"`  // 类型/系列
	Value interface{} `json:"value,omitempty"` // 值（Y轴）
	// 其他组件特有字段可以通过 map[string]interface{} 或具体的组件结构体来处理
	// 这里使用 interface{} 表示可以包含任意字段比如 // 年份（X轴）
	Extra map[string]interface{} `json:",inline"`
}
