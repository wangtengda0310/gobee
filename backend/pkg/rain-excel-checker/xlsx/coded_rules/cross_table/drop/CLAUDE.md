# drop/ 目录文档

掉落道具相关的跨表校验规则（package: `drop`）。

## 规则文件

| 文件 | 规则类型 | 说明 | 源表 | 依赖表 |
|------|---------|------|------|--------|
| drop_item_must_in_item.go | DROP_ITEM_MUST_IN_ITEM_CHECK | 掉落道具存在性检查：DropItem 配置的道具在 Item 表存在 | DropItem | Item |
| drop_item_validity.go | DROP_ITEM_VALIDITY_CHECK | DropItem 条件和互斥检查：ReplaceGroup 引用、布尔互斥、Item 道具有效性 | DropItem | DropGroup, Item |
| drop_rule_conditional.go | DROP_RULE_CONDITIONAL_CHECK | DropRule 条件引用检查：保底机制和 EnsureItem 条件引用有效性 | DropRule | DropGroup, Item |
| drop_rule_group_id.go | DROP_RULE_GROUP_ID_CHECK | 掉落规则组ID引用检查：DropGroup、EnsureSmallGroup、EnsureBigGroup 引用存在 | DropRule | DropGroup |

## 开发注意事项

- MustHave 和 ExcludeExist 可以同时为 true（作用时机不同）
- Item 字段格式：`{道具ID;数量}{道具ID;数量}...`，使用正则解析
- ReplaceGroup 为 0 表示无替换组，不需要检查 DropGroup 表
- EnsureSmallGroup 不应与 DropGroup 完全相同
- DropGroup 表与 DropRule 在同一 Excel 文件（`掉落分组表|DropGroup` sheet）
- 所有规则都需要加载 Item 表构建有效道具 ID 集合
