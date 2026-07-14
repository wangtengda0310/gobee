package audio

const (
	Id                   = iota // 音频ID
	None                        // 音频资源说明
	Path                        // 资源路径
	PathRandom                  // 资源路径随机
	WwiseEventPath              // Wwise事件
	StopEvent                   // 音频停止时的Wwise事件
	AudioType                   // 类型
	Priority                    // 优先级
	IsRepaceSamePriority        // 是否顶替同优先级
	InitVolum                   // 初始音量（0~100）
	BgmFadeOut                  // 背景音乐淡出（毫秒）
	BgmFadeIn                   // 背景音乐淡入（毫秒）
	IsLoop                      // 是否是循环音效（仅Unity版本的Amb类型音效使用）
)
