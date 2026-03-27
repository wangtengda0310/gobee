package feishu

type MsgData struct {
	MsgType string   `json:"msg_type"`
	Content *Content `json:"content,omitempty"`
	Card    Card     `json:"card,omitempty"`
}

type Content struct {
	Text string `json:"text"`
}

type Card interface {
	isFeishuCard()
}

// TemplateCard JsonCard 飞书卡片 预设结构体
type TemplateCard struct {
	Type string   `json:"type"`
	Data CardData `json:"data"`
}

func (t *TemplateCard) isFeishuCard() {}

type CardData struct {
	TemplateId          string                 `json:"template_id"`
	TemplateVersionName string                 `json:"template_version_name"`
	TemplateVariable    map[string]interface{} `json:"template_variable"`
}

// JsonCard 飞书卡片 JSON 2.0 结构体
type JsonCard struct {
	Schema   string    `json:"schema,omitempty"`    // 卡片 JSON 结构的版本。默认为 1.0。要使用 JSON 2.0 结构，必须显示声明 2.0。
	Config   *Config   `json:"config,omitempty"`    // 配置卡片的全局行为
	CardLink *CardLink `json:"card_link,omitempty"` // 卡片整体的跳转链接
	Header   *Header   `json:"header,omitempty"`    // 标题组件相关配置
	Body     *Body     `json:"body,omitempty"`      // 卡片正文
}

func (t *JsonCard) isFeishuCard() {}

// Config 卡片全局行为配置
type Config struct {
	StreamingMode            bool             `json:"streaming_mode,omitempty"`             // 卡片是否处于流式更新模式，默认值为 false
	StreamingConfig          *StreamingConfig `json:"streaming_config,omitempty"`           // 流式更新配置
	Summary                  *Summary         `json:"summary,omitempty"`                    // 卡片摘要信息
	Locales                  []string         `json:"locales,omitempty"`                    // JSON 2.0 新增属性。用于指定生效的语言
	EnableForward            bool             `json:"enable_forward,omitempty"`             // 是否支持转发卡片。默认值为 true
	UpdateMulti              bool             `json:"update_multi,omitempty"`               // 是否为共享卡片。默认值为 true
	WidthMode                string           `json:"width_mode,omitempty"`                 // 卡片宽度模式
	UseCustomTranslation     bool             `json:"use_custom_translation,omitempty"`     // 是否使用自定义翻译数据
	EnableForwardInteraction bool             `json:"enable_forward_interaction,omitempty"` // 转发的卡片是否仍然支持回传交互
	Style                    *Style           `json:"style,omitempty"`                      // 自定义字号和颜色配置
}

// StreamingConfig 流式更新配置
type StreamingConfig struct {
	PrintFrequencyMS *PrintFrequencyMS `json:"print_frequency_ms,omitempty"` // 流式更新频率，单位：ms
	PrintStep        *PrintStep        `json:"print_step,omitempty"`         // 流式更新步长，单位：字符数
	PrintStrategy    string            `json:"print_strategy,omitempty"`     // 流式更新策略，枚举值：fast/delay
}

// PrintFrequencyMS 流式更新频率配置
type PrintFrequencyMS struct {
	Default int `json:"default,omitempty"` // 默认更新频率
	Android int `json:"android,omitempty"` // Android端更新频率
	IOS     int `json:"ios,omitempty"`     // iOS端更新频率
	PC      int `json:"pc,omitempty"`      // PC端更新频率
}

// PrintStep 流式更新步长配置
type PrintStep struct {
	Default int `json:"default,omitempty"` // 默认步长
	Android int `json:"android,omitempty"` // Android端步长
	IOS     int `json:"ios,omitempty"`     // iOS端步长
	PC      int `json:"pc,omitempty"`      // PC端步长
}

// Summary 卡片摘要信息
type Summary struct {
	Content     string            `json:"content,omitempty"`      // 自定义摘要信息
	I18nContent map[string]string `json:"i18n_content,omitempty"` // 摘要信息的多语言配置
}

// Style 自定义样式配置
type Style struct {
	TextSize map[string]*TextSizeConfig `json:"text_size,omitempty"` // 自定义字号配置
	Color    map[string]*ColorConfig    `json:"color,omitempty"`     // 自定义颜色配置
}

// TextSizeConfig 字号配置
type TextSizeConfig struct {
	Default string `json:"default,omitempty"` // 兜底字号
	PC      string `json:"pc,omitempty"`      // 桌面端的字号
	Mobile  string `json:"mobile,omitempty"`  // 移动端的字号
}

// ColorConfig 颜色配置
type ColorConfig struct {
	LightMode string `json:"light_mode,omitempty"` // 浅色主题下的自定义颜色
	DarkMode  string `json:"dark_mode,omitempty"`  // 深色主题下的自定义颜色
}

// CardLink 卡片跳转链接
type CardLink struct {
	URL        string `json:"url,omitempty"`         // 默认链接地址
	AndroidURL string `json:"android_url,omitempty"` // Android端链接
	IOSURL     string `json:"ios_url,omitempty"`     // iOS端链接
	PCURL      string `json:"pc_url,omitempty"`      // PC端链接
}

// Header 卡片标题组件
type Header struct {
	Title           *TextElement          `json:"title,omitempty"`              // 卡片主标题。必填
	Subtitle        *TextElement          `json:"subtitle,omitempty"`           // 卡片副标题
	TextTagList     []*TextTag            `json:"text_tag_list,omitempty"`      // 标题后缀标签，最多设置3个
	I18nTextTagList map[string][]*TextTag `json:"i18n_text_tag_list,omitempty"` // 多语言标题后缀标签
	Template        string                `json:"template,omitempty"`           // 标题主题样式颜色
	Icon            *Icon                 `json:"icon,omitempty"`               // 前缀图标
	Padding         string                `json:"padding,omitempty"`            // 标题组件的内边距
}

// TextElement 文本元素
type TextElement struct {
	Tag     string `json:"tag"`               // 文本类型的标签。可选值：plain_text 和 lark_md
	Content string `json:"content,omitempty"` // 文本内容
}

// TextTag 文本标签
type TextTag struct {
	Tag       string       `json:"tag"`                  // 固定为 "text_tag"
	ElementID string       `json:"element_id,omitempty"` // 操作元素的唯一标识
	Text      *TextElement `json:"text"`                 // 标签内容
	Color     string       `json:"color,omitempty"`      // 标签颜色
}

// Icon 图标
type Icon struct {
	Tag    string `json:"tag"`               // 图标类型：standard_icon 或 custom_icon
	Token  string `json:"token,omitempty"`   // 图标的 token。仅在 tag 为 standard_icon 时生效
	Color  string `json:"color,omitempty"`   // 图标颜色。仅在 tag 为 standard_icon 时生效
	ImgKey string `json:"img_key,omitempty"` // 图片的 key。仅在 tag 为 custom_icon 时生效
}

// Body 卡片正文
type Body struct {
	Direction         string    `json:"direction,omitempty"`          // 正文或容器内组件的排列方向
	Padding           string    `json:"padding,omitempty"`            // 正文或容器内组件的内边距
	HorizontalSpacing string    `json:"horizontal_spacing,omitempty"` // 正文或容器内组件的水平间距
	HorizontalAlign   string    `json:"horizontal_align,omitempty"`   // 正文或容器内组件的水平对齐方式
	VerticalSpacing   string    `json:"vertical_spacing,omitempty"`   // 正文或容器内组件的垂直间距
	VerticalAlign     string    `json:"vertical_align,omitempty"`     // 正文或容器内组件的垂直对齐方式
	Elements          []Element `json:"elements,omitempty"`           // 组件数组
}

// Element 卡片元素接口
type Element interface {
	isElement()
}

// NewCardV2 创建新的卡片 JSON 2.0 实例
func NewCardV2() *JsonCard {
	return &JsonCard{
		Schema: SchemaV2,
		Config: &Config{
			EnableForward: true,
			UpdateMulti:   true,
		},
	}
}
