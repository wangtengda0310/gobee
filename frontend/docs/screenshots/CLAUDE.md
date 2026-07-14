# rain-qa-func 截图文档索引

本目录包含 rain-qa-func 前端各页面的截图和标注文档。

## 截图列表

| 页面 | 截图 | 标注 | 说明 |
|------|------|------|------|
| 战斗测试 | [overview.png](./FunctionTestEditor/overview.png) | [annotations.md](./FunctionTestEditor/annotations.md) | 用例配置和执行 |
| 配表测试 | [overview.png](./ExcelTestEditor/overview.png) | [annotations.md](./ExcelTestEditor/annotations.md) | Excel 检查配置 |
| Wiki检查 | [overview.png](./HeroWikiCheck/overview.png) | [annotations.md](./HeroWikiCheck/annotations.md) | 武将差异检查 |
| 语音检查 | [overview.png](./HeroVoiceResourceCheck/overview.png) | [annotations.md](./HeroVoiceResourceCheck/annotations.md) | 资源检查 |
| 首页 | [overview.png](./Home/overview.png) | [annotations.md](./Home/annotations.md) | 应用首页 |

## 截图方式

使用 Playwright MCP 工具截取：

1. 启动 `wails3 dev` 运行应用
2. 使用 `browser_navigate` 导航到 `http://localhost:9245`
3. 点击导航按钮切换页面
4. 使用 `browser_take_screenshot` 截图

## 维护说明

当页面布局发生重大变更时，请更新截图：

```bash
# 1. 启动应用
wails3 dev

# 2. 使用 Playwright 重新截图
# 或手动截图后放入对应目录
```

## 截图命名规范

| 文件名 | 说明 |
|--------|------|
| overview.png | 页面整体截图 |
| detail-*.png | 局部细节截图（可选） |

## 标注文档内容

每个 annotations.md 包含：

1. **整体布局图** - ASCII 坐标参考
2. **区域标注表** - 位置、组件、交互说明
3. **组件详解** - 关键区域的详细布局
4. **交互说明** - 用户操作说明
5. **数据结构** - 相关数据类型（如适用）
