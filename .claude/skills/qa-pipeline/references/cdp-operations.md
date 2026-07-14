# 蓝鲸流水线 CDP 操作手册

> 通过 web-access skill（CDP Proxy）操作蓝鲸流水线的关键步骤和选择器。

## 前置条件

- 需安装 web-access skill：`https://github.com/eze-is/web-access`
- Chrome 需开启 CDP 端口（`chrome://inspect/#remote-debugging` → 勾选 Allow remote debugging）
- 用户需在 Chrome 中已登录蓝鲸平台（devops.devcloud.ztgame.com）

## 关常量

```
PIPELINE_URL=https://devops.devcloud.ztgame.com/projects/xcard/pipelines/p-5e007d6a5e3e472e9b3abb99b48e1064
```

## 操作流程：启动新构建

### 1. 检查 CDP 环境

```bash
node "$CLAUDE_SKILL_DIR/scripts/check-deps.mjs"
```

### 2. 打开流水线页面

```bash
curl -s "http://localhost:3456/new?url=https://devops.devcloud.ztgame.com/projects/xcard/pipelines/p-5e007d6a5e3e472e9b3abb99b48e1064"
# 返回 {"targetId":"XXX"}，后续用这个 targetId
```

### 3. 确认页面加载

```bash
curl -s "http://localhost:3456/eval?target=XXX" -d "document.title"
# 期望：包含"配表检查-测试"
```

### 4. 点击"启动新构建"按钮

```bash
curl -s -X POST "http://localhost:3456/click?target=XXX" -d 'button.ant-btn-primary'
# 验证：返回 {"clicked":true,"text":"启动新构建"}
```

**注意**：页面上有多个 `button.ant-btn-primary`，但"启动新构建"是主页面的 primary 按钮，`/click` 默认选第一个匹配的。如果点错，需要用更精确的选择器。

### 5. 修改启动变量

对话框弹出后，变量由**成对的两个 input** 组成（2026-07-09 实测确认）：

- `disabled input`（`.ant-input-disabled`）：显示**变量名**（如 `CI_CLAUDE_FEISHU_ROBOT`）
- 紧邻的下一个 `editable input`（`input.ant-input.w-full`）：**变量值**

```
[disabled: CI_CLAUDE_FEISHU_ROBOT]  [editable: 36732a0b-...]
[disabled: CI_CLAUDE_FEISHU_DM_APP_ID] [editable: cli_a94b...]
...
```

### ⚠️ 不要按 idx 顺序假设变量

变量在对话框的显示顺序会随「变量设置」页签的「显示选项」配置变化，**不能假设 `inputs[0]` 一定是 `CI_CLAUDE_FEISHU_ROBOT`**。必须按 disabled label 的 value 定位对应的 editable input。

### ⚠️ 不要在同一次 eval 里连续 setVal 多个变量

React 批处理会导致连续设值丢失（实测：连设 inputs[0]/[1]/[2]，前两个被重置回默认值，只有最后一个生效）。**逐个变量、每次单独 eval 设值并验证**。

### 正确的设值方法：按变量名定位 + 逐个设值

```bash
# 设单个变量（以 CI_CLAUDE_FEISHU_ROBOT=none 为例）
curl -s -X POST "http://localhost:3456/eval?target=XXX" -d '(() => {
  const all = [...document.querySelectorAll(".ant-modal input")];
  const labelInput = all.find(i => i.disabled && i.value === "CI_CLAUDE_FEISHU_ROBOT");
  if (!labelInput) return { error: "变量名 label 未找到" };
  const valueInput = all[all.indexOf(labelInput) + 1];
  const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value").set;
  setter.call(valueInput, "none");
  valueInput.dispatchEvent(new Event("input", { bubbles: true }));
  valueInput.dispatchEvent(new Event("change", { bubbles: true }));
  valueInput.dispatchEvent(new Event("blur", { bubbles: true }));
  return { varName: labelInput.value, newValue: valueInput.value };
})()'
```

设值后**立即读回验证**（`valueInput.value` 是否为预期值），确认生效再设下一个。

### 批量设值的正确做法（逐个，非一次性）

对每个要改的变量，**单独发一次上述 eval**，而不是在一个 eval 里连设多个。例如要屏蔽通知（ROBOT=none、清空 DM 凭证），分 3 次调用，每次验证返回的 `newValue`。
```

### 6. 点击"执行"按钮

**关键**：执行按钮在 antd Modal 对话框内，不能直接用 `button.ant-btn-primary`（会点错到主页面的按钮）。

```bash
curl -s -X POST "http://localhost:3456/eval?target=XXX" -d \
  "document.querySelector('.ant-modal button.ant-btn-primary').click()"
```

### 7. 确认启动成功

```bash
curl -s "http://localhost:3456/eval?target=XXX" -d "location.href"
# 期望 URL 包含 buildId=b-xxxxxxx
```

### 8. 清理 tab

```bash
curl -s "http://localhost:3456/close?target=XXX"
```

## 已知问题与解决

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| `/click` 点错按钮 | 多个 `button.ant-btn-primary` | 用 eval + querySelector 精确定位 |
| 变量修改不生效 | React 受控组件需要触发事件 | 用 `nativeInputValueSetter` + `input`/`change` 事件 |
| 对话框按钮找不到 | Modal DOM 在 `.ant-modal` 下 | 选择器加 `.ant-modal` 前缀 |
| 对话框 `display:none` | 点击执行后对话框已关闭 | 说明启动成功，检查 URL 确认 |
| tab 内容读取为空 | Vue SPA 懒渲染，内容不在 tabpane 内 | 用 `[data-v-app]` 根节点读取 |

## 读取页面内容（Vue SPA 注意事项）

蓝鲸平台是 Vue 3 SPA，tab pane 使用懒加载/条件渲染。**不能**用 `innerText` 从 `.ant-tabs-tabpane` 读取内容。

### 错误方式

```bash
# ❌ 返回空 — tab 内容通过 Vue 虚拟 DOM 渲染，不在 .ant-tabs-tabpane 内
curl -s -X POST "http://localhost:3456/eval?target=XXX" -d \
  "document.querySelector('.ant-tabs-tabpane-active').innerText"
```

### 正确方式

```bash
# ✅ 从 Vue 3 应用根节点读取（拿到"当前已渲染"的页面内容）
curl -s -X POST "http://localhost:3456/eval?target=XXX" -d \
  "document.querySelector('[data-v-app]').innerText"
```

> ⚠️ **`[data-v-app]` 不是"完整快照"，而是"当前已渲染 DOM 的快照"**。tab 内详细配置（变量表、Job/插件卡片）是切 tab 时才挂载的——**不切 tab，`[data-v-app]` 同样拿不到目标 tab 的内容**。实测（2026-07-09）：编辑页默认在「流程配置」tab 时 `innerText` 仅 273 字（只有导航/tab名/按钮），切到「变量设置」tab + sleep 后涨到 844 字（变量列表才灌入）。要读某 tab 的内容，必须按下方「读取编辑页 tab 内容」流程先切 tab。

### 读取编辑页 tab 内容

编辑流水线页面的各 tab（流程配置、变量设置、触发设置等），需要先切换到目标 tab 再读取：

```bash
# 1. 点击目标 tab 内的 .ant-tabs-tab-btn 元素（直接 .click() 外层 .ant-tabs-tab 可能无效）
#    用索引定位：[0]流程配置 [1]流水线设置 [2]变量设置 [3]触发设置 [4]代码源 [5]流水线版本 [6]执行通知
curl -s -X POST "http://localhost:3456/eval?target=XXX" -d \
  "document.querySelectorAll('.ant-tabs-tab')[3].querySelector('.ant-tabs-tab-btn').click()"

# 2. 等待 tab 内容渲染（Vue 异步渲染需要短暂等待）
sleep 2

# 3. 读取完整页面内容
curl -s -X POST "http://localhost:3456/eval?target=XXX" -d \
  "document.querySelector('[data-v-app]').innerText"
```

### ⚠️ 提取结构化数据：勿用宽泛 textContent 选择器

从页面提取结构化数据（如变量表的全部变量名/值）时，**不要**用 `[data-v-app]` 配合 `querySelectorAll('td,span,div,input').textContent` 这类宽泛选择器——父容器 `textContent` 包含所有子孙节点文本，多个选择器又重复命中同一节点，导致同一份数据重复十几次（实测变量名重复 10+ 次）。

推荐做法（按精确度从高到低）：
- **读表单值**：变量表的变量名/默认值都是 `input`，直接读 `.value`
- **正则提取**：从 `[data-v-app].innerText` 用正则匹配（如 `/CI_CLAUDE_\w+/g`），适合"只需判断某值是否存在"
- **定位行容器**：先找到表格行/卡片的最小容器再读 `innerText`，避免从根节点读

## 查看构建结果

```bash
# 打开构建详情页
curl -s "http://localhost:3456/new?url=PIPELINE_URL?buildId=b-XXXX"

# 查看执行状态
curl -s -X POST "http://localhost:3456/eval?target=XXX" -d \
  "document.querySelector('.bkci-pipeline-status, [class*=status]')?.innerText"
```

## 查看插件脚本内容（Content）

排查 Shell 插件（如 `claude`）执行问题时，常需要看该插件配置的脚本正文。构建详情页（`PIPELINE_URL?buildId=b-XXXX`）操作：

### 1. 选中报错插件节点

插件树节点选择器 `.name-wrapper`（含插件名 `.name`）。选中后 className 含 `active`，状态含 `FAILED`/`SUCCEED`：

```bash
# 选中 claude 插件（按名匹配，避免传中文损坏）
curl -s -X POST "http://localhost:3456/eval?target=XXX" -d \
  "(() => { const n = [...document.querySelectorAll('.name-wrapper')].find(x => x.querySelector('.name')?.innerText.trim()==='claude'); if(n) n.click(); return n?.className; })()"
```

### 2. 切换到「配置」页签

`.menus` 容器内第 2 个 span（文本「配置」）：

```bash
curl -s -X POST "http://localhost:3456/eval?target=XXX" -d \
  "(() => { const s=[...document.querySelectorAll('.menus span')]; s[1]&&s[1].click(); return s[1]?.innerText.trim(); })()"
```

`.menus` 下 span 顺序：`[0]插件日志/全量日志`、`[1]配置`、`[2..]换行/显示行号/显示时间戳`、`[末]fullscreen`。

Content 字段 DOM 路径：`.menu-content` > `.ant-collapse-content-active` > `.field` > `.monaco-editor` > `.overflow-guard` > `.view-lines`。

### ⚠️ 陷阱：后台 tab 读不到 monaco 内容

web-access 所有操作在**后台 tab**（不可见）进行。**monaco editor 在后台 tab 不渲染内容**——浏览器后台节流（`document.visibilityState="hidden"`，`.view-lines` 尺寸 0×0，`viewLineCount=0`）。

- 现象：`.menu-content .view-lines` innerText 为空；`window.monaco` 未暴露；无 `require`/`define`；Vue3 prod build 无组件实例
- **无法绕过**：`document.visibilityState` 只读，dispatch `focus`/`visibilitychange`/`resize` 均无效

**拿脚本正文的可靠方法**：请用户在其**前台 tab** 点 Content 代码框 → `Ctrl+A` → `Ctrl+C` → 粘贴过来。monaco 在前台正常渲染、可选中复制。

> **已知 Content（claude 插件，2026-07-09 通过蓝鲸 API 确认）**：蓝鲸把这段脚本生成到 `/tmp/devops_script*.sh` 执行——这正是 build.sh 的 PWD 变成 `/tmp`、进而触发 `go.mod not found` 的根因（详见 `docker/CLAUDE.md`「PWD 陷阱」）。`settings.json.<model>` 后缀随当前模型变化（曾为 `.glm`，现为 `.kimi`）：
> ```bash
> #!/usr/bin/env bash
> bash /root/ExcelChecker/rain-qa-func/docker/build.sh
> cp ~/.claude/settings.json.kimi /root/ExcelChecker/rain-qa-func/docker/settings.json
> bash /root/ExcelChecker/rain-qa-func/docker/ci_pipeline.sh
> docker image prune -f
> docker container prune -f
> ```
>
> **获取 Content 的可靠方式**：通过蓝鲸 API `GET /ms/process/api/user/pipelines/xcard/<pipelineId>` 读取 `stages[].containers[].elements[].script` 字段（见 [show-config.md](show-config.md)），比操作 monaco editor 可靠。
