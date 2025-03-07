# 需求 [见此链接pdf](./客户端转表服务.pdf)
- 支持http接口
    1 代理调用自定义工具
    2 查询运行结果(客户端轮询)
- 支持调用自定义工具链
- 历史记录的管理
- 自定义工具的多版本控制

# 设计
## 使用
  golang run exporter或者
  go build exporter && ./exporter
  启动服务
### 启动参数
    -p --port 端口号 默认80
    -h --help 帮助
    -v --version 版本
    -u --upgrade 升级
## 提供http接口
### 错误码
| 状态码 | 描述 |
|--------|------|
| 202    | 任务进行中 |
| 400    | 请求参数错误 |
| 404    | 任务不存在 |

### `http://host:port/cmd?body=${request_body_type}&sse=true&onlyid=true`
    调用命令行工具 重要 必须
#### post请求
- query param: body
支持json和yaml, 通过`?body=${request_body_type}`控制,默认yaml
    json格式 body
    ```json
    {
        "cmd": "cmd",
        "version": "0.1",
        "args": ["arg1", "arg2"]
    }
    ```
    yaml格式 body
    ```yaml
        cmd: cmd
        version: 0.1
        args: [arg1, arg2]
    ```
    `cmd`
    需要执行的自定义工具链
    `args`
    自定义工具链的参数
    `version`
    支持语义版本号或latest代表最新版本
    后期可能扩展支持git commit hash
- query param: onlyid
返回结果只显示任务id, 通过`?onlyid=true`控制,默认false
任务id后面详细设计 必要
- query param: sse 不重要 不必要
是否开启http2 server send event, 通过?sse=true控制,默认false
#### 返回结果
- 如果使用`sse=true`参数 则启用http2 `server send event`将调用自定义工具链的输出流返回给客户端
- 如果使用`onlyid=true`参数 则返回任务id
- 如果使用`sse=false`参数 则返回调用自定义工具链的输出流
#### 支持多个用户多任务访问
需要保存日志,包括任务id,时间,请求参数,任务结果
#### TODO
- [ ] 日志落地设计
- [ ] 接口使用说明维护

### `http://host:port/result/${id}?sse=true`
- 带有任务id
查询运行结果,返回工具链输出流
如果不使用`sse=true`且任务仍在执行返回202
### `http://host:port/result/help`
返回接口使用说明,使用说明如何维护后面单独设计 不重要 todo
使用内嵌静态web资源展示[接口使用说明](http-doc.txt)

## 如何使用本程序
```bash
go run exporter -p 8080
```

## 其他优化
- 自定义工作任务链工作留模板
- 良好的帮助文档
1 命令行工具的使用说明
2 http接口文档
- 记录历史任务
- 分布式支持
- 日志持久化
- 命令行工具的版本控制和自动更新
- 支持linux环境
### 任务ID
- 格式：UUID v4
- 示例：`550e8400-e29b-41d4-a716-446655440000`