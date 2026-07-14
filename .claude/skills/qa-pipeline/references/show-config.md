# 查看流水线配置

使用 `web-access` skill 读取蓝鲸流水线配置。

## 推荐方法：蓝鲸 API（绕过 UI 渲染问题）

蓝鲸流水线配置有完整 JSON model，通过已登录 Chrome 的 `fetch`（天然携带 cookie）直接获取，**比操作 UI 更可靠**。UI 方式中 monaco editor 在后台 tab 不渲染、元素 textContent 含不可见字符导致选择器误判，API 方式可完全绕过。

### 获取完整流水线 model

```bash
curl -s -X POST "http://localhost:3456/eval?target=<TARGET_ID>" -d '(async () => {
  const resp = await fetch("https://devops.devcloud.ztgame.com/ms/process/api/user/pipelines/xcard/p-5e007d6a5e3e472e9b3abb99b48e1064", { credentials: "include" });
  const json = await resp.json();
  const stages = json.data.stages || [];
  const result = [];
  for (const stage of stages) {
    for (const job of (stage.containers || [])) {
      for (const el of (job.elements || [])) {
        const info = { stage: stage.name, job: job.name, name: el.name, atomCode: el.atomCode };
        if (el.atomCode === "git") {
          const input = el.data?.input || {};
          info.refName = input.refName;        // 分支
          info.localPath = input.localPath;    // 保存路径
        }
        if (el.atomCode === "linuxScript") {
          info.script = el.script;             // Shell Content
        }
        result.push(info);
      }
    }
  }
  return result;
})()'
```

### 关键字段对照

| 蓝鲸 model 字段 | 含义 |
|----------------|------|
| `stages[].containers[]` | Job（如「Linux」） |
| `containers[].elements[]` | 插件 |
| `elements[].name` | 任务别名（如 `claude`） |
| `elements[].atomCode` | 插件类型（`git` / `linuxScript` / `manualTrigger` / `timerTrigger` 等） |
| git: `data.input.refName` | 分支/TAG/COMMIT |
| git: `data.input.localPath` | 代码保存路径 |
| linuxScript: `script` | Shell Content 全文 |

### 当前实测配置（2026-07-09）

```
Stage「构建触发」：手动触发 / 远程触发 / 定时任务1

Stage「使用新版校验工具增量校验全表规则」Job「Linux」：
  1. config commit      (git)        refName=v0.0.8-pre-release  localPath=xcard-excel-config
  2. config schedule    (git)        refName=v0.0.8-pre-release  localPath=xcard-excel-config
  3. （GIT）rain-qa-func (git)        refName=worktree-docker     localPath=rain-qa-func
  4. claude             (linuxScript)  Shell Content 见下
```

> Job「Linux」有 **4 个插件**（2 个 git 拉取 xcard-excel-config 分别用于 commit 触发和定时触发 + 1 个 git 拉取 rain-qa-func + 1 个 Shell）。

### claude 插件的 Shell Content

```bash
#!/usr/bin/env bash
bash /root/ExcelChecker/rain-qa-func/docker/build.sh
cp ~/.claude/settings.json.kimi /root/ExcelChecker/rain-qa-func/docker/settings.json
bash /root/ExcelChecker/rain-qa-func/docker/ci_pipeline.sh

docker image prune -f
docker container prune -f
```

> `settings.json.<model>` 后缀随当前使用的模型变化（曾为 `.glm`，现为 `.kimi`）。

---

## 备选方法：UI 操作（API 不可用时）

1. 打开流水线编辑页：
   ```bash
   curl -s "http://localhost:3456/new?url=https://devops.devcloud.ztgame.com/projects/xcard/pipelines/p-5e007d6a5e3e472e9b3abb99b48e1064/edit"
   ```

2. 插件列表在 `.container-wrapper .element-row` 下，名称在 `span.name`。

3. 点击插件后右侧弹出 `ant-drawer`，内含配置表单。**注意：抽屉打开时其 `.ant-drawer-mask` 遮罩全屏覆盖，会挡住顶部保存按钮——需先关闭抽屉才能保存**。

4. Shell Content 在抽屉内的 monaco editor 中。**后台 tab 的 monaco editor 不渲染（view-line 为空）**，无法通过 DOM 读取。可靠方式是上述蓝鲸 API。

## ⚠️ 已知 UI 陷阱

- **monaco editor 后台不渲染**：后台 tab（`document.visibilityState="hidden"`）中 monaco 的 `.view-lines` 尺寸 0×0，`window.monaco` 未暴露，无 `require/define`。Content 只能通过蓝鲸 API 的 `elements[].script` 字段获取。
- **后台 tab textContent 不稳定**：跨多次 `/eval` 调用时，同一按钮的 `textContent === "保存"` 可能时而匹配时而不匹配（后台节流导致 React 渲染帧不一致）。**应在同一次 eval 中完成 filter + click**，不要分步。
- **按钮 innerText 为空**：后台 tab 中按钮 `innerText` 可能为空，用 `textContent` 或 `querySelector(".anticon-xxx")` 判断图标更可靠。
