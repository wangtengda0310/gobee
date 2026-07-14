package src

// 表校验规则
// demo:
// 占领赛季最后一行时间距离现在不足5天
type sheetRule interface {
	check(data [][]string) // 当前校验规则对应表的二维数据
}

// 列校验规则
// demo:
// x行数值为非字符串格式
// y行时间格式不符合要求
type colRule interface {
	check(data map[int]string) // 当前校验规则对应的列数据 key 行号
}

// 行校验规则
// demo:
// 配置了x字段后则y字段不能为空
type rowRule interface {
	check(data map[string]string) // 当前校验规则对应的行数据 key 标题
}

// 跨表校验规则
// demo:
// x表中的y字段在z表中没有对应的id
type multiSheetRule interface { // 跨表校验
	check(data map[any]any) // todo 仔细设计一下参数
}
