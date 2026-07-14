package src

type flow interface {
	execute()
}

type locolRepo struct {
}

func (l *locolRepo) execute() {
	//TODO implement me

	// 本地读取全量表格
	// 加载所有规则
	// 执行校验
	panic("implement me")
}

type git struct{}

func (l *git) execute() {
	//TODO implement me

	// git 读取merge/普通提交
	// 加载所有规则
	// 筛选使用与commit文件的规则
	// 读取关联表数据
	// 执行校验规则
	panic("implement me")
}
