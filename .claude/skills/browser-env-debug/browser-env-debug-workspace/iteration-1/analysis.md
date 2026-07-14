# browser-env-debug Skill 模拟测试分析报告

## 测试概述

3 个浏览器环境 bug 场景，每个场景分别用 with-skill 和 without-skill 运行，对比排查质量。

## 逐场景对比

### Eval 1: Monaco Editor + Vue 3 (getModel 返回 undefined)

| 维度 | With Skill | Without Skill |
|------|-----------|---------------|
| 结构化程度 | 4步流程，清晰的决策树 | 6个排查方向，但无优先级排序 |
| Vue Proxy 识别 | **命中**，列为第一嫌疑 | **完全遗漏**，未提及 reactive 干扰 |
| Vite 缓存识别 | 命中，列为第二嫌疑 | 未提及 .vite 缓存问题 |
| 隔离测试建议 | page.evaluate 隔离法 | 未建议隔离，建议"最小化实验" |
| Monaco 专项知识 | 较少，泛化指导 | 较多（Worker加载、DOM时机、容器尺寸） |
| 修复方案 | markRaw/shallowRef | 未给出明确修复（因为方向不对） |

**结论**: skill 在识别 Vue Proxy 干扰方面有显著优势，这是本次测试最关键的发现。without-skill 完全遗漏了这个高频陷阱。

### Eval 2: pdf.js + Vite build (workerSrc undefined)

| 维度 | With Skill | Without Skill |
|------|-----------|---------------|
| 核心问题定位 | 正确：dev vs build 模块解析差异 | 正确：同上 |
| 结构化程度 | 4步流程 | 详细排查步骤+代码示例 |
| pdf.js 专项知识 | 泛化指导，缺少版本差异说明 | 详尽：3.x vs 4.x、ESM/CJS、具体 import 路径 |
| 修复方案 | 方向正确但不够具体 | 多种方案+代码示例 |
| Worker 处理 | 提及但不够深入 | ?url 后缀、public 目录、CDN 等多种方案 |

**结论**: 两者都正确定位了核心问题。without-skill 在 pdf.js 专项知识上更优。skill 的方法论框架在这里价值有限。

### Eval 3: PeerJS + React (WebSocket 偶发超时)

| 维度 | With Skill | Without Skill |
|------|-----------|---------------|
| 核因定位 | 正确：生命周期不匹配 | 正确：同上 |
| React StrictMode | 提及双重渲染 | 提及 double invoke |
| 闭包陷阱 | 提及 stale closure | 提及闭包捕获旧值 |
| useRef vs useState | 明确建议 | 明确建议 |
| 网络层排查 | 较少 | 更深入（DNS、TLS、HAR分析） |
| 回归验证 | Playwright 自动化 | 手动步骤为主 |

**结论**: 两者表现接近，都正确定位了根因。without-skill 在网络层排查更深入，with-skill 在回归验证方案上更优。

## 综合评估

### Skill 的核心价值

1. **Vue Proxy 陷阱识别**（Eval 1 显著优势）— 这是最容易遗漏的陷阱
2. **page.evaluate 隔离测试方法论** — 提供了系统化的第一步操作
3. **结构化排查流程** — 避免无头苍蝇式排查

### Skill 的改进方向

1. **陷阱清单需要更深** — 目前 5 个陷阱只给了框架级描述，缺少框架特定指导
2. **缺少 React 专项陷阱** — React StrictMode、闭包陷阱、useRef vs useState 等未覆盖
3. **Vite build vs dev 差异** — 需要更明确的排查指引（不仅是缓存，还有 esbuild vs Rollup）
4. **缺少"排查优先级排序"指导** — 应该帮助 AI 判断先排查哪个方向

### 建议的 Skill 优化

1. 增加 React 陷阱（StrictMode 双重渲染、闭包陷阱、生命周期不匹配）
2. 增加 Vite dev vs build 差异排查指引
3. 增加排查优先级决策树（根据症状特征快速缩小范围）
4. 考虑增加框架特定参考文件（vue-traps.md, react-traps.md, vite-traps.md）
