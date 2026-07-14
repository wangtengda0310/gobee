# common

通用功能支持包，提供飞书通知、Excel 配置加载、机器人扩展和 MCP 工具注册。

## 目录结构
```
common/
├── appconfig/    # 统一配置文件 .rain-qa-func.json section 读写
├── feishu/       # 飞书通知与消息劫持
├── game/         # 游戏配置加载（英雄/卡牌/技能）
├── mcp/          # MCP 工具注册（游戏数据、工会操作）
└── robotext/     # 机器人客户端扩展（工会创建/升级）
```

## 核心类型
| 类型 | 文件位置 | 说明 |
|------|----------|------|
| FeishuNotifyConfig | feishu/notify.go:22 | 飞书通知配置 |
| FeishuNotifyConfigService | feishu/notify.go:30 | 飞书通知配置管理服务 |
| appconfig.Section | appconfig/appconfig.go:44 | 统一配置文件的 section 读写器 |
| InterceptService | feishu/intercept.go:21 | 飞书消息劫持服务（测试阶段使用） |
| GameExcelService | game/excel.go:22 | 游戏配置加载服务 |
| RobotGuildTools | mcp/robot_guild.go:12 | 工会操作 MCP 工具集 |
| GuildClient | robotext/robotext.go:159 | 工会操作客户端 |

## 核心函数
| 函数 | 文件位置 | 职责 |
|------|----------|------|
| InitExcel | game/excel.go:90 | 初始化 Excel 数据（首次 Init，后续 Reload） |
| GetAllHeroCfg | game/excel.go:135 | 获取所有英雄配置 |
| GetAllCardCfg | game/excel.go:148 | 获取所有卡牌配置 |
| GetAllSkillCfg | game/excel.go:161 | 获取所有技能配置 |
| SendNotification | feishu/notify.go:88 | 发送飞书通知 |
| GetGuildLeaders | robotext/robotext.go:100 | 查询服务器工会列表及会长账号 |
| CreateGuildWithCity | robotext/robotext.go:119 | 创建工会并设置野战城池 |
| UpgradeGuildCity | robotext/robotext.go:138 | 升级工会城池等级 |
| RegisterGameExcelTools | mcp/game_excel.go:12 | 注册游戏数据 MCP Tools |
| RegisterRobotGuildTools | mcp/robot_guild.go:24 | 注册工会操作 MCP Tools |

## 开发注意事项
- **统一配置**：function-test / excel-test / hero-wiki-check / activity-wiki-check / mcp 的配置统一存储在 `.rain-qa-func.json`，通过 `appconfig.New("section_name")` 读写各自 section
- FeishuNotifyConfigService 使用单例模式，配置文件默认 `feishu_notify_config.json`（尚未合并到统一配置）
- GameExcelService 首次调用使用 Init，后续调用使用 Reload
- RobotExtService 方法设置 120-180 秒超时
- 所有 MCP 工具注册函数使用统一错误处理格式

## 依赖关系
- 依赖：rain-robot、mcp-go-sdk
