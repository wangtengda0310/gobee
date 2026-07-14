# 蓝鲸流水线页面定位器参考

> 蓝鲸 DevOps 平台（Vue3 + Ant Design）关键元素的 CSS 选择器，用于 web-access（CDP）操作。所有选择器均经过实测验证。
>
> **最后验证日期**：2026-07-09
>
> **优先级提醒**：读取配置用 [blueking-api.md](blueking-api.md) 的 API 更可靠；UI 选择器主要用于修改配置、点击按钮等交互场景。

---

## 项目导航与代码库页

| 目标 | 选择器 / URL | 备注 |
|------|-------------|------|
| 项目左侧菜单项 | `._menuItem_toow2_13` | 顺序：0 仪表盘 / 1 流水线 / 2 代码库 / 3 制品库 / ...；后台 tab textContent 不稳定，**用 index 定位** |
| 流水线页 | `/projects/xcard/pipelines/<id>` | 主页 |
| 流水线编辑页 | `/projects/xcard/pipelines/<id>/edit` | |
| **代码库页** | `/projects/xcard/code-library/association` | 直接输 URL（试 /repositories 会 404） |
| 关联代码库按钮 | 代码库页 `button.ant-btn-primary` | 页面唯一 primary 按钮 |

### 关联代码库对话框（`.ant-modal`）

| 字段 | 控件 | 选择器 |
|------|------|--------|
| 代码库类型 | radio（Perforce/SVN/GitLab） | `label.ant-radio-wrapper` 含 `input[value=gitlab]`；选中看 `input.checked` |
| 认证方式 | radio（AccessToken/OAuth） | `input[value=OAUTH]` / `input[value=ACCESSTOKEN]` |
| 访问凭据 | Ant Select | `.ant-modal .ant-select`（第 1 个），选项在全局 `.ant-select-dropdown` |
| 代码库地址 | Ant Select | `.ant-modal .ant-select`（第 2 个），凭据选定后加载仓库列表 |
| 名称 | input | `input[placeholder="请输入名称"]` |

> 关联流程详见 [register-repo.md](register-repo.md)。GitLab 类型强制 OAuth 凭据。

---

## 编辑页结构

```
.pipeline-editor-wrapper              ← 整个流水线编辑器
└── .pipeline-wrapper
    ├── .stage（构建触发）             ← Stage 1
    │   ├── .stage-head
    │   └── .stage-containers
    │       └── .container-wrapper
    │           └── .container         ← 触发器 Job
    │               ├── .container-header
    │               └── .element-wrapper
    │                   ├── .element.row ← 触发器插件（手动/远程/定时）
    │                   └── .footer .addAtomButton ← "添加新插件"
    └── .stage.last（构建 Stage）       ← Stage 2
        └── .container（Linux Job）
            ├── .container-header .name ← Job 名（如 "Linux"）
            └── .element-wrapper
                ├── .element.row        ← 插件项（git/linuxScript）
                │   ├── .element-handler
                │   ├── .atom-icon
                │   ├── .name           ← 插件名（任务别名）
                │   └── .extra .actions ← 复制/删除/添加
                └── .footer
```

### 插件项定位

| 目标 | 选择器 |
|------|--------|
| 所有插件 | `.element.row`（含触发器，跨 Stage 需限定容器） |
| 构建插件（不含触发器） | `.stage:not(:first-child) .element.row` 或在第二个 `.container-wrapper` 下查 |
| 插件名称 | `.element.row .name`（`innerText`） |
| 指定名称的插件 | 遍历 `.element.row` 比对 `.name` 的 `innerText` |

### Job 定位

| 目标 | 选择器 |
|------|--------|
| 所有 Job | `.container` |
| Job 名称 | `.container-header .name` |
| Job 编号（如 1-1） | `.container-header .icon.keep span` |

### 添加按钮

| 目标 | 选择器 |
|------|--------|
| 添加新插件 | `.addAtomButton`（Job 内 footer，多个取可见的那个） |
| 添加新 Job | `button.ant-btn-primary` 文本含「添加新 Job」 |

### 添加新插件流程（UI，2026-07-09 实测）

⚠️ **UI 加插件涉及搜索型 Select（代码库地址），后台 tab 选项加载不稳定。改 plugin 配置优先用 [blueking-api.md](blueking-api.md) 的 PUT model**。UI 流程仅记录备用：

1. 点 `.addAtomButton` → 弹出 `.ant-drawer` 插件选择面板
2. 插件列表是「选择」按钮数组（`button.ant-btn-primary`，按顺序：Shell/人工审核/调用流水线/飞书/GIT/Perforce/SVN...）。**GIT 是第 5 个"选择"按钮**（`querySelectorAll("button")` filter 可见后 index 6，因前面还有环境变量/搜索按钮）
3. 点 GIT 的"选择" → 抽屉变为 GIT 配置面板（仓库来源/代码库地址/分支/拉取策略/代码保存路径）
4. 代码库地址是**搜索型 Select**（后台 tab 远程加载选项不稳定）——这是 UI 加插件的主要障碍

---

## ⚠️ CDP 后台 tab 可见性（重要）

web-access 的 `/new` 创建的是**后台 tab（不激活）**，用户在浏览器里**看不到 tab 内容**：

- DOM 操作（点击、填表单、弹窗）用户全部看不到——我"看到"是因为用 `/eval` 读了 DOM，但那个 tab 没显示给用户
- 用户**偶尔能看到**的：tab 栏新标签出现、导航导致的 tab 标题变化、浏览器偶尔把 tab 切到前台
- 设计上是「最小侵入」（不打扰用户当前 tab）

**若需用户实时监督操作**：每步用 CDP `Target.activateTarget` 把 tab 切到前台（代价：会打断用户当前 tab）。默认不激活。

**结论**：CDP 后台操作是「真实但不可见」。要可靠地改流水线，**优先用蓝鲸 API（PUT model / fetch）**，不依赖 UI 渲染。

---

## 配置抽屉（点击插件后弹出）

点击 `.element.row` 后，右侧弹出 Ant Design Drawer。

| 目标 | 选择器 | 备注 |
|------|--------|------|
| 抽屉容器 | `.ant-drawer-right` | |
| 抽屉打开状态 | `.ant-drawer-right.ant-drawer-open` | 判断抽屉是否打开 |
| 抽屉关闭按钮 | `.ant-drawer-right .ant-drawer-close` | el.click() 可触发关闭 |
| **抽屉遮罩** | `.ant-drawer-mask` | ⚠️ 抽屉打开时全屏覆盖，**挡住顶部保存按钮**，必须先关闭抽屉才能保存 |
| 抽屉内容区 | `.ant-drawer-right .drawer-body` | |
| 抽屉头部 | `.ant-drawer-right .drawer-header` | |

### 抽屉内的输入框（linuxScript 插件）

抽屉内有多个 `input`，按 `drawer.querySelectorAll("input")` 顺序（idx 从 0）：

| idx | 字段 | 说明 |
|-----|------|------|
| 0 | Type 选择器 | `id` 形如 `rc_select_*`，Ant Select 组件 |
| **1** | **任务别名** | 即插件 `.name` 对应的值，修改插件名改这里 |
| 2 | SHELL（Type 值） | 只读 |
| 3 | 换行开关 | Ant Switch，value "on"/"off" |
| 4 | 执行超时（分钟） | value 如 "900" |

> git 插件的抽屉输入框顺序不同（仓库、分支等）。修改前先遍历 `input.value` 和相邻 `label` 文本定位目标字段。

### 修改 React 受控 input 值

Ant Design 表单是 React 受控组件，直接赋 `input.value` 不生效，必须用 nativeInputValueSetter + 事件：

```javascript
const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value").set;
setter.call(targetInput, "新值");
targetInput.dispatchEvent(new Event("input", { bubbles: true }));
targetInput.dispatchEvent(new Event("change", { bubbles: true }));
targetInput.dispatchEvent(new Event("blur", { bubbles: true }));
```

---

## 顶部工具栏按钮

抽屉关闭后可见。

| 目标 | 选择器 | 备注 |
|------|--------|------|
| **保存** | `button.ant-btn-primary` 含 `.anticon-save`，`textContent`==「保存」 | 不含「并执行」 |
| 保存并执行 | `button.ant-btn-default` 含 `.anticon-save`，「保存并执行」 | |
| 版本号信息 | 顶部工具栏文本「流水线版本 NNN」 | |

### 点击保存按钮（验证可用）

后台 tab 中 `textContent` 跨帧不稳定，**应在同一次 eval 中 filter + click**：

```javascript
const all = [...document.querySelectorAll("button")];
const target = all.find(b => b.querySelector(".anticon-save")
  && /ant-btn-primary/.test(b.className)
  && !(b.textContent || "").includes("并执行"));
target && target.click();
```

---

## 启动构建对话框

主页点击「启动新构建」后弹出 Ant Modal。

| 目标 | 选择器 | 备注 |
|------|--------|------|
| 对话框 | `.ant-modal` | |
| 启动新构建按钮（主页） | `button.ant-btn-primary` | 主页面首个 primary 按钮 |
| 对话框内执行按钮 | `.ant-modal button.ant-btn-primary` | 需 `.ant-modal` 前缀避免点错到主页按钮 |
| 变量值输入框 | `.ant-modal input.ant-input.w-full` | editable input（变量值） |
| 变量名 label | `.ant-modal input.ant-input-disabled` | disabled input，value 是变量名（如 CI_CLAUDE_FEISHU_ROBOT） |
| 对话框关闭 | `.ant-modal button.ant-drawer-close` 或 Esc | |

### ⚠️ 变量定位：按 disabled label 的 value，不要按 idx

变量在对话框由**成对 input** 组成：disabled input（变量名）+ 紧邻的下一个 editable input（值）。显示顺序随「变量设置」页签的「显示选项」变化，**不要假设 `inputs[0]` 是某个固定变量**。定位方式：遍历所有 input，找 `disabled && value === "目标变量名"` 的 label input，取其下一个 editable input 设值。

### ⚠️ 设值：逐个 eval，不要连续 setVal

同一次 eval 里连续 setVal 多个变量会因 React 批处理丢失（实测前几个被重置回默认）。**每个变量单独一次 eval，设值后立即读回验证**。

详见 [cdp-operations.md](cdp-operations.md)「修改启动变量」。

---

## 构建详情页（排查构建失败）

| 目标 | 选择器 | 备注 |
|------|--------|------|
| 插件树节点 | `.name-wrapper` | 含 `.name`（插件名），选中后 className 含 `active` |
| 插件状态 | `.name-wrapper` 的 className | 含 `FAILED` / `SUCCEED` |
| 日志/配置页签 | `.menus span` | `[0]` 插件日志/全量日志，`[1]` 配置 |
| 插件日志内容 | `.menu-content` | |

### ⚠️ Monaco Editor 后台不渲染

构建详情页的 Shell Content、日志等用 monaco editor 渲染。**后台 tab（`document.visibilityState="hidden"`）中 monaco 的 `.view-lines` 尺寸 0×0，内容为空**。读取 monaco 内容的可靠方式：

- Shell Content → 用 [blueking-api.md](blueking-api.md) 的"流水线完整配置"端点读 `elements[].script`
- 日志 → 请用户在前台 tab 查看，或用构建详情 API（不稳定）

### Ant Select 后台操作（2026-07-09 实测）

蓝鲸多处用 Ant Select（关联代码库的凭据/地址、流水线变量类型等）。后台 tab 操作分两种情况：

**① 预加载选项的 Select（可后台操作）**

选项 DOM 已预存（即使 dropdown `getBoundingClientRect().width===0` 不可见），**直接操作 option 不必展开 dropdown**。但 option 必须用 **`mousedown` 事件**触发选择（`click()` 不生效）：

```javascript
const dropdowns = [...document.querySelectorAll(".ant-select-dropdown")];
// dropdown 和 modal 内 .ant-select 的对应关系：靠选项内容区分（凭据名 vs URL）
const target = dropdowns[0].querySelectorAll(".ant-select-item-option");
const opt = [...target].find(o => o.textContent.trim() === "目标选项");
opt.dispatchEvent(new MouseEvent("mousedown", { bubbles: true }));
opt.dispatchEvent(new MouseEvent("click", { bubbles: true }));
```

- 验证选中：`.ant-select .ant-select-selection-item` 的 textContent
- dropdown 与 select 无 id 关联，**靠选项内容（凭据名 vs URL）或 select 在 modal 中的顺序**对应

**② 远程加载选项的 Select（后台无法操作）**

部分 Select 的选项在 dropdown 真正可见时才发请求加载（如 OAuth 凭据列表、凭据选定后的代码库地址列表）。后台 tab 不触发加载，**选项永远为空或停留在旧缓存**——这类 Select 必须前台操作。

### Ant Radio 后台切换

Ant Radio 点 `label.ant-radio-wrapper` 触发，但后台 tab textContent 不稳定。**用 input[value] 精确定位**：

```javascript
const radioInput = modal.querySelector("input[type=radio][value=OAUTH]");
radioInput.closest("label.ant-radio-wrapper").click();
```

### 后台 tab 通用规律（汇总）

凡是**需要真正渲染或远程加载**的控件，后台 tab 都做不了：

| 控件 | 后台限制 | 应对 |
|------|---------|------|
| Monaco editor | `.view-lines` 0×0，内容空 | 用蓝鲸 API 读 |
| 远程加载的 Ant Select | 选项不加载 | 前台手动 |
| 元素 textContent | 跨 eval 不稳定 | 同一次 eval 完成 filter+click，或用 index/input[value] 定位 |
| 按钮 innerText | 可能为空 | 用 `textContent` 或 `.anticon-xxx` 图标判断 |

定位稳定性排序：`input[value]` / DOM index / class + 图标 > `textContent` 匹配 > `innerText` 匹配。

---

## 已知 UI 陷阱汇总

| 陷阱 | 原因 | 应对 |
|------|------|------|
| **抽屉遮罩挡保存按钮** | `.ant-drawer-mask` 全屏覆盖 | 改完字段先点 `.ant-drawer-close` 关闭抽屉，再点保存 |
| **后台 textContent 跨帧不稳定** | 后台节流致 React 渲染帧不一致 | 同一次 eval 内完成 filter + click，不分步 |
| **按钮 innerText 为空** | 后台 tab innerText 行为异常 | 用 `textContent` 或 `querySelector(".anticon-xxx")` 判断图标 |
| **monaco editor 空白** | 后台 tab 不渲染 | 改用蓝鲸 API 读取 |
| **保存后弹到执行历史页** | 蓝鲸已知 bug | 重新进入流水线页面确认 |
| **运行中切换历史记录显示错乱** | 蓝鲸已知 bug | 以构建历史 API 的数据为准 |

---

## 页签切换（编辑页顶部）

| 页签 | 文本 | 用途 |
|------|------|------|
| 流程配置 | 默认页签 | Job/插件结构 |
| 流水线配置 | | 基础设置 |
| 变量设置 | | 启动变量定义 |
| 触发设置 | | 手动/远程/定时触发 |
| 代码源 | | GitLab 仓库、触发分支规则 |
| 流水线版本 | | 历史版本 |
| 执行通知 | | 成功/失败通知配置 |

页签切换用文本匹配点击，选择器不固定时遍历页签容器比对文本。
