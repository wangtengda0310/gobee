# 代码生成工具需求说明

## 一、目标
将多个xml结构导出为golang代码，自动生成日志函数、参数结构体、辅助函数及单元测试。

## 二、输入输出
- 输入：xml文件路径、golang代码输出目录
- 输出：符合规范的golang代码文件

## 三、XML结构说明
- 结构：metalib > struct > entry
- struct标签：每个生成一个日志函数
- entry标签：分base/ext，详见下文

## 四、生成规则
### 1. 日志函数
- 每个struct标签生成一个日志函数，name属性为函数名
- 其他属性作为注释，每行一条

### 2. 参数结构体
- struct标签生成参数结构体，命名为函数名+Param
- base类型entry生成公共BaseParam结构体，所有参数结构体嵌入
- ext类型entry生成各自结构体字段，name驼峰化，type映射见下

### 3. 类型映射
- type属性与golang类型不一致时，需通过配置文件映射
- 映射文件格式示例：
  ```xml
  <mapping>
    <type xml="int64" go="int64"/>
    <type xml="string" go="string"/>
    <!-- 可扩展更多类型映射 -->
  </mapping>
  ```
- 未映射类型需报错或给出提示

### 4. 辅助函数
- 每个日志函数生成一个不导出的辅助函数，参数为结构体字段，顺序与order一致
- 辅助函数内部用stringbuffer拼接参数，'|'分隔，调用write写入log.txt
- 辅助函数命名规则：函数名+Helper，私有
- **辅助函数对write的调用采用依赖注入，write为参数传入，便于mock和性能测试。**

### 5. 单元测试
- 每个日志函数生成对应的单元测试
- 测试需验证日志内容写入log.txt，覆盖正常、异常、边界情况
- 测试前清理log.txt，确保环境干净
- **测试写入日志文件应使用t.TempDir()等临时文件，避免污染主日志。**

### 6. 错误处理与健壮性
- 生成代码需健壮，处理异常和空值
- type未映射、order缺失、属性缺失等需有报错或默认处理
- 注释需完整，便于维护

## 五、代码示例
```go
import (
    "log"
    "os"
    "strconv"
    "strings"
)

type WriteFunc func(contents []string)

var DefaultWriter WriteFunc = NewFileWriter("log.txt", 1)

func SetDefaultWriter(factory func() WriteFunc) {
    DefaultWriter = factory()
}

func NewFileWriter(filePath string, batchSize int) WriteFunc {
    return func(contents []string) {
        f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            log.Println("open file error:", err)
            return
        }
        defer f.Close()
        for i := 0; i < len(contents); i += batchSize {
            end := i + batchSize
            if end > len(contents) {
                end = len(contents)
            }
            batch := contents[i:end]
            f.WriteString(strings.Join(batch, "|") + "\n")
        }
    }
}

// online 日志函数
// name: online
// version: 1.0
// desc: 角色在线
// obj: 服务器事件
// source: 服务端
// code: 201010
// level: A
// isglog: true
// type: 通用
// trigger: 每分钟每个区服触发一次，统计在线角色数
// use: 实时数据采集
func OnlineLog(param OnlineLogParam) {
    onlineLogHelper(
        param.LogId,
        param.EventTime,
        param.ZoneId,
        param.ZoneName,
        param.ServerId,
        param.GameVersion,
        param.OnlineNum,
        param.AccountList,
        DefaultWriter,
    )
}

// onlineLogHelper 辅助函数
func onlineLogHelper(
    logId string,
    eventTime string,
    zoneId string,
    zoneName string,
    serverId string,
    gameVersion string,
    onlineNum int64,
    accountList string,
    write WriteFunc,
) {
    var buf strings.Builder
    buf.Grow(128)
    buf.WriteString(logId)
    buf.WriteByte('|')
    buf.WriteString(eventTime)
    buf.WriteByte('|')
    buf.WriteString(zoneId)
    buf.WriteByte('|')
    buf.WriteString(zoneName)
    buf.WriteByte('|')
    buf.WriteString(serverId)
    buf.WriteByte('|')
    buf.WriteString(gameVersion)
    buf.WriteByte('|')
    buf.WriteString(strconv.FormatInt(onlineNum, 10))
    buf.WriteByte('|')
    buf.WriteString(accountList)
    write([]string{buf.String()})
}

// Param 公共基础参数
type BaseParam struct {
    // 日志id
    LogId string
    // 事件时间
    EventTime string
    // 区服组id
    ZoneId string
    // 区服组名称
    ZoneName string
    // 区服id
    ServerId string
    // 游戏版本号
    GameVersion string
    // 发行商id
    PubId string
    // 通道id
    TunnelId string
    // 账号id
    CpsdkAccountId string
    // 游戏账号id
    GameAccountId string
    // 角色id
    RoleId string
    // 角色名
    RoleName string
}

// OnlineLogParam online 参数结构体
type OnlineLogParam struct {
    BaseParam
    // 在线人数
    OnlineNum int64
    // 账号id列表
    AccountList string
}

// chat 日志函数
// name: chat
// version: 1.0
// desc: 聊天记录
// obj: 角色事件
// source: 服务端
// code: 108010
// level: B
// isglog: true
// type: 通用
// trigger: 角色在各频道聊天时记录
// use: 
func ChatLog(param ChatLogParam) {
    chatLogHelper(
        param.LogId,
        param.EventTime,
        param.ZoneId,
        param.GameVersion,
        param.PubId,
        param.TunnelId,
        param.ServerId,
        param.CpsdkAccountId,
        param.GameAccountId,
        param.RoleId,
        param.RoleName,
        param.ChatType,
        param.ChatTxt,
        param.ChatRoleId,
        param.ChatRoleName,
        DefaultWriter,
    )
}

// chatLogHelper 辅助函数
func chatLogHelper(
    logId string,
    eventTime string,
    zoneId string,
    gameVersion string,
    pubId string,
    tunnelId string,
    serverId string,
    cpsdkAccountId string,
    gameAccountId string,
    roleId string,
    roleName string,
    chatType string,
    chatTxt string,
    chatRoleId string,
    chatRoleName string,
    write WriteFunc,
) {
    // 预估每个字段平均长度为20，15个字段，预分配320字节
    var buf strings.Builder
    buf.Grow(320)
    fields := []string{
        logId,
        eventTime,
        zoneId,
        gameVersion,
        pubId,
        tunnelId,
        serverId,
        cpsdkAccountId,
        gameAccountId,
        roleId,
        roleName,
        chatType,
        chatTxt,
        chatRoleId,
        chatRoleName,
    }
    for i, v := range fields {
        buf.WriteString(v)
        if i != len(fields)-1 {
            buf.WriteByte('|')
        }
    }
    write([]string{buf.String()})
}

// ChatLogParam chat 参数结构体
type ChatLogParam struct {
    BaseParam
    // 聊天频道
    ChatType string
    // 聊天内容
    ChatTxt string
    // 聊天对象的角色id
    ChatRoleId string
    // 聊天对象的角色名
    ChatRoleName string
}

func TestChatLog(t *testing.T) {
    // 清理log.txt，确保测试环境干净
    os.Remove("log.txt")

    // 构造测试参数
    param := ChatLogParam{
        BaseParam: BaseParam{
            LogId:          "10001",
            EventTime:      "+0800 2024-06-01 12:00:00",
            ZoneId:         "zone_1",
            GameVersion:    "v1.2.3",
            PubId:          "pub_123",
            TunnelId:       "tunnel_456",
            ServerId:       "server_789",
            CpsdkAccountId: "cpsdk_abc",
            GameAccountId:  "game_acc_001",
            RoleId:         "role_001",
            RoleName:       "测试角色",
        },
        ChatType:     "世界",
        ChatTxt:      "大家好",
        ChatRoleId:   "role_002",
        ChatRoleName: "目标角色",
    }

    // 调用日志函数
    ChatLog(param)

    // 读取log.txt内容
    data, err := os.ReadFile("log.txt")
    if err != nil {
        t.Fatalf("读取log.txt失败: %v", err)
    }
    lines := strings.Split(strings.TrimSpace(string(data)), "\n")
    if len(lines) == 0 {
        t.Fatal("log.txt没有内容")
    }

    // 构造期望输出
    expected := strings.Join([]string{
        param.LogId,
        param.EventTime,
        param.ZoneId,
        param.GameVersion,
        param.PubId,
        param.TunnelId,
        param.ServerId,
        param.CpsdkAccountId,
        param.GameAccountId,
        param.RoleId,
        param.RoleName,
        param.ChatType,
        param.ChatTxt,
        param.ChatRoleId,
        param.ChatRoleName,
    }, "|")

    if lines[0] != expected {
        t.Errorf("日志内容不正确\n期望: %s\n实际: %s", expected, lines[0])
    }
}
```

## 六、XML示例
```xml
<metalib>
  <struct name="online" version="1.0" desc="角色在线" obj="服务器事件" source="服务端" code="201010" level="A" isglog="true" type="通用" trigger="每分钟每个区服触发一次，统计在线角色数" use="实时数据采集">
    <entry name="LogId" catelogd="base" type="string" order="1" title="日志id"/>
    <entry name="EventTime" catelogd="base" type="string" order="2" title="事件时间"/>
    <entry name="ZoneId" catelogd="base" type="string" order="3" title="区服组id"/>
    <entry name="OnlineNum" catelogd="ext" type="int64" order="4" title="在线人数"/>
    <entry name="AccountList" catelogd="ext" type="string" order="5" title="账号id列表"/>
  </struct>
</metalib>
```

## 七、边界用例
- struct无entry
- entry缺少order/type
- type未映射
- 属性值为空
- 多struct、entry顺序错乱

## 八、注意事项
- 生成代码需健壮，处理异常和空值
- 注释需完整，便于维护
- 类型映射需灵活可扩展
- 日志内容顺序严格按order
- 辅助函数、单元测试需覆盖所有生成内容
- **所有生成的类型、方法、辅助函数名需全局唯一，避免命名冲突。建议每个日志类型单独生成一个包/目录。**

---

# 九、模板化与现代化最佳实践（合并新增）

### 1. 代码生成主流程模板化
- 使用`text/template`包替换原有字符串拼接，所有结构体、日志函数、辅助函数等均用模板渲染。
- 定义模板数据结构（如TemplateStruct、TemplateEntry、TemplateData），将XML解析结果和类型映射填充到模板数据中。
- 在模板中，`range .Structs`时用`$struct := .`保存当前struct作用域，`range $i, $e := .Entries`时用`len $struct.Entries`获取长度，避免作用域丢失。
- 所有Go代码注释、顺序、类型映射、字段名驼峰化等细节均在模板中实现。

### 2. 类型映射与兼容性
- 默认类型映射支持`int64`、`bigint`（映射为Go的int64）、`string`，如需扩展可通过外部xml配置。
- 读取文件统一用`os.ReadFile`，移除`ioutil`，兼容Go 1.16+。

### 3. 健壮性与边界处理
- 生成代码前，严格校验所有字段类型是否有映射，未映射时报错提示。
- 结构体、辅助函数、日志函数参数顺序严格按order排序。
- 注释完整，便于维护。

### 4. 自动化与测试
- 生成代码后自动运行`go test`，输出测试结果，确保生成代码无误。
- **每个日志函数除生成单元测试外，还需强制生成对应的性能测试（Benchmark），用于评估日志写入性能。**
- 单元测试和性能测试均可模板化，保证风格一致。
- **性能测试应避免I/O瓶颈影响评估，建议写入临时文件或mock write函数。**

#### 性能测试（Benchmark）模板建议
- 每个日志函数生成如下性能测试函数：
  ```go
  func Benchmark<日志函数名>LogHelper(b *testing.B) {
      batchSizes := []int{1, 10, 100, 1000}
      for _, batchSize := range batchSizes {
          b.Run(fmt.Sprintf("batchSize_%d", batchSize), func(b *testing.B) {
              tempDir := b.TempDir()
              filePath := filepath.Join(tempDir, "bench.log")
              writer := func(contents []string) {
                  f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
                  if err != nil {
                      b.Fatalf("open file error: %v", err)
                  }
                  defer f.Close()
                  for i := 0; i < len(contents); i += batchSize {
                      end := i + batchSize
                      if end > len(contents) {
                          end = len(contents)
                      }
                      batch := contents[i:end]
                      f.WriteString(strings.Join(batch, "\n") + "\n")
                  }
              }
              params := make([]<参数结构体>, b.N)
              gofakeit.Seed(0)
              for i := 0; i < b.N; i++ {
                  gofakeit.Struct(&params[i])
              }
              b.ResetTimer()
              for i := 0; i < b.N; i++ {
                  <日志函数名>LogHelper(
                      ...参数...
                      writer,
                  )
              }
          })
      }
  }
  ```
- 性能测试应避免I/O瓶颈影响评估（如有需要可mock write函数或写入/dev/null）。
- 建议在CI中定期跑Benchmark，监控日志写入性能变化。如有性能阈值要求，可在文档中说明如何设置和报警。

### 5. 常见模板片段
```gotemplate
{{- range .Structs }}
  {{- $struct := . }}
  ...
  {{- range $i, $e := .Entries }}
    ...
    {{- if lt $i (sub1 (len $struct.Entries)) }}
      ...
    {{- end }}
  {{- end }}
{{- end }}
```

### 6. 命名唯一化与包结构
- 所有生成的类型、方法、辅助函数名需全局唯一，避免命名冲突。
- 推荐每个日志类型单独生成一个包/目录，或在命名中加入唯一标识。

### 7. mock与依赖注入
- mock数据生成推荐使用 gofakeit/faker 等库，自动填充结构体。
- write函数采用依赖注入，便于mock和性能测试。

### 8. 临时文件与测试隔离
- 单元测试和性能测试写入临时文件，避免污染主日志。

---

