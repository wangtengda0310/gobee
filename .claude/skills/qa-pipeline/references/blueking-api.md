# 蓝鲸流水线 API 参考

> 蓝鲸 CI/CD 配置有完整 JSON model，通过已登录 Chrome 的 `fetch`（`credentials:"include"` 天然携带 cookie）直接获取，**比操作 UI 可靠得多**。所有端点均在 `https://devops.devcloud.ztgame.com` 下，项目 ID 为 `xcard`，流水线 ID 为 `p-5e007d6a5e3e472e9b3abb99b48e1064`。

**最后验证日期**：2026-07-09

---

## 验证可用的 API

### 1. 流水线完整配置（最常用）

```
GET /ms/process/api/user/pipelines/xcard/p-5e007d6a5e3e472e9b3abb99b48e1064
```

返回完整 model，含 stages/containers/elements，能读到插件名、git 分支、shell 脚本全文。

**关键字段路径**：

| 字段路径 | 含义 |
|---------|------|
| `data.stages[]` | Stage 列表 |
| `stages[].containers[]` | Job（如「Linux」） |
| `containers[].elements[]` | 插件列表 |
| `elements[].name` | 任务别名（如 `claude`） |
| `elements[].atomCode` | 插件类型（`git` / `linuxScript` / `manualTrigger` / `timerTrigger` / `remoteTrigger`） |
| git: `elements[].data.input.refName` | 分支/TAG/COMMIT |
| git: `elements[].data.input.localPath` | 代码保存路径 |
| git: `elements[].data.input.repoFrom` | 仓库来源（`CUSTOM` 等） |
| git: `elements[].data.input.codeRepoId` | 代码库 ID |
| linuxScript: `elements[].script` | Shell Content 全文 |
| `data.branchRegex` | 代码源触发分支规则 |
| `data.triggerMethod` | 触发方式 |

**使用示例**见 [show-config.md](show-config.md)。

### 2. 构建历史（排查构建失败必用）

```
GET /ms/process/api/user/builds/xcard/p-5e007d6a5e3e472e9b3abb99b48e1064/history/new
```

返回最近最多 1000 条构建记录。

**每条记录关键字段**：

| 字段 | 含义 |
|------|------|
| `id` / `buildId` | 构建 ID（形如 `b-xxxx`），用于查构建详情 |
| `status` | 状态：`SUCCEED` / `FAILED` / `RUNNING` / `QUEUE` 等 |
| `trigger` | 触发方式（如「代码变更」「手动触发」） |
| `userId` | 触发人 |
| `startTime` / `endTime` | 起止时间戳（毫秒） |
| `buildNum` | 构建序号 |
| `pipelineVersion` | 对应的流水线版本号 |

**排查构建失败流程**：调用此 API 拿到最近 FAILED 记录的 `id` → 用该 buildId 查日志/详情。

### 3. 构建日志（排查失败核心）— ⭐

```
GET /ms/log/api/user/logs/xcard/{pipelineId}/{buildId}?executeCount=1&num=1
```

返回**整个构建的结构化日志流**（所有插件混合），绕过 WebSocket/UI，直接 HTTP 获取。实测 2026-07-09。

> ⚠️ **`elementId` 参数不过滤**（传不传都返回全局流）。区分插件靠每条日志的 `tag` 字段（值 = elementId）。按 `tag` 过滤即得到单个插件日志。

**返回结构**：

| 字段 | 含义 |
|------|------|
| `buildId` | 构建 ID |
| `finished` | 日志是否写完 |
| `hasMore` | 是否还有更多段（分页，`num` 递增取下一段） |
| `logs[]` | 日志行数组，每行 `{lineNo, timestamp, message, priority, tag, jobId, executeCount}` |
| `timeUsed` | 耗时 |

**`tag` 取值**（= 各 element 的 id）：`buildStartControl`（构建控制）、`startVM-1`（VM 启动）、`e-xxxx`（各构建插件，如 git/linuxScript）、`""`（空）。从 [流水线完整配置](#1-流水线完整配置最常用) GET model 拿到各 element id，再按 `tag` 过滤。

**参数**：

- `executeCount`：执行次数，一般 `1`（失败重试时递增）
- `num`：日志分段号，从 `1` 开始，`hasMore=true` 时递增取下一段（大日志分多段）

**排查流程（按插件过滤）**：

```javascript
// 1. 从 history 拿 FAILED 的 buildId
// 2. 从 model 拿目标插件的 elementId
// 3. 拉日志（不传 elementId），按 tag 过滤
const targetEid = "e-fd8403108634d819d2d568fb50";  // 目标插件 elementId
const r = await fetch(`${BASE}/ms/log/api/user/logs/xcard/${PID}/${BID}?executeCount=1&num=1`, { credentials:"include" });
const allLogs = (await r.json()).data.logs;
const pluginLogs = allLogs.filter(l => l.tag === targetEid).map(l => l.message).join("\n");
// hasMore=true 时继续 num=2,3... 拼接（注意分段时 tag 过滤要对所有段做）
```

### ⚠️ 避免 token 爆炸：浏览器端过滤 + 摘要

构建日志量大（claude 插件单次 175 行，Claude 输出含 `thinking.signature` 几千字符 base64、`result` JSON 壳）。**不要把原始 logs 全量返回到 AI 上下文**——在 eval 里（浏览器端）做过滤提取，只返回紧凑摘要。

**推荐：摘要模式 eval 模板**（一次返回结构化摘要，token 可控）：

```javascript
// 目标：拿到构建概况 + 错误 + Claude 最终结果文本，而非原始日志
(async () => {
  const r = await fetch(`${BASE}/ms/log/api/user/logs/xcard/${PID}/${BID}?executeCount=1&num=1`, { credentials:"include" });
  const data = (await r.json()).data;
  const logs = data.logs || [];
  // 1. tag 分组统计（各插件行数概览）
  const byTag = {}; logs.forEach(l => byTag[l.tag] = (byTag[l.tag]||0)+1);
  // 2. 错误/警告行（全插件，关键字过滤，截断）
  const errors = logs
    .filter(l => /error|fail|exception|错误|失败|traceback/i.test(l.message))
    .map(l => l.message.slice(0, 200))
    .slice(0, 8);
  // 3. Claude(linuxScript) 插件：只取最终 result 文本，丢掉 signature/usage/thinking
  let resultText = "";
  const claudeTag = Object.keys(byTag).find(t => t.startsWith("e-") && logs.find(l=>l.tag===t)?.message?.includes("linuxScript") || /claude|linuxScript/i.test(t));
  const resultLine = logs.filter(l => l.tag === claudeTag).map(l=>l.message)
    .find(m => m.includes('"type":"result"'));
  if (resultLine) { try { resultText = JSON.parse(resultLine).result || ""; } catch(e){} }
  return {
    status: data.status, finished: data.finished, hasMore: data.hasMore,
    tagSummary: byTag,
    errorCount: errors.length,
    errors: errors,
    claudeResult: resultText.slice(0, 2000)  // 截断，按需调
  };
})()
```

**省 token 关键点**：

| 做法 | 效果 |
|------|------|
| `JSON.parse(resultLine).result` 取正文 | 丢掉 thinking.signature（几千 base64）、usage、session_id 等 |
| 错误行 `.slice(0,200)` + 限量 8 条 | 避免单个长堆栈撑爆 |
| tag 分组返回行数而非原文 | 概览廉价，需深入再按 tag 取 |
| `claudeResult.slice(0,2000)` | 按需截断，要全文再单独取 |

> 反面教训：直接 `logs.map(l=>l.message).join("\n")` 返回全部，claude 插件一次 20k+ token，大部分是 thinking signature 垃圾。永远在 eval 里过滤后再返回。

### 深入某插件时（按需）

摘要发现某插件有问题后，再针对性拉该 tag 的日志（仍做截断/关键字过滤）：

```javascript
const pluginLogs = logs.filter(l => l.tag === targetEid);
// 只看错误行 + 头尾，不全量
const head = pluginLogs.slice(0, 10).map(l=>l.message.slice(0,150));
const errs = pluginLogs.filter(l=>/error|fail/i.test(l.message)).map(l=>l.message.slice(0,200));
const tail = pluginLogs.slice(-5).map(l=>l.message.slice(0,150));
return { head, errs, tail };
```

> 日志也通过 WebSocket 实时推送（构建详情页用），但 HTTP 端点足够用于事后排查（`finished=true` 后日志完整）。

### 4. 构建参数定义（理解蓝鲸注入变量）

```
GET /ms/process/api/user/buildParam/xcard
```

返回所有内置 `BK_CI_*` 变量的定义（name/desc/type）。对 entrypoint.sh、skill 中读取 `BK_CI_HOOK_BRANCH`、`BK_CI_START_USER_NAME` 等变量时的参考价值很高。

常见变量：`BK_CI_START_USER_NAME`（启动人）、`BK_CI_START_TYPE`（启动方式：MANUAL/TIME_TRIGGER/WEB_HOOK/...）、`BK_CI_HOOK_BRANCH`（触发分支）、`BK_CI_HOOK_REVISION`（触发提交）。

### 5. 流水线元数据

```
GET /ms/process/api/user/pipelines/xcard/p-5e007d6a5e3e472e9b3abb99b48e1064/detail
```

只返回元数据（pipelineName、version、createTime、updateTime、creator、lastModifyUser、permission、taskCount），**不含 stages/插件配置**。读插件配置用上面的"完整配置"端点。

### 6. 流水线权限

```
GET /ms/process/api/user/pipelines/xcard/p-5e007d6a5e3e472e9b3abb99b48e1064/permissions
```

返回 execMode/editMode/viewMode、execUsers 等权限信息。

---

## 修改流水线配置（PUT model）— ⭐ 最高价值

```
PUT /ms/process/api/user/pipelines/xcard/p-5e007d6a5e3e472e9b3abb99b48e1064
Content-Type: application/json
Body: <完整 model（GET 返回的 data）>
```

**绕过所有 UI 限制**（远程加载的 Ant Select、monaco editor、后台 tab 不稳定），是修改流水线最可靠的方式。实测 2026-07-09：加 GIT 插件一次成功，`status:0, data:true`。

### 标准流程

```javascript
// 1. GET 当前 model
const r = await fetch(BASE + "/ms/process/api/user/pipelines/xcard/" + PID, { credentials: "include" });
const model = (await r.json()).data;

// 2. 深拷贝 + 改 model（加/改/删 element、改 input、改变量等）
const buildStage = model.stages.find(s => s.containers?.some(c => c.elements?.some(e => e.atomCode==="git")));
const job = buildStage.containers[0];
// 例：加插件——复制现有同类 element 作模板，改 id/name/input
const newEl = JSON.parse(JSON.stringify(job.elements[0]));
newEl.id = "e-xxx" + Date.now();  // 必须唯一
newEl.name = "新插件";
newEl.data.input = { ... };       // 按 atomCode 要求填
job.elements.splice(insertIdx, 0, newEl);

// 3. PUT 回去（body = 修改后的 model）
await fetch(BASE + "/ms/process/api/user/pipelines/xcard/" + PID, {
  method: "PUT", credentials: "include",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(model)
});
```

### 关键点

- **body 是完整 model**（GET 返回的 data 直接改后 PUT 回去，不要删字段）
- **element id 必须唯一**：用 `"e-" + 前缀 + Date.now()` 生成
- **新 element 用同类作模板**（复制现有 git/linuxScript element 再改），保证 `@type`/`classType`/`logoUrl`/`version` 等字段完整
- **element 顺序 = 执行顺序**：用 splice 控制插入位置（GIT 插件要在 Shell 之前）
- **runCondition**：`PRE_TASK_SUCCESS`（前置成功时运行，默认）/ `CUSTOM_VARIABLE_MATCH_NOT_RUN`（条件跳过，如 config commit 用）
- PUT 后 model version 自动 +1，**不需要手动传 version**
- 无需特殊 header（cookie 即可），返回 `{status:0, data:true}` 表示成功

> 这是「独立操作蓝鲸」的核心能力：改插件、加插件、改变量、改分支，都能用 GET→改→PUT 完成，不依赖前端。

## 代码库 API（scm 服务，非 repository）

代码库有两套 API，数据源不同（scm 17 个 vs repository 13 个），**流水线 GIT 插件用 scm 的 id（codeRepoId）**：

```
GET /ms/scm/api/user/v1/repository/xcard    # scm 代码库列表（流水线用这套）
GET /ms/repository/api/user/repositories/xcard   # 另一套（凭据管理视角，数据不同）
```

scm 列表返回：`{id, repoName, repoUrl, repoType, credentialId, alias}`。`id` 就是 GIT 插件的 `codeRepoId`（如 mjs-skills id=463, config=314, rain-qa-func=384）。

> 关联代码库（注册新仓库）的 POST API 未确认（试探端点 404/500）。前端关联是 GitLab OAuth 流程，后台 tab 受限。**若 scm 列表已有目标仓库，直接用其 id 配置 GIT 插件即可，无需关联**。

## 不稳定 / 不可用

| 端点 | 结果 | 说明 |
|------|------|------|
| `GET /builds/xcard/{buildId}/detail` | **间歇性 500** | 首次调用曾成功返回 13KB（含构建信息和 stages），但重试返回 500「访问后台数据失败」。蓝鲸后台不稳定，依赖时需重试 |
| `GET /builds/xcard/{buildId}/status` | 500 | 不可用 |
| `GET /builds/xcard/{buildId}` | 500 | 不可用 |

---

## 未探测（按需验证）

| 端点（推测） | 用途 | 风险 |
|-------------|------|------|
| `POST /builds/xcard/{pipelineId}/start` | 启动构建 | 会真实触发构建，需构造请求体，勿随意探测 |
| `/ms/log/api/user/logs/...` | 构建日志 | 蓝鲸日志通常走 WebSocket，结构复杂 |

> ✅ 保存流水线（PUT model）已验证可用，见上方「修改流水线配置」章节——**优先用 API（GET→改→PUT），比 UI 可靠**。

---

## 通用 fetch 模板

所有 API 调用通过 CDP proxy 的 `/eval` 在用户已登录 Chrome 中执行：

```bash
curl -s -X POST "http://localhost:3456/eval?target=<TARGET_ID>" -d '(async () => {
  const resp = await fetch("https://devops.devcloud.ztgame.com/<端点路径>", { credentials: "include" });
  const json = await resp.json();
  // 处理 json.data ...
  return 结果;
})()'
```

返回的 JSON 通常是 `{ "status": 0, "data": ..., "message": "" }` 结构（status=0 表示成功，HTTP 200 但业务 status 可能非 0）。
