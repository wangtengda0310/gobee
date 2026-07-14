# feishu

飞书通知工具库（原 feishu-lib），提供文本/卡片消息发送、检查结果格式化输出和事件分发。

## 目录结构
```
feishu/
├── robot.go              # 飞书机器人消息发送（文本/卡片模板/JSON）
├── openapi.go            # 飞书 OpenAPI 客户端（私聊消息）
├── utils.go              # 工具函数（八进制解码、红绿卡片）
├── notify.go             # 战斗测试通知配置 + FeishuNotifier 接口
├── intercept.go          # 消息劫持服务（测试阶段使用）
├── notification/         # 检查结果通知模块
│   ├── dispatcher.go     # 事件分发器
│   ├── event.go          # 事件数据结构
│   ├── formatter.go      # 检查结果格式化器
│   └── handler.go        # 处理器接口定义
└── notification/handlers/ # 输出处理器实现
    ├── console.go        # 控制台输出
    ├── feishu.go         # 飞书群卡片消息
    ├── feishu_dm.go      # 飞书私聊消息
    └── intercept.go      # 消息劫持处理器
```

## 核心类型
| 类型 | 文件位置 | 说明 |
|------|----------|------|
| OpenAPIClient | openapi.go | 飞书 OpenAPI 客户端（私聊） |
| FeishuNotifyConfig | notify.go | 飞书通知配置 |
| FeishuNotifyConfigService | notify.go | 飞书通知配置管理服务 |
| FeishuNotifier | notify.go | 飞书通知器接口（避免循环依赖） |
| InterceptService | intercept.go | 飞书消息劫持服务（测试阶段使用） |
| CheckResultDispatcher | notification/dispatcher.go | 事件分发器 |
| CheckResultEvent | notification/event.go | 检查结果事件 |
| ErrorFormatter | notification/formatter.go | 检查结果格式化器 |

## 核心函数
| 函数 | 文件位置 | 职责 |
|------|----------|------|
| SendFeiShuRobotText | robot.go | 发送文本消息到飞书群 |
| SendFeiShuRobotCardTemplate | robot.go | 发送卡片模板消息 |
| SendFeiShuRobotCardJson | robot.go | 发送 JSON 卡片消息 |
| WarningRed | utils.go | 发送红色警告卡片 |
| SuccessGreen | utils.go | 发送绿色成功卡片 |
| DecodeOctalEscape | utils.go | 解码八进制转义序列（中文路径） |
| NewOpenAPIClient | openapi.go | 创建 OpenAPI 客户端 |
| NewDispatcher | notification/dispatcher.go | 创建事件分发器 |
| GetSummary | notification/event.go | 获取检查结果统计 |

## 依赖关系
- 被依赖：rain-excel-checker（通知）、function-test（通知）、excel-test（通知）、settings（通知配置）
