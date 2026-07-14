# 配置 GIT/Shell 插件

## 修改插件配置（任务别名、分支等）

通过 UI 修改，**核心流程：打开抽屉 → 改字段值 → 关闭抽屉 → 点保存**。

### 1. 打开插件配置抽屉

```bash
# 点击目标插件（以 claude 为例）
curl -s -X POST "http://localhost:3456/eval?target=<TARGET_ID>" -d '(() => {
  const elements = [...document.querySelectorAll(".element.row")];
  const target = elements.find(e => e.querySelector(".name")?.innerText?.trim() === "claude");
  if (target) { target.click(); return { clicked: true }; }
  return { error: "not found" };
})()'
```

### 2. 修改字段值（React 受控组件）

抽屉打开后，字段是 React 受控 input，必须用 nativeInputValueSetter + dispatch 事件才能生效：

```bash
curl -s -X POST "http://localhost:3456/eval?target=<TARGET_ID>" -d '(() => {
  const drawer = document.querySelector(".ant-drawer-right");
  const inputs = [...drawer.querySelectorAll("input")];
  // 任务别名通常是 inputs[1]（inputs[0] 可能是 Type 选择器）
  const target = inputs[1];
  const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, "value").set;
  setter.call(target, "新别名");
  target.dispatchEvent(new Event("input", { bubbles: true }));
  target.dispatchEvent(new Event("change", { bubbles: true }));
  target.dispatchEvent(new Event("blur", { bubbles: true }));
  return { newValue: target.value };
})()'
```

### 3. ⚠️ 关闭抽屉（关键，否则保存按钮被遮罩挡住）

```bash
curl -s -X POST "http://localhost:3456/eval?target=<TARGET_ID>" -d '(() => {
  const closeBtn = document.querySelector(".ant-drawer-right .ant-drawer-close");
  if (closeBtn) { closeBtn.click(); return { closed: true }; }
  return { error: "close btn not found" };
})()'
```

抽屉关闭后，字段修改值保留在 React state 中。

### 4. 点击保存

```bash
curl -s -X POST "http://localhost:3456/eval?target=<TARGET_ID>" -d '(() => {
  const all = [...document.querySelectorAll("button")];
  // 同一次 eval 中 filter + click，避免后台 tab textContent 跨帧不稳定
  const target = all.find(b => b.querySelector(".anticon-save") && /ant-btn-primary/.test(b.className) && !(b.textContent||"").includes("并执行"));
  if (target) { target.click(); return { clicked: true }; }
  return { error: "save btn not found" };
})()'
```

### 5. 验证保存成功

通过蓝鲸 API 确认修改已生效（见 [show-config.md](show-config.md)）。保存成功时页面会提示"流水线保存成功"。

## 配置 GIT 插件分支

GIT 插件的分支字段是 `data.input.refName`（不是 branchName）。UI 上对应抽屉内"分支/TAG/COMMIT"输入框，用上述 React 受控组件方法修改。

## 配置 Shell 插件 Content

Shell Content 在抽屉内的 monaco editor 中，**后台 tab 无法直接编辑 monaco**。两种方式：

1. **通过蓝鲸 API PUT 整个 model**（含修改后的 `elements[].script`）——风险较高，需完整 model 和版本号
2. **请用户在前台 tab 操作**：手动点击 Shell 插件 → 在 Content 代码框编辑 → 保存

Shell Content 统一脚本模板：

```bash
#!/usr/bin/env bash
bash /root/ExcelChecker/rain-qa-func/docker/build.sh
cp ~/.claude/settings.json.<model> /root/ExcelChecker/rain-qa-func/docker/settings.json
bash /root/ExcelChecker/rain-qa-func/docker/ci_pipeline.sh

docker image prune -f
docker container prune -f
```

> `settings.json.<model>` 后缀随当前使用的模型配置文件变化（曾用 `.glm`，现为 `.kimi`）。该文件由流水线宿主机 `~/.claude/` 提供，需提前放置。

## 注意事项

- **抽屉遮罩挡保存按钮**：抽屉打开时 `.ant-drawer-mask` 全屏覆盖，必须先关闭抽屉才能点击顶部保存按钮
- **保存后页面 bug**：保存后蓝鲸可能弹出到执行历史页且不显示执行值，需重新进入流水线页面确认（已知问题）
- **后台 tab 操作限制**：el.click() 对某些按钮有效（如插件选中、抽屉关闭），但 monaco editor 无法在后台 tab 编辑
- **不要在生产流水线直接测试**，先在旧流水线验证
