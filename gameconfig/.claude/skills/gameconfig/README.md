# gameconfig Skill 目录

## 文件说明

| 文件 | 说明 | 读者 |
|------|------|------|
| `SKILL.md` | 主 skill 文件，包含完整的使用指南 | AI + 人类 |
| `abilities/AI指导.md` | AI 原生能力指导，告诉 AI 如何协助用户 | AI |

## 使用方式

### 用户
直接在对话中提及 gameconfig 相关需求：
- "审查一下装备表配置"
- "生成测试数据"
- "创建配置结构体"

### AI
1. 首先阅读 `SKILL.md` 了解基础用法
2. 遇到用户请求时，查阅 `abilities/AI指导.md` 了解如何响应
3. 按照决策树选择合适的行动

## 快速链接

- [主 Skill 文件](./SKILL.md)
- [AI 指导](./abilities/AI指导.md)
- [项目 README](../../../../../wangtengda/gobee/gameconfig/README.md)
- [AI 原生规划](../../../../../wangtengda/gobee/gameconfig/AI_NATIVE.md)
