# server-config — Server Config 扩展能力

本包为 Stream Proxy 前端页面提供扩展能力，当前主要封装 Unity 服务器列表注入与客户端配置导出（pkg/proto-test/server-config 子包）。

## 核心类型

| 类型 | 文件 | 说明 |
|------|------|------|
| `ServerXlsxConfig` | [server_config.go](server_config.go:30) | Unity 服务器列表行配置，字段与服务器配置表.xlsx 第3行一致 |
| `ServerConfigService` | [wails.go](wails.go:10) | 前端 Wails 服务，暴露 InjectUnityServer 和 ExportClientConfig |

## 核心函数

| 函数 | 文件 | 职责 |
|------|------|------|
| `InjectUnityServer(cfg)` | [server_config.go](server_config.go:89) | 写入/更新服务器配置表.xlsx 中 Id=cfg.Id 的行；IpPort 为空时自动用本机IP+HTTPListenPort 构造 |
| `ExportClientConfig(excelDir)` | [server_config.go](server_config.go:234) | 在策划配表目录下执行 export_client.bat 触发客户端导出 |

## 依赖配置

- 统一策划配表目录由 `settings.ExcelConfigService` 管理（`.rain-qa-func.json` 的 `excel_config` section），默认值 `../../config/excel`
- `ServerConfigService` 在 `ExcelDir` 为空时自动从上述配置读取

## 服务器配置表格式

- 文件位置：`{excelDir}/excel/服务器配置表.xlsx`
- Sheet 名称：`服务器配置表|Server`
- 表头（第3行字段名）：Id, ServelName, IsSave, IpPort, IpPortHeroPoint, KeepAlive, ServerZoneId
- 前4行为表头行（中文名/类型/字段名/导出标识），数据从第5行开始

## 设计决策

### InjectUnityServer 自动构造 IpPort（2026-06-17）

当 `cfg.IpPort` 为空时，后端自动获取本机非回环 IPv4 地址，结合 `cfg.HTTPListenPort` 构造为 `http://{ip}:{port}/authlogin`。这样前端只需传入抽屉面板配置的 HTTP 监听端口号，无需关心本机 IP 获取逻辑。

### ExcelDir 回退到统一配置（2026-06-17）

`ServerXlsxConfig.ExcelDir` 为空时，`ServerConfigService` 从 `settings.ExcelConfigService` 读取统一策划配表目录。前端可在 settings 页面配置此目录，持久化到 `.rain-qa-func.json`。

### ExportClientConfig 始终调用 bat、接受窗口手动关闭（2026-06-17）

`ExportClientConfig` 始终调用策划配表目录下的 `export_client.bat`，即使该目录同时存在 `export.py` 也不绕过 bat 自行拼接 python 命令。bat 内部 `start python/python.exe -B export.py client` 会在独立控制台窗口运行 export.py，而 export.py 末尾 `finally: os.system("pause")` 是无条件阻塞，导出完成后窗口需用户手动按键关闭。

曾尝试用 `echo.|` 管道向 python stdin 喂入换行让 pause 自动返回，但 `start` 创建的新控制台**不继承**父进程 stdin 重定向，pause 读取的是真实控制台键盘事件，管道换行对其无效。此为 Windows 控制台机制限制，无法在不修改策划文件（bat/export.py）的前提下自动关闭窗口，故最终方案为直接调用 bat 并接受手动关闭。

## 测试

| 文件 | 覆盖范围 |
|------|---------|
| [server_config_test.go](server_config_test.go) | InjectUnityServer 新增行/更新已有行/指定IpPort/缺目录/缺文件/IP获取失败；ExportClientConfig 执行batch/缺脚本报错 |

运行：`go test ./backend/pkg/proto-test/server-config/... -v`
