# rain-qa-func MCP 接口使用手册

## 1. 快速开始

启动应用后，MCP 服务自动在 `http://127.0.0.1:8765` 启动。

## 2. 验证服务

### 方式一：运行自动测试脚本

```batch
rain-qa-func\pkg\settings\mcp\tests\test_mcp.bat
```

> **重要**：每次新增 MCP 工具时，必须同步更新 `rain-qa-func\pkg\settings\mcp\tests\test_mcp.bat` 测试脚本，补充对应的测试用例并运行验证。

### 方式二：使用 MCP Inspector

```bash
npx @modelcontextprotocol/inspector http://127.0.0.1:8765/sse
```

## 3. 可用 Tools

### 配置管理 Tools

| Tool | 描述 | 必需参数 |
|------|------|----------|
| `get_func_config` | 获取功能测试配置 | - |
| `save_func_config` | 保存功能测试配置 | config |
| `get_excel_config` | 获取 Excel 配置 | - |
| `save_excel_config` | 保存 Excel 配置 | config |

### 全局设置 Tools

| Tool | 描述 | 必需参数 |
|------|------|----------|
| `get_feishu_config` | 获取飞书通知配置 | - |
| `update_feishu_config` | 更新飞书通知配置（部分更新） | fei_shu_ntf?, fei_shu_guid? |
| `send_feishu_message` | 发送飞书消息（文本） | message, robot_guid? |
| `get_mcp_config` | 获取 MCP 服务配置 | - |
| `save_mcp_config` | 保存 MCP 服务配置（自动重启服务） | enabled, port, host |
| `get_mcp_status` | 获取 MCP 服务运行状态 | - |

### 功能测试 Tools

| Tool | 描述 | 必需参数 |
|------|------|----------|
| `get_case_list` | 获取测试用例列表 | filePath |
| `get_case_by_name` | 获取单个用例完整数据 | filePath, caseName |
| `get_categories` | 获取分类信息 | dirPath |
| `search_cases` | 搜索用例 | filePath, keyword |
| `run_robot_test` | 执行测试 | ip, port, prefix |
| `stop_robot_test` | 停止测试 | - |
| `is_running` | 检查运行状态 | - |
| `get_test_logs` | 获取测试日志 | - |
| `clear_test_logs` | 清除测试日志 | - |

### 战斗测试 Tools

| Tool | 描述 | 必需参数 |
|------|------|----------|
| `get_fight_cases` | 获取战斗用例列表 | dirPath |
| `run_fight_test` | 运行战斗测试 | ip, port, heroId, caseId |
| `get_hero_list` | 获取有测试用例的英雄列表 | dirPath |

### Excel 检查 Tools

| Tool | 描述 | 必需参数 |
|------|------|----------|
| `get_excel_rules` | 获取检查规则 | dir |
| `save_excel_rules` | 保存检查规则 | dir, rules |
| `check_excel_rules` | 执行检查 | dir, rules |
| `get_all_excels` | 获取所有 Excel | dirPath |
| `get_excel_sheets` | 获取单个 Excel 中的 Sheet 列表 | filePath |
| `get_table_rules` | 获取表级规则 | dir, sheetName |
| `add_table_rule` | 添加表级规则 | dir, sheetName, rule |
| `del_table_rule` | 删除表级规则 | dir, sheetName, ruleId |
| `list_table_rule_types` | 列出可用表级规则类型 | - |
| `create_excel_file` | 创建符合项目规范的 Excel 文件 | filePath, sheets |
| `filter_excel_data` | 根据条件过滤 Excel 数据 | filePath, sheetName, conditions |
| `query_excel_range` | 查询指定行范围的 Excel 数据 | filePath, sheetName, startRow |
| `preview_excel_sheet` | 预览 Sheet 数据 | filePath, sheetName, rows? |
| `get_excel_column_info` | 获取 Sheet 列详细信息 | filePath, sheetName |
| `get_git_changed_excels` | 获取 git 变更的 Excel 文件列表 | repoPath |
| `check_table_rules_only` | 只运行指定的表级规则 | dir, ruleTypes |

### 游戏数据 Tools

| Tool | 描述 | 必需参数 |
|------|------|----------|
| `get_all_hero_cfg` | 获取英雄配置 | - |
| `get_all_card_cfg` | 获取卡牌配置 | - |
| `get_all_skill_cfg` | 获取技能配置 | - |
| `get_msg_id_map` | 获取消息 ID 映射 | - |
| `get_error_code_map` | 获取错误码映射 | - |
| `get_property_type_map` | 获取属性类型映射 | - |

### Wiki 检查 Tools

| Tool | 描述 | 必需参数 |
|------|------|----------|
| `check_hero_wiki` | 执行武将 Wiki 检查，对比新旧数据返回差异 | excelDir |
| `save_hero_wiki` | 保存武将 Wiki 检查结果到指定路径 | savePath, data |

## 4. 配置

编辑 `.rain-qa-func.json` 的 `mcp` section 可修改端口和启用状态：

```json
{
  "enabled": true,
  "port": 8765,
  "host": "127.0.0.1"
}
```

## 5. 调用格式

所有调用使用标准 JSON-RPC 2.0 格式：

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "<tool_name>",
    "arguments": { ... }
  },
  "id": 1
}
```

### 参数示例

| 工具 | 参数示例 |
|------|----------|
| `get_func_config` | `{}` |
| `search_cases` | `{"filePath": "cases/fight_cases", "keyword": "测试"}` |
| `get_case_by_name` | `{"filePath": "cases/fight_cases", "caseName": "司马懿_狼顾鹰视_1"}` |
| `get_feishu_config` | `{}` |
| `send_feishu_message` | `{"message": "...", "robot_guid?": "..."}` |
| `get_mcp_status` | `{}` |

## 6. 表级规则类型

通过 `list_table_rule_types` 工具可以获取所有可用的表级规则类型。完整规则列表参见 [rain-excel-checker/docs/规则映射表.md](../rain-excel-checker/docs/规则映射表.md)。

## 7. 与 Claude Code 集成

在 Claude Code 的配置中添加 MCP 服务器：

```json
{
  "mcpServers": {
    "rain-qa-func": {
      "url": "http://127.0.0.1:8765/sse"
    }
  }
}
```

## 8. 工具详细参数

各工具的完整参数结构、返回值格式和示例，请通过以下方式查看：
- **MCP Inspector**：`npx @modelcontextprotocol/inspector http://127.0.0.1:8765/sse`
- **工具代码注释**：`backend/pkg/settings/mcp/` 目录下各工具实现

### 常用操作速查

| 操作 | 工具 |
|------|------|
| 创建 Excel | `create_excel_file` |
| 过滤 Excel | `filter_excel_data`（支持 `eq`/`neq`/`contains`/`startsWith`/`endsWith`） |
| 分页查询 | `query_excel_range`（数据行从 1 开始，每页推荐 20 行） |
| 查看 Git 变更 | `get_git_changed_excels`（`A`=新增, `M`=修改, `D`=删除） |
| 添加表级规则 | `add_table_rule` |
| 查询表级规则 | `get_table_rules` |

## 12. 指定表级规则检查

### check_table_rules_only 工具

只运行指定的表级规则（不运行列级规则和通用规则），用于增量检查或特定规则验证。

**参数结构和规则类型参见工具代码注释或 MCP Inspector。**

### 示例：指定规则检查

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "check_table_rules_only",
    "arguments": {
      "dir": "D:/work/config/excel",
      "ruleTypes": ["NEW_ROW_NOTIFY", "ROW_CHANGE_NOTIFY"]
    }
  },
  "id": 3
}
```

> **注意**：`timeRangeBefore` 参数必须使用 Go duration 格式（如 `168h`、`720h`），不支持中文单位（如 `7天`）或纯数字（如 `168`）。如果参数格式错误，规则会自动回退到默认值 `168h`（7天）并继续检查，不会中断整个检查流程。

> **综合流程**：先用 `get_git_changed_excels` 获取变更文件，再用 `check_table_rules_only` 对变更文件运行指定规则。

