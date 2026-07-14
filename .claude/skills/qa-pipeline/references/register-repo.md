# 关联（注册）代码库

蓝鲸「代码库」页关联新的 Git 仓库（如 mjs-skills），供流水线 GIT 插件引用。

> **2026-07-09 探索确认**。关联流程涉及 Ant Select 下拉，后台 tab 操作不稳定，**推荐前台手动**（见末尾）。

## 代码库页

```
https://devops.devcloud.ztgame.com/projects/xcard/code-library/association
```

导航：项目左侧菜单「代码库」（菜单项 `._menuItem_toow2_13`，index 2）。直接输 URL 也可。

## 关联按钮

页面右侧「关联代码库」按钮：`button.ant-btn-primary`（页面上唯一的 primary 按钮）。

点击后弹出 `.ant-modal` 关联对话框。

## 关联表单字段

| 字段 | 控件 | 说明 |
|------|------|------|
| 代码库类型 | radio（Perforce / SVN / **GitLab**） | 选 GitLab |
| 认证方式 | radio（**AccessToken** / OAuth） | 选 AccessToken（对应 authType=HTTP） |
| 访问凭据 | select（已保存的 credential） | 选有目标仓库访问权限的凭据 |
| 代码库地址 | select（凭据选定后列出可访问的仓库） | 选目标仓库（如 mjs-skills） |
| 名称 | input | 别名，如 `mjs-skills` |

### 字段选择器（CDP 用）

- 代码库类型 radio：`label.ant-radio-wrapper`，按 textContent 匹配「GitLab」
- radio input 的 value：`perforce` / `svn` / `gitlab`；选中状态看 `input.checked`
- 认证方式 radio value：`ACCESSTOKEN` / `OAUTH`
- 访问凭据 / 代码库地址：`.ant-modal .ant-select`（2 个，按顺序）
- 名称 input：`.ant-modal input[placeholder="请输入名称"]`

## 确定按钮 + 验证

- 确定：`.ant-modal button.ant-btn-primary`
- 验证：关联成功后，代码库列表 API（见 [blueking-api.md](blueking-api.md)）能查到新仓库；失败的凭据蓝鲸会即时报错

## ⚠️ GitLab 关联强制 OAuth 凭据

GitLab 类型的代码库关联，访问凭据**只允许本人创建的 OAuth 凭据**（提示「仅可以使用由本人创建的 OAuth 凭据」），AccessToken 凭据会被拒。关联后蓝鲸用该 OAuth 凭据访问 GitLab。

## ⚠️ CDP 后台 tab 的限制（已踩坑）

关联表单的「访问凭据」「代码库地址」是 Ant Select。后台 tab 有两层限制：

1. **预加载选项的 Select（可后台操作）**：选项 DOM 预存，直接操作 option，用 `mousedown` 事件触发（见 [page-selectors.md](page-selectors.md)「Ant Select 后台操作」）
2. **远程加载选项的 Select（后台无法操作）**：OAuth 凭据列表需要 dropdown 真正可见时才发请求加载，后台 tab（`visibilityState=hidden`）不触发加载，选项永远为空或显示旧缓存——**OAuth 凭据无法在后台 tab 选中**

因此「GitLab + OAuth 凭据」组合在 CDP 后台 tab **无法完成**（卡在选 OAuth 凭据），必须前台手动。这与 monaco editor 后台不渲染是同一类「需要真正渲染/远程加载」的限制。

> 其它坑：表单提交失败会**整体重置**（GitLab/凭据/地址全部清空回 Perforce）；后台 tab textContent 跨 eval 不稳定，radio 切换要用 `input[value=XXX].closest(label)` 定位。

## 推荐：前台手动关联

1. 在自己的 Chrome 打开代码库页（上述 URL）
2. 点「关联代码库」
3. 代码库类型选 **GitLab** → 认证方式选 **OAuth**
4. 访问凭据选**本人的 OAuth 凭据**（OAuth 模式下 dropdown 会加载 OAuth 凭据列表）
5. 代码库地址选目标仓库 URL
6. 名称填别名 → 确定
7. 用代码库列表 API 验证关联成功（见下方），记下返回的 `id`（= 流水线 GIT 插件的 `codeRepoId`）

> 关联的具体参数（哪个仓库、哪个凭据、codeRepoId 是多少）属于**项目特定知识**，应记在项目文档（如 `docker/CLAUDE.md`），不放本 skill。

关联成功后，用 [修改流水线配置（PUT）](blueking-api.md) 给流水线加 GIT 插件（`codeRepoId` 填关联返回的 id），比 UI 配置可靠。

## 现有代码库查询（判断是否已关联）

关联前先查是否已存在，避免重复：

```bash
curl -s -X POST "http://localhost:3456/eval?target=<TARGET_ID>" -d '(async () => {
  const r = await fetch("https://devops.devcloud.ztgame.com/ms/repository/api/user/repositories/xcard", { credentials: "include" });
  const j = await r.json();
  return (j.data?.records||[]).map(rec => ({ aliasName: rec.aliasName, url: rec.url, authType: rec.authType }));
})()'
```
