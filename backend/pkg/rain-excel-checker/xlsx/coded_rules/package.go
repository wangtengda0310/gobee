package coded_rules

// coded_rules 包提供 Excel 配表校验规则
// 规则按三层结构组织：
// - cross_table/: 跨表级别规则（需要读取多个 Excel 表）
// - general/: 通用表级别规则（适用于所有表）
// - table/: 表级别规则（针对单个 Excel 表的特定业务规则）
//
// 使用示例：
//   import "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/table"
//   import "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/general/column_check"
//   import "git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/coded_rules/cross_table"
