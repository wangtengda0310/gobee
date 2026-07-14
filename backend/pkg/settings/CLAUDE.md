# settings

应用配置管理和服务注册包，包含 MCP 服务、日志、路线图和聊天服务等子模块。

## 目录结构
```
settings/
├── mcp/          # MCP 服务器实现（配置、启停、连接管理）
├── serverlog/    # 日志服务（收集、级别管理、轮转）
├── roadmap/      # 路线图/计划功能展示
├── home/         # 聊天服务（含独立 CLAUDE.md）
├── mcp.go        # MCP 工具注册
├── wails.go      # Wails 前端绑定（配置管理、版本信息）
└── utils.go      # MCP 工具辅助函数
```

## 核心类型
| 类型 | 文件位置 | 说明 |
|------|----------|------|
| SettingsService | wails.go | 前端配置管理服务 |
| MCPConfig | mcp/config.go | MCP 服务配置 |
| MCPServer | mcp/server.go | MCP 服务器实例管理 |
| ServerLogService | serverlog/service.go | 日志服务 |
| **ExcelConfig** | **excel_config.go** | **统一策划配表目录配置** |
| **ExcelConfigService** | **excel_config.go** | **统一策划配表目录配置服务** |

## 核心函数
| 函数 | 文件位置 | 职责 |
|------|----------|------|
| RegisterSettingsTools | mcp.go | 注册设置相关 MCP Tools |
| GetMCPStatus | mcp/server.go | 获取 MCP 服务运行状态 |

## 子模块文档
| 目录 | 说明 | 文档 |
|------|------|------|
| home | 聊天服务 | [CLAUDE.md](./home/CLAUDE.md) |
| mcp | MCP 服务器实现 | [CLAUDE.md](./mcp/CLAUDE.md) |

## E2E 测试

| 测试文件 | 覆盖范围 |
|----------|----------|
| [`e2e/settings/settings.spec.ts`](../../../frontend/e2e/settings/settings.spec.ts) | 页面加载、飞书通知配置、MCP服务配置、服务端日志、页面布局 |
| [`e2e/roadmap.spec.ts`](../../../frontend/e2e/roadmap.spec.ts) | 抽屉打开、筛选排序、项目详情、投票、评论、提交建议 |

## 依赖关系
- 依赖：common（配置服务）、mcp-go-sdk
- 被依赖：main.go 服务注册
