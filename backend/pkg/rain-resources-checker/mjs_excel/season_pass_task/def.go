package season_pass_task

const (
	Id              = iota // 任务ID
	SeasonPassId           // 赛季ID
	Name                   // 任务名称
	Class                  // 任务类别
	SubType                // 子类型
	IsSpecial              // 是否专属任务
	IsAutoAccept           // 是否自动接取
	AcceptCond             // 接取条件
	AcceptCondParam        // 接取条件参数
	CompleteCond           // 完成条件
	Reward                 // 奖励
	PassExp                // 战令经验
	IgnoreExpLimit         // 忽略经验上限
	IsAutoSubmit           // 是否自动提交
	ResetType              // 重置类型
	SendStatus             // 发送状态
	ShowType               // 显示类型
	ShowPos                // 显示位置
	ShowDate               // 显示日期
	BeginDate              // 开始日期
	EndDate                // 结束日期
	ExpireDate             // 过期日期
)
