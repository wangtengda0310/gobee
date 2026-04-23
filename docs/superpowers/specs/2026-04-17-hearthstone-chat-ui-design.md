# 炉石传说风格聊天界面设计文档

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 Exporter Chat 界面重构为炉石传说风格的游戏化界面，包含卡牌队列、战场区、双模式导航区。

**Architecture:** 前端纯 HTML/CSS/JS，无框架依赖。状态机驱动 UI 更新，CSS 动画实现卡牌效果。

**Tech Stack:** 原生 HTML/JS/CSS, CSS 3D 变换, IntersectionObserver, MutationObserver

---

## 需求总结

- **卡牌队列（手牌区）**：用户提交的问题以卡牌形式进入底部队列，AI 依次处理
- **战场区**：当前正在回答的问题和 AI 回复内容展示区域
- **双模式导航区（右侧）**：
  - 墓地模式：时间轴展示已完成问答
  - 导图模式：节点连线展示关联关系
- **关联机制**：用户可标记问题间的父子关系，预留自动推断接口

---

## 布局设计

### 三栏布局

```
┌────────────────────────────────────────┬─────────────┐
│ [模型▼] [清空] [帮助]                  │  🔄 切换    │
├────────────────────────────────────────┤  墓地/导图  │
│                                        │             │
│   战场区（answering / answered）        │  ┌───┐     │
│   ┌────────────────────────────┐      │  │A1 │     │
│   │ Q2: 如何查看日志？  ✓       │      │  ├───┤     │
│   │ ──────────────────────      │      │  │A2 │     │
│   │ AI: 你可以使用 run_shell... │      │  └───┘     │
│   │ [下一题 ▶] 或自动切换...    │      │             │
│   └────────────────────────────┘      │  墓地模式    │
│                                        │  (时间轴)   │
├────────────────────────────────────────┤             │
│ 手牌区（pending）                      │             │
│ ┌───┐ ┌───┐ ┌───┐                     │             │
│ │Q3 │ │Q4 │ │Q5 │                     │             │
│ └───┘ └───┘ └───┘                     │             │
├────────────────────────────────────────┤             │
│ [输入消息...] [发送]                   │             │
└────────────────────────────────────────┴─────────────┘
```

### 区域说明

| 区域 | 位置 | 内容 | 状态 |
|------|------|------|------|
| 手牌区 | 底部 | pending 卡牌 | 等待处理 |
| 战场区 | 中间 | answering/answered 卡牌 + AI 回答 | 当前处理/可阅读 |
| 导航区 | 右侧 | answered 卡牌归档 + 关联导图 | 已完成 |

---

## 卡牌设计

### 视觉样式

- **尺寸**：120px × 160px（炉石标准比例 3:4）
- **圆角**：8px
- **边框**：2px solid，根据状态变色
- **阴影**：box-shadow 营造悬浮感
- **3D 效果**：hover 时轻微旋转（rotateY/rotateX）

### 状态样式

| 状态 | 颜色 | 效果 |
|------|------|------|
| pending | 灰色 #666 | 静态 |
| answering | 金色 #ffd700 | 发光动画 pulse |
| answered | 绿色 #4CAF50 | 翻转动画 flip |

### 卡牌内容

```
┌─────────────┐
│  💬         │  <- 类型图标
│             │
│  问题内容    │  <- 截断显示前20字
│  预览...    │
│             │
│  #12 14:32  │  <- 编号 + 时间
└─────────────┘
```

---

## 状态流转

```
用户输入 → 生成卡牌 → pending（手牌区末尾）
                ↓
         点击发送/自动出牌 → answering（战场区）
                ↓
         AI 回答完成 → answered（战场区保留，可阅读）
                ↓
         用户浏览完毕/点击下一张 → 自动切换下一卡牌
                ↓
         手牌区下一张 → answering（战场区）
```

### 自动出牌机制

- **正常流程**：前一卡牌 answered 后，自动将手牌区最左侧卡牌移入战场区开始 answering
- **手动插队**：用户可点击手牌区任意 pending 卡牌，插队到最前优先回答
- **阅读保护**：answered 卡牌不立即移出战场区，新卡牌在后台构建答案，用户浏览完当前回答后自动切换

---

## 数据模型

### Card 结构

```javascript
{
  id: string,           // UUID
  content: string,      // 问题内容
  status: 'pending' | 'answering' | 'answered',
  parentId: string | null,      // 用户显式标记的父问题
  inferredParent: string | null, // 系统推断的关联（预留）
  related: string[],    // 合并后的关联列表
  answer: string,       // AI 回答内容
  toolCalls: ToolCall[], // 工具调用记录
  timestamp: number,    // 创建时间
  answeredAt: number    // 完成时间
}
```

### 关联规则

1. **显式关联**：用户右键卡牌 → "标记为追问" → 选择父问题
2. **推断关联**（预留）：在 answering 期间提交的问题，自动推断为关联
3. **合并规则**：`related = [parentId, inferredParent].filter(Boolean)`

---

## 导航区双模式

### 墓地模式（默认）

- 垂直时间轴列表
- 每张卡牌显示：问题摘要 + 回答摘要 + 时间
- 点击展开完整问答
- 关联卡牌高亮显示

### 导图模式

- 力导向图布局（D3.js 或自研简单实现）
- 节点 = 问答卡牌
- 边 = 关联关系
- 点击节点高亮相关路径
- 支持拖拽、缩放

### 切换按钮

```
┌─────────────┐
│ 🪦 墓地  🗺️ │  <- 切换按钮
└─────────────┘
```

---

## 交互设计

### 手牌区交互

| 操作 | 效果 |
|------|------|
| 点击 pending 卡牌 | 将该卡牌移到战场区（插队） |
| 右键卡牌 | 上下文菜单：标记为追问、删除、置顶 |
| hover | 3D 倾斜效果 |
| 拖拽 | 调整队列顺序 |

### 战场区交互

| 操作 | 效果 |
|------|------|
| 输入框输入 | 正常输入，Enter 发送 |
| 发送按钮 | 新卡牌进入手牌区末尾 |
| 停止按钮 | 中断当前 SSE 流 |

### 导航区交互

| 操作 | 效果 |
|------|------|
| 点击 answered 卡牌 | 在战场区展开完整问答 |
| 点击关联高亮 | 高亮所有 related 卡牌 |
| 切换模式按钮 | 墓地/导图模式切换 |

---

## 动画设计

### 卡牌移动动画

- 手牌区 → 战场区：飞入动画（translateX/Y + scale）
- 战场区 → 墓地：淡出 + 缩小
- 新卡牌加入手牌区：从底部滑入

### 状态变化动画

- pending → answering：金色发光 pulse
- answering → answered：3D 翻转（rotateY 180°）

### CSS 关键帧

```css
@keyframes cardGlow {
  0%, 100% { box-shadow: 0 0 5px #ffd700; }
  50% { box-shadow: 0 0 20px #ffd700, 0 0 40px #ffd700; }
}

@keyframes cardFlip {
  0% { transform: rotateY(0); }
  100% { transform: rotateY(180deg); }
}
```

---

## API 变更

### 后端无需变更

前端通过现有 SSE 接口获取数据，状态管理完全在前端。

### 前端新增本地存储

```javascript
// localStorage 键
const STORAGE_KEY = 'hearthstone_chat_state';

// 保存内容
{
  cards: Card[],
  currentCardId: string | null,
  navMode: 'graveyard' | 'mindmap'
}
```

---

## 文件结构

```
chatui/
├── index.html          # 三栏布局骨架
├── app.js              # 主逻辑 + 状态机
├── components/
│   ├── hand.js         # 手牌区组件
│   ├── battlefield.js  # 战场区组件
│   ├── graveyard.js    # 墓地模式组件
│   ├── mindmap.js      # 导图模式组件
│   └── card.js         # 卡牌通用组件
├── styles/
│   ├── main.css        # 布局 + 变量
│   ├── card.css        # 卡牌样式 + 动画
│   └── themes.css      # 主题色
└── utils/
    ├── state.js        # 状态管理
    ├── storage.js      # localStorage 封装
    └── animator.js     # 动画工具
```

---

## 实现计划

### Task 1: 基础布局重构
- 三栏 HTML 结构
- CSS Grid/Flex 布局
- 响应式适配

### Task 2: 卡牌组件
- Card DOM 结构
- 状态样式
- 3D hover 效果

### Task 3: 手牌区
- pending 卡牌队列
- 点击出牌逻辑
- 拖拽排序

### Task 4: 战场区
- answering 卡牌展示
- AI 回答流式显示
- 自动出牌机制

### Task 5: 墓地导航区
- 时间轴列表
- 点击展开
- 关联高亮

### Task 6: 导图模式
- 节点布局
- 连线绘制
- 拖拽缩放

### Task 7: 状态管理
- State 类
- localStorage 持久化
- 与后端 API 集成

### Task 8: 动画系统
- 卡牌飞入/飞出
- 状态切换动画
- 性能优化

---

## 风险与注意事项

1. **性能**：大量卡牌时 DOM 操作性能，考虑虚拟滚动
2. **移动端**：三栏布局在窄屏下的适配（可能隐藏导航区）
3. **导图模式**：自研简单实现还是引入 D3.js（增加体积）
4. **状态同步**：多标签页同时打开时的状态一致性
