# 字段4态值开发 - 问题与经验沉淀

> 开发日期：2026-06-04，streamproxy字段4态值完整存储系统
> 状态：已完成

---

## 核心技术问题与解决方案

### 1. JSON反序列化类型转换问题

Go 的 `json.Unmarshal` 对 `any` 类型会将数字默认解析为 `float64`。测试中需使用 `float64` 比较或 `assert.Contains()` 验证关键字段。

相关文件：`backend/pkg/proto-test/service_record_file_test.go`

### 2. Vue3 ref 引用回调函数导致运行时错误

使用 `:ref="(el) => fieldRefs[key] = el"` 回调函数方式会导致 `Cannot access 'hasChanges' before initialization`。应使用字符串 key 方式：`:ref="key"`。

相关文件：`frontend/src/pages/stream-proxy/components/req-card-editor.vue`

### 3. Vue3 变量定义顺序导致初始化错误

`watch` 回调中引用的变量必须在 `watch` 之前定义，特别是 `immediate: true` 的 watch 会立即执行回调。

相关文件：`frontend/src/pages/stream-proxy/components/req-card-editor.vue`

---

## 架构决策

### 方案选择：修改底层结构体 vs 独立文件存储

选择修改 `RecordEntry` 结构体添加 `FieldValues` 字段（而非独立文件），原因：
1. 数据一致性：FieldValues 与消息在同一 JSON 文件中
2. 减少复杂度：无需额外的文件同步和关联机制
3. 向后兼容：`json:"field_values,omitempty"` 保证旧文件可加载

**最后更新**: 2026-06-09

---

## CLI 用例目录自动定位（2026-06-15）

> 状态：已完成

### 问题

`proto-test` CLI 依赖相对路径 `cases/proto_cases/` 加载用例，必须从项目根目录运行。一旦在其他目录运行（例如 skill 从任意 cwd 调用 exe），`case list` 显示"无测试用例"，`replay` 报"用例加载失败"。

### 解决方案

在 `wails_test_case.go` 的 `getProtoCaseDir()` 中实现两级 fallback（用 `sync.Once` 缓存结果）：

1. **分支 1**：cwd 下存在 `cases/proto_cases/` → 用项目用例（GUI 和项目内 CLI 的原有行为，不变）
2. **分支 2**：否则 fallback 到 exe 同级的 `cases/`（skill 自带用例，exe 在 `bin/` 下，cases 在 `bin/` 的上级）

skill 目录结构：
```
.claude/skills/proto-test-cli/
├── bin/rain-qa-func.exe   # 预编译二进制
├── cases/*.json           # 用例快照
└── SKILL.md
```

### 关键陷阱：ensureCaseDir 的副作用

实现过程中发现一个隐藏 bug：`LoadTestCaseList`（读取操作）原来会调用 `ensureCaseDir()`，而 `ensureCaseDir` 内部是 `os.MkdirAll(getProtoCaseDir(), 0755)`。

**问题链**：
1. 从非项目目录运行，分支 1 的 `os.Stat("cases/proto_cases")` 失败
2. 分支 2 返回 fallback 路径（或兜底返回相对路径）
3. `ensureCaseDir` 用返回的路径 `MkdirAll`——如果返回的是兜底的相对路径，就会在 cwd 下**创建空的 `cases/proto_cases/` 目录**
4. 后续所有运行中，分支 1 的 `os.Stat` 都会命中这个空目录，fallback 永远失效

**修复**：读取操作（`LoadTestCaseList`）不再调用 `ensureCaseDir()`，只有写入操作（`SaveTestCase`）才创建目录。

相关文件：`backend/pkg/proto-test/wails_test_case.go`（`getProtoCaseDir`、`ensureCaseDir`、`LoadTestCaseList`）
