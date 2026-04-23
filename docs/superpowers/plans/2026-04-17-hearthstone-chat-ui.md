# 炉石传说风格聊天界面实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Exporter Chat 前端重构为炉石传说风格的三栏游戏化界面

**Architecture:** 单文件重构（index.html + app.js + style.css），保持零依赖。状态机驱动，CSS 动画实现卡牌效果。

**Tech Stack:** 原生 HTML/JS/CSS, CSS 3D 变换, IntersectionObserver

---

## 文件结构

```
chatui/
├── index.html          # 三栏布局骨架（重构）
├── app.js              # 状态机 + 组件逻辑（重构）
└── style.css           # 炉石主题样式 + 动画（重构）
```

**原则：** 保持三个文件，不增加目录复杂度。当前代码已足够聚焦，无需拆分。

---

## Task 1: 三栏布局骨架（index.html）

**Files:**
- Modify: `lvan/cmd/exporter/chat/chatui/index.html`

- [ ] **Step 1: 重写 HTML 为三栏布局**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Exporter Chat Agent</title>
    <link rel="stylesheet" href="/chat/style.css">
</head>
<body>
    <div id="app">
        <!-- 顶部工具栏 -->
        <header id="toolbar">
            <select id="model-select"><option value="">加载中...</option></select>
            <div class="toolbar-actions">
                <button id="btn-clear" title="清空历史">清空历史</button>
                <a href="/chat/help" target="_blank"><button id="btn-help" title="帮助">帮助</button></a>
            </div>
        </header>

        <!-- 主内容区：战场 + 导航 -->
        <div id="main-content">
            <!-- 战场区（滚动长文档） -->
            <main id="battlefield">
                <!-- 动态插入卡牌问答 -->
            </main>

            <!-- 右侧导航区 -->
            <aside id="sidebar">
                <div id="sidebar-toggle">
                    <button id="btn-graveyard" class="active">🪦 墓地</button>
                    <button id="btn-mindmap">🗺️ 导图</button>
                </div>
                <div id="graveyard-panel" class="sidebar-panel active">
                    <!-- 时间轴列表 -->
                </div>
                <div id="mindmap-panel" class="sidebar-panel">
                    <!-- 导图画布 -->
                    <canvas id="mindmap-canvas"></canvas>
                </div>
            </aside>
        </div>

        <!-- 手牌区 -->
        <div id="hand-area">
            <div id="hand-cards">
                <!-- pending 卡牌 -->
            </div>
            <div id="input-area">
                <textarea id="input" rows="1" placeholder="输入消息... (Enter 发送, Shift+Enter 换行)"></textarea>
                <button id="btn-send">发送</button>
            </div>
        </div>
    </div>
    <script src="/chat/app.js"></script>
</body>
</html>
```

- [ ] **Step 2: 验证 HTML 结构**

在浏览器打开 `http://localhost:18080/chat/`，确认三栏布局渲染正确。

- [ ] **Step 3: Commit**

```bash
git add lvan/cmd/exporter/chat/chatui/index.html
git commit -m "feat(lvan/chat/ui): restructure HTML to three-column hearthstone layout"
```

---

## Task 2: 炉石主题样式（style.css）

**Files:**
- Modify: `lvan/cmd/exporter/chat/chatui/style.css`

- [ ] **Step 1: CSS 变量和基础布局**

```css
:root {
    /* 炉石主题色 */
    --hs-bg-primary: #1a1a2e;
    --hs-bg-secondary: #16213e;
    --hs-bg-dark: #0d1b2a;
    --hs-border: #0f3460;
    --hs-gold: #ffd700;
    --hs-gold-glow: rgba(255, 215, 0, 0.4);
    --hs-green: #4CAF50;
    --hs-red: #e94560;
    --hs-text: #e0e0e0;
    --hs-text-dim: #888;
    --hs-card-width: 120px;
    --hs-card-height: 160px;
}

* { margin: 0; padding: 0; box-sizing: border-box; }

body {
    font-family: 'Georgia', 'Times New Roman', serif;
    background: var(--hs-bg-primary);
    color: var(--hs-text);
    line-height: 1.5;
    overflow: hidden;
}

#app {
    display: flex;
    flex-direction: column;
    height: 100vh;
}
```

- [ ] **Step 2: 三栏布局样式**

```css
/* 顶部工具栏 */
#toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    background: var(--hs-bg-secondary);
    padding: 8px 16px;
    border-bottom: 2px solid var(--hs-border);
    height: 48px;
    flex-shrink: 0;
}

.toolbar-actions { display: flex; gap: 8px; margin-left: auto; }

/* 主内容区 */
#main-content {
    display: flex;
    flex: 1;
    overflow: hidden;
}

/* 战场区 */
#battlefield {
    flex: 1;
    overflow-y: auto;
    padding: 20px;
    scroll-behavior: smooth;
}

/* 右侧导航区 */
#sidebar {
    width: 240px;
    background: var(--hs-bg-secondary);
    border-left: 2px solid var(--hs-border);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
}

#sidebar-toggle {
    display: flex;
    border-bottom: 1px solid var(--hs-border);
}

#sidebar-toggle button {
    flex: 1;
    padding: 10px;
    background: transparent;
    border: none;
    color: var(--hs-text-dim);
    cursor: pointer;
    font-size: 14px;
}

#sidebar-toggle button.active {
    color: var(--hs-gold);
    background: var(--hs-bg-dark);
}

.sidebar-panel {
    display: none;
    flex: 1;
    overflow-y: auto;
    padding: 12px;
}

.sidebar-panel.active { display: block; }

/* 手牌区 */
#hand-area {
    background: var(--hs-bg-secondary);
    border-top: 2px solid var(--hs-border);
    padding: 12px 16px;
    flex-shrink: 0;
}

#hand-cards {
    display: flex;
    gap: 12px;
    margin-bottom: 12px;
    min-height: var(--hs-card-height);
    align-items: flex-end;
    overflow-x: auto;
    padding-bottom: 4px;
}

#input-area {
    display: flex;
    gap: 8px;
}

#input {
    flex: 1;
    background: var(--hs-bg-dark);
    border: 1px solid var(--hs-border);
    border-radius: 4px;
    color: var(--hs-text);
    padding: 8px;
    resize: none;
    font-size: 14px;
    font-family: inherit;
    max-height: 120px;
    min-height: 36px;
}

#input:focus { outline: none; border-color: var(--hs-red); }

#btn-send {
    background: var(--hs-red);
    color: white;
    border: none;
    border-radius: 4px;
    padding: 8px 20px;
    cursor: pointer;
    font-weight: bold;
    min-width: 64px;
}

#btn-send:hover { background: #ff6b6b; }
#btn-send:disabled { background: #555; cursor: not-allowed; opacity: 0.6; }
#btn-send.sending { background: #ff6b6b; }
```

- [ ] **Step 3: 卡牌样式**

```css
/* 卡牌基础 */
.hs-card {
    width: var(--hs-card-width);
    height: var(--hs-card-height);
    background: linear-gradient(135deg, #2a2a4a 0%, #1a1a2e 100%);
    border: 2px solid #555;
    border-radius: 8px;
    padding: 8px;
    display: flex;
    flex-direction: column;
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;
    position: relative;
    flex-shrink: 0;
}

.hs-card:hover {
    transform: translateY(-8px) rotateY(5deg);
    box-shadow: 0 8px 24px rgba(0,0,0,0.4);
}

.hs-card .card-icon {
    font-size: 24px;
    text-align: center;
    margin-bottom: 8px;
}

.hs-card .card-content {
    flex: 1;
    font-size: 12px;
    line-height: 1.4;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 4;
    -webkit-box-orient: vertical;
    word-break: break-all;
}

.hs-card .card-meta {
    font-size: 10px;
    color: var(--hs-text-dim);
    margin-top: 4px;
    text-align: right;
}

/* pending 状态 */
.hs-card.pending {
    border-color: #666;
    opacity: 0.8;
}

/* answering 状态 */
.hs-card.answering {
    border-color: var(--hs-gold);
    animation: cardGlow 2s ease-in-out infinite;
}

/* answered 状态 */
.hs-card.answered {
    border-color: var(--hs-green);
}

@keyframes cardGlow {
    0%, 100% { box-shadow: 0 0 5px var(--hs-gold-glow); }
    50% { box-shadow: 0 0 20px var(--hs-gold-glow), 0 0 40px var(--hs-gold-glow); }
}
```

- [ ] **Step 4: 战场区卡牌展示样式**

```css
/* 战场区中的卡牌（大卡片） */
.battlefield-card {
    background: linear-gradient(135deg, #2a2a4a 0%, #1a1a2e 100%);
    border: 2px solid var(--hs-border);
    border-radius: 12px;
    padding: 20px;
    margin-bottom: 20px;
    min-height: 200px;
}

.battlefield-card.answering {
    border-color: var(--hs-gold);
    animation: cardGlow 2s ease-in-out infinite;
}

.battlefield-card.answered {
    border-color: var(--hs-green);
}

.battlefield-card .card-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--hs-border);
}

.battlefield-card .card-header .card-number {
    background: var(--hs-gold);
    color: var(--hs-bg-primary);
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
    font-size: 14px;
}

.battlefield-card .card-header .card-question {
    flex: 1;
    font-size: 16px;
    font-weight: bold;
}

.battlefield-card .card-header .card-status {
    font-size: 12px;
    color: var(--hs-text-dim);
}

.battlefield-card .card-answer {
    line-height: 1.6;
}

.battlefield-card .card-answer code {
    background: var(--hs-bg-dark);
    padding: 2px 6px;
    border-radius: 3px;
    font-family: 'Consolas', monospace;
    font-size: 0.9em;
}

.battlefield-card .card-answer pre {
    background: var(--hs-bg-dark);
    padding: 12px;
    border-radius: 6px;
    overflow-x: auto;
    margin: 8px 0;
}

.battlefield-card .next-card-hint {
    text-align: center;
    padding: 20px;
    color: var(--hs-text-dim);
    font-style: italic;
    border-top: 1px dashed var(--hs-border);
    margin-top: 16px;
}
```

- [ ] **Step 5: 墓地导航区样式**

```css
/* 墓地时间轴 */
.graveyard-item {
    background: var(--hs-bg-dark);
    border: 1px solid var(--hs-border);
    border-radius: 6px;
    padding: 10px;
    margin-bottom: 8px;
    cursor: pointer;
    transition: border-color 0.2s;
}

.graveyard-item:hover {
    border-color: var(--hs-gold);
}

.graveyard-item .gy-question {
    font-size: 13px;
    font-weight: bold;
    margin-bottom: 4px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.graveyard-item .gy-answer {
    font-size: 11px;
    color: var(--hs-text-dim);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.graveyard-item .gy-time {
    font-size: 10px;
    color: #666;
    margin-top: 4px;
    text-align: right;
}

/* 导图画布 */
#mindmap-canvas {
    width: 100%;
    height: 100%;
    cursor: grab;
}

#mindmap-canvas:active {
    cursor: grabbing;
}
```

- [ ] **Step 6: 滚动条和工具提示**

```css
/* 滚动条 */
::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: var(--hs-bg-dark); }
::-webkit-scrollbar-thumb { background: var(--hs-border); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: #1a4a7a; }

/* 错误提示 */
.error-toast {
    position: fixed;
    top: 20px;
    left: 50%;
    transform: translateX(-50%);
    background: #ff4444;
    color: white;
    padding: 10px 20px;
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
    z-index: 1000;
    animation: slideDown 0.3s ease;
}

@keyframes slideDown {
    from { opacity: 0; transform: translateX(-50%) translateY(-20px); }
    to { opacity: 1; transform: translateX(-50%) translateY(0); }
}

/* 工具调用 */
.tool-call {
    background: var(--hs-bg-dark);
    border-left: 3px solid var(--hs-red);
    padding: 8px 12px;
    margin: 8px 0;
    border-radius: 4px;
}

.tool-call summary {
    cursor: pointer;
    color: var(--hs-red);
    font-weight: bold;
    user-select: none;
}
```

- [ ] **Step 7: 在浏览器中验证样式**

打开 `http://localhost:18080/chat/`，确认：
- 三栏布局正确
- 卡牌样式正常
- 颜色主题统一

- [ ] **Step 8: Commit**

```bash
git add lvan/cmd/exporter/chat/chatui/style.css
git commit -m "feat(lvan/chat/ui): add hearthstone theme styles with card animations"
```

---

## Task 3: 状态机和卡牌逻辑（app.js）

**Files:**
- Modify: `lvan/cmd/exporter/chat/chatui/app.js`

- [ ] **Step 1: Card 类定义**

```javascript
// Card 类
class Card {
    constructor(content) {
        this.id = generateUUID();
        this.content = content;
        this.status = 'pending'; // pending | answering | answered
        this.parentId = null;
        this.inferredParent = null;
        this.related = [];
        this.answer = '';
        this.toolCalls = [];
        this.timestamp = Date.now();
        this.answeredAt = null;
        this.cardNumber = 0; // 分配序号
    }

    addAnswer(text) {
        this.answer += text;
    }

    setAnswered() {
        this.status = 'answered';
        this.answeredAt = Date.now();
    }

    setAnswering() {
        this.status = 'answering';
    }

    addToolCall(tool) {
        this.toolCalls.push(tool);
    }

    // 获取合并后的关联列表
    getRelated() {
        const related = new Set();
        if (this.parentId) related.add(this.parentId);
        if (this.inferredParent) related.add(this.inferredParent);
        return Array.from(related);
    }
}
```

- [ ] **Step 2: State 状态管理**

```javascript
// 全局状态
const State = {
    sessionID: localStorage.getItem('chat_session_id') || generateUUID(),
    currentModel: '',
    abortController: null,
    isSending: false,
    cards: [],           // 所有卡牌
    currentCardId: null, // 当前战场区卡牌
    nextCardNumber: 1,   // 下一个卡牌序号
    navMode: 'graveyard', // 'graveyard' | 'mindmap'

    // 获取 pending 卡牌（手牌区）
    getPendingCards() {
        return this.cards.filter(c => c.status === 'pending');
    },

    // 获取 answering 卡牌
    getAnsweringCard() {
        return this.cards.find(c => c.status === 'answering');
    },

    // 获取 answered 卡牌（按时间排序）
    getAnsweredCards() {
        return this.cards.filter(c => c.status === 'answered')
            .sort((a, b) => a.answeredAt - b.answeredAt);
    },

    // 获取当前战场区卡牌（answering 或最新的 answered）
    getCurrentBattlefieldCard() {
        const answering = this.getAnsweringCard();
        if (answering) return answering;
        const answered = this.getAnsweredCards();
        return answered[answered.length - 1] || null;
    },

    // 添加新卡牌
    addCard(content) {
        const card = new Card(content);
        card.cardNumber = this.nextCardNumber++;
        this.cards.push(card);
        this.save();
        return card;
    },

    // 出牌（pending → answering）
    playCard(cardId) {
        const card = this.cards.find(c => c.id === cardId);
        if (!card || card.status !== 'pending') return null;

        // 如果有正在回答的，先标记为 answered
        const current = this.getAnsweringCard();
        if (current) {
            current.setAnswered();
        }

        card.setAnswering();
        this.currentCardId = card.id;
        this.save();
        return card;
    },

    // 完成当前卡牌
    completeCurrentCard() {
        const card = this.getAnsweringCard();
        if (card) {
            card.setAnswered();
            this.currentCardId = null;
            this.save();
        }
        return card;
    },

    // 自动出牌（手牌区最左侧）
    autoPlay() {
        const pending = this.getPendingCards();
        if (pending.length === 0) return null;
        return this.playCard(pending[0].id);
    },

    // 保存到 localStorage
    save() {
        localStorage.setItem('hearthstone_chat_state', JSON.stringify({
            sessionID: this.sessionID,
            cards: this.cards,
            nextCardNumber: this.nextCardNumber,
            navMode: this.navMode,
        }));
    },

    // 从 localStorage 加载
    load() {
        const saved = localStorage.getItem('hearthstone_chat_state');
        if (saved) {
            const data = JSON.parse(saved);
            this.sessionID = data.sessionID || this.sessionID;
            this.nextCardNumber = data.nextCardNumber || 1;
            this.navMode = data.navMode || 'graveyard';
            if (data.cards) {
                this.cards = data.cards.map(c => Object.assign(new Card(''), c));
            }
        }
    },

    // 清空
    clear() {
        this.cards = [];
        this.currentCardId = null;
        this.nextCardNumber = 1;
        this.save();
    }
};

// 初始化加载
State.load();
localStorage.setItem('chat_session_id', State.sessionID);
```

- [ ] **Step 3: DOM 渲染函数**

```javascript
// 渲染手牌区
function renderHand() {
    const container = document.getElementById('hand-cards');
    container.innerHTML = '';

    const pending = State.getPendingCards();
    pending.forEach(card => {
        const el = createCardElement(card);
        el.addEventListener('click', () => {
            // 插队：移到最前
            const idx = State.cards.findIndex(c => c.id === card.id);
            if (idx > 0) {
                State.cards.splice(idx, 1);
                State.cards.unshift(card);
                State.save();
                renderHand();
            }
            // 立即出牌
            playCard(card.id);
        });
        container.appendChild(el);
    });
}

// 创建卡牌 DOM 元素
function createCardElement(card, isSmall = true) {
    const el = document.createElement('div');
    el.className = `hs-card ${card.status}`;
    el.dataset.cardId = card.id;

    const icon = card.status === 'pending' ? '💬' :
                 card.status === 'answering' ? '⚡' : '✓';

    const content = card.content.substring(0, 20) + (card.content.length > 20 ? '...' : '');
    const time = new Date(card.timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });

    if (isSmall) {
        el.innerHTML = `
            <div class="card-icon">${icon}</div>
            <div class="card-content">${escapeHtml(content)}</div>
            <div class="card-meta">#${card.cardNumber} ${time}</div>
        `;
    }

    return el;
}

// 渲染战场区
function renderBattlefield() {
    const container = document.getElementById('battlefield');
    container.innerHTML = '';

    // 渲染所有 answered 卡牌（按顺序）
    const answered = State.getAnsweredCards();
    answered.forEach(card => {
        container.appendChild(createBattlefieldCard(card));
    });

    // 渲染 answering 卡牌
    const answering = State.getAnsweringCard();
    if (answering) {
        container.appendChild(createBattlefieldCard(answering, true));
    }

    // 滚动到 answering 卡牌或底部
    if (answering) {
        const el = container.querySelector('.battlefield-card.answering');
        if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    } else {
        scrollToBottom();
    }
}

// 创建战场区卡牌元素
function createBattlefieldCard(card, isActive = false) {
    const el = document.createElement('div');
    el.className = `battlefield-card ${card.status}`;
    el.dataset.cardId = card.id;

    const statusText = card.status === 'answering' ? '回答中...' :
                       card.status === 'answered' ? '已完成' : '等待中';

    let html = `
        <div class="card-header">
            <div class="card-number">${card.cardNumber}</div>
            <div class="card-question">${escapeHtml(card.content)}</div>
            <div class="card-status">${statusText}</div>
        </div>
    `;

    if (card.answer || card.toolCalls.length > 0) {
        html += `<div class="card-answer">${renderMarkdown(card.answer)}</div>`;

        // 工具调用
        card.toolCalls.forEach(tool => {
            html += `
                <div class="tool-call">
                    <details ${tool.open ? 'open' : ''}>
                        <summary>🔧 ${tool.name}</summary>
                        <pre>${escapeHtml(JSON.stringify(tool.result, null, 2))}</pre>
                    </details>
                </div>
            `;
        });
    }

    // 如果是 answered 且后面还有 pending，显示提示
    if (card.status === 'answered' && !isActive && State.getPendingCards().length > 0) {
        html += `<div class="next-card-hint">▼ 向下滚动查看下一题 ▼</div>`;
    }

    el.innerHTML = html;
    return el;
}

// 渲染墓地导航区
function renderGraveyard() {
    const container = document.getElementById('graveyard-panel');
    container.innerHTML = '';

    const answered = State.getAnsweredCards();
    answered.forEach(card => {
        const el = document.createElement('div');
        el.className = 'graveyard-item';
        el.dataset.cardId = card.id;

        const answerPreview = card.answer.substring(0, 30) + (card.answer.length > 30 ? '...' : '');
        const time = new Date(card.answeredAt).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });

        el.innerHTML = `
            <div class="gy-question">#${card.cardNumber} ${escapeHtml(card.content.substring(0, 20))}</div>
            <div class="gy-answer">${escapeHtml(answerPreview)}</div>
            <div class="gy-time">${time}</div>
        `;

        el.addEventListener('click', () => {
            // 滚动到对应的战场区卡牌
            const bfCard = document.querySelector(`.battlefield-card[data-card-id="${card.id}"]`);
            if (bfCard) bfCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
        });

        container.appendChild(el);
    });
}
```

- [ ] **Step 4: 出牌和 SSE 逻辑**

```javascript
// 出牌并开始回答
async function playCard(cardId) {
    const card = State.playCard(cardId);
    if (!card) return;

    renderHand();
    renderBattlefield();
    renderGraveyard();

    // 发送 SSE 请求
    await sendCardAnswer(card);
}

// 发送卡牌回答（SSE）
async function sendCardAnswer(card) {
    State.abortController = new AbortController();
    State.isSending = true;
    updateSendButton();

    try {
        const resp = await fetch('/chat/api/message', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                content: card.content,
                model: State.currentModel,
                session_id: State.sessionID,
            }),
            signal: State.abortController.signal,
        });

        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);

        const reader = resp.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        while (true) {
            const { done, value } = await reader.read();
            if (done) break;

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');
            buffer = lines.pop();

            let currentEvent = '';
            for (const line of lines) {
                if (line.startsWith('event: ')) {
                    currentEvent = line.slice(7);
                } else if (line.startsWith('data: ')) {
                    try {
                        const data = JSON.parse(line.slice(6));
                        handleCardEvent(card, currentEvent, data);
                    } catch(e) {
                        console.error('解析 SSE 失败:', line, e);
                    }
                }
            }
        }
    } catch(e) {
        if (e.name !== 'AbortError') showError('连接错误: ' + e.message);
    } finally {
        // 完成当前卡牌
        State.completeCurrentCard();
        State.isSending = false;
        State.abortController = null;
        updateSendButton();

        // 渲染更新
        renderBattlefield();
        renderGraveyard();

        // 自动出下一张
        setTimeout(() => {
            const next = State.autoPlay();
            if (next) playCard(next.id);
        }, 500); // 短暂延迟，让用户看到完成状态
    }
}

// 处理卡牌 SSE 事件
function handleCardEvent(card, event, data) {
    switch(event) {
        case 'content':
            card.addAnswer(data.text);
            updateBattlefieldCard(card);
            break;
        case 'tool_use':
            card.addToolCall({ name: data.name, args: data.args, result: null, open: true });
            updateBattlefieldCard(card);
            break;
        case 'tool_result':
            // 更新最后一个工具调用的结果
            const lastTool = card.toolCalls[card.toolCalls.length - 1];
            if (lastTool) {
                lastTool.result = data.result;
                lastTool.success = data.success;
            }
            updateBattlefieldCard(card);
            break;
        case 'error':
            showError(data.message);
            break;
    }
}

// 更新战场区中的单个卡牌（不重新渲染全部）
function updateBattlefieldCard(card) {
    const el = document.querySelector(`.battlefield-card[data-card-id="${card.id}"]`);
    if (!el) return;

    const answerEl = el.querySelector('.card-answer');
    if (answerEl) {
        answerEl.innerHTML = renderMarkdown(card.answer);
    }

    // 滚动到底部（如果用户在底部附近）
    const bf = document.getElementById('battlefield');
    const isNearBottom = bf.scrollHeight - bf.scrollTop - bf.clientHeight < 100;
    if (isNearBottom) {
        scrollToBottom();
    }
}
```

- [ ] **Step 5: 事件绑定和初始化**

```javascript
// 初始化
document.addEventListener('DOMContentLoaded', () => {
    loadConfig();
    bindEvents();
    renderHand();
    renderBattlefield();
    renderGraveyard();
});

// 绑定事件
function bindEvents() {
    const input = document.getElementById('input');
    const btnSend = document.getElementById('btn-send');
    const btnClear = document.getElementById('btn-clear');

    input.addEventListener('keydown', e => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            submitNewCard();
        }
    });

    btnSend.addEventListener('click', () => {
        if (State.isSending) stopSending();
        else submitNewCard();
    });

    btnClear.addEventListener('click', clearAll);

    document.getElementById('model-select').addEventListener('change', e => {
        switchModel(e.target.value);
    });

    // 侧边栏切换
    document.getElementById('btn-graveyard').addEventListener('click', () => {
        setNavMode('graveyard');
    });
    document.getElementById('btn-mindmap').addEventListener('click', () => {
        setNavMode('mindmap');
    });
}

// 提交新卡牌
function submitNewCard() {
    const input = document.getElementById('input');
    const content = input.value.trim();
    if (!content) return;

    input.value = '';
    input.style.height = 'auto';

    const card = State.addCard(content);
    renderHand();

    // 如果没有正在回答的，自动出牌
    if (!State.getAnsweringCard()) {
        playCard(card.id);
    }
}

// 停止发送
function stopSending() {
    if (State.abortController) {
        State.abortController.abort();
    }
}

// 清空所有
async function clearAll() {
    try {
        const resp = await fetch('/chat/api/history?session_id=' + State.sessionID, {
            method: 'DELETE'
        });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);

        State.clear();
        renderHand();
        renderBattlefield();
        renderGraveyard();
    } catch(e) {
        showError('清空失败: ' + e.message);
    }
}

// 切换导航模式
function setNavMode(mode) {
    State.navMode = mode;
    State.save();

    document.getElementById('btn-graveyard').classList.toggle('active', mode === 'graveyard');
    document.getElementById('btn-mindmap').classList.toggle('active', mode === 'mindmap');
    document.getElementById('graveyard-panel').classList.toggle('active', mode === 'graveyard');
    document.getElementById('mindmap-panel').classList.toggle('active', mode === 'mindmap');
}

// 更新发送按钮状态
function updateSendButton() {
    const btn = document.getElementById('btn-send');
    if (State.isSending) {
        btn.textContent = '停止';
        btn.classList.add('sending');
    } else {
        btn.textContent = '发送';
        btn.classList.remove('sending');
    }
}

// 滚动到底部
function scrollToBottom() {
    const el = document.getElementById('battlefield');
    el.scrollTop = el.scrollHeight;
}

// 加载配置（复用原有逻辑）
async function loadConfig() {
    try {
        const resp = await fetch('/chat/api/config');
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const config = await resp.json();
        const select = document.getElementById('model-select');
        select.innerHTML = '';
        config.models.forEach(m => {
            const opt = document.createElement('option');
            opt.value = m;
            opt.textContent = m;
            if (m === config.current) opt.selected = true;
            select.appendChild(opt);
        });
        State.currentModel = config.current;
    } catch(e) {
        showError('加载配置失败: ' + e.message);
    }
}

async function switchModel(model) {
    try {
        const resp = await fetch('/chat/api/config', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ model }),
        });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        State.currentModel = model;
    } catch(e) {
        showError('切换模型失败: ' + e.message);
        document.getElementById('model-select').value = State.currentModel;
    }
}

// 工具函数（复用原有）
function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
        const r = Math.random() * 16 | 0;
        return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function renderMarkdown(text) {
    if (!text) return '';
    text = escapeHtml(text);
    text = text.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>');
    text = text.replace(/`([^`]+)`/g, '<code>$1</code>');
    text = text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    text = text.replace(/^- (.+)$/gm, '<li>$1</li>');
    text = text.replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>');
    text = text.replace(/\n/g, '<br>');
    return text;
}

function showError(msg) {
    const toast = document.createElement('div');
    toast.className = 'error-toast';
    toast.textContent = msg;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 5000);
}

// textarea 自适应高度
document.addEventListener('input', e => {
    if (e.target.id === 'input') {
        e.target.style.height = 'auto';
        e.target.style.height = Math.min(e.target.scrollHeight, 120) + 'px';
    }
});
```

- [ ] **Step 6: 滚动无缝衔接检测**

```javascript
// 在 bindEvents 中添加滚动检测
document.getElementById('battlefield').addEventListener('scroll', handleScroll);

function handleScroll() {
    const bf = document.getElementById('battlefield');
    const isAtBottom = bf.scrollHeight - bf.scrollTop - bf.clientHeight < 50;

    // 如果滚动到底部，且有 answered 卡牌后面跟着 answering
    if (isAtBottom) {
        const answering = State.getAnsweringCard();
        if (answering) {
            // 滚动到 answering 卡牌
            const el = document.querySelector('.battlefield-card.answering');
            if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    }
}
```

- [ ] **Step 7: 在浏览器中测试完整流程**

测试场景：
1. 输入问题1 → 进入手牌区 → 自动出牌 → 战场区显示 answering
2. 输入问题2 → 进入手牌区末尾
3. 等待 AI 回答完问题1 → 状态变为 answered → 自动出问题2
4. 滚动战场区，确认 answered 和 answering 卡牌垂直排列
5. 点击右侧墓地条目，确认跳转到对应战场区卡牌

- [ ] **Step 8: Commit**

```bash
git add lvan/cmd/exporter/chat/chatui/app.js
git commit -m "feat(lvan/chat/ui): implement hearthstone card queue with state machine"
```

---

## Task 4: 导图模式（Canvas 实现）

**Files:**
- Modify: `lvan/cmd/exporter/chat/chatui/app.js`

- [ ] **Step 1: 简单力导向图实现**

```javascript
// 导图渲染
function renderMindmap() {
    const canvas = document.getElementById('mindmap-canvas');
    const ctx = canvas.getContext('2d');
    const dpr = window.devicePixelRatio || 1;

    // 设置画布尺寸
    const rect = canvas.parentElement.getBoundingClientRect();
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    ctx.scale(dpr, dpr);

    const width = rect.width;
    const height = rect.height;

    // 获取 answered 卡牌
    const cards = State.getAnsweredCards();
    if (cards.length === 0) {
        ctx.fillStyle = '#666';
        ctx.font = '14px sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText('暂无历史问答', width / 2, height / 2);
        return;
    }

    // 简单网格布局
    const nodeRadius = 20;
    const cols = 2;
    const spacingX = width / cols;
    const spacingY = 80;

    const nodes = cards.map((card, i) => ({
        id: card.id,
        x: (i % cols + 0.5) * spacingX,
        y: Math.floor(i / cols) * spacingY + 40,
        card: card,
        related: card.getRelated(),
    }));

    // 绘制连线
    ctx.strokeStyle = 'rgba(255, 215, 0, 0.3)';
    ctx.lineWidth = 1;
    nodes.forEach(node => {
        node.related.forEach(relatedId => {
            const target = nodes.find(n => n.id === relatedId);
            if (target) {
                ctx.beginPath();
                ctx.moveTo(node.x, node.y);
                ctx.lineTo(target.x, target.y);
                ctx.stroke();
            }
        });
    });

    // 绘制节点
    nodes.forEach(node => {
        // 圆形背景
        ctx.beginPath();
        ctx.arc(node.x, node.y, nodeRadius, 0, Math.PI * 2);
        ctx.fillStyle = '#2a2a4a';
        ctx.fill();
        ctx.strokeStyle = '#ffd700';
        ctx.lineWidth = 2;
        ctx.stroke();

        // 编号
        ctx.fillStyle = '#ffd700';
        ctx.font = 'bold 12px sans-serif';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText('#' + node.card.cardNumber, node.x, node.y);

        // 问题预览（下方）
        ctx.fillStyle = '#aaa';
        ctx.font = '10px sans-serif';
        const preview = node.card.content.substring(0, 8) + '...';
        ctx.fillText(preview, node.x, node.y + nodeRadius + 12);
    });

    // 点击检测
    canvas.onclick = e => {
        const rect = canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        nodes.forEach(node => {
            const dx = x - node.x;
            const dy = y - node.y;
            if (dx * dx + dy * dy < nodeRadius * nodeRadius) {
                // 跳转到对应战场区卡牌
                const bfCard = document.querySelector(`.battlefield-card[data-card-id="${node.id}"]`);
                if (bfCard) bfCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }
        });
    };
}
```

- [ ] **Step 2: 在 setNavMode 中调用导图渲染**

```javascript
function setNavMode(mode) {
    State.navMode = mode;
    State.save();

    document.getElementById('btn-graveyard').classList.toggle('active', mode === 'graveyard');
    document.getElementById('btn-mindmap').classList.toggle('active', mode === 'mindmap');
    document.getElementById('graveyard-panel').classList.toggle('active', mode === 'graveyard');
    document.getElementById('mindmap-panel').classList.toggle('active', mode === 'mindmap');

    if (mode === 'mindmap') {
        renderMindmap();
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add lvan/cmd/exporter/chat/chatui/app.js
git commit -m "feat(lvan/chat/ui): add mindmap canvas with simple grid layout"
```

---

## Task 5: 关联标记功能（右键菜单）

**Files:**
- Modify: `lvan/cmd/exporter/chat/chatui/app.js`
- Modify: `lvan/cmd/exporter/chat/chatui/style.css`

- [ ] **Step 1: 右键菜单实现**

```javascript
// 在 createCardElement 中添加右键菜单
function createCardElement(card, isSmall = true) {
    const el = document.createElement('div');
    el.className = `hs-card ${card.status}`;
    el.dataset.cardId = card.id;

    // ... 原有内容 ...

    // 右键菜单
    el.addEventListener('contextmenu', e => {
        e.preventDefault();
        showContextMenu(e, card);
    });

    return el;
}

// 显示上下文菜单
function showContextMenu(e, card) {
    // 移除旧菜单
    document.querySelectorAll('.context-menu').forEach(m => m.remove());

    const menu = document.createElement('div');
    menu.className = 'context-menu';
    menu.style.left = e.pageX + 'px';
    menu.style.top = e.pageY + 'px';

    const answeredCards = State.getAnsweredCards();
    const canMarkParent = answeredCards.length > 0 && card.status === 'pending';

    menu.innerHTML = `
        <div class="context-menu-item" data-action="top">置顶</div>
        ${canMarkParent ? `
            <div class="context-menu-divider"></div>
            <div class="context-menu-label">标记为追问：</div>
            ${answeredCards.map(c => `
                <div class="context-menu-item" data-action="parent" data-parent-id="${c.id}">
                    #${c.cardNumber} ${escapeHtml(c.content.substring(0, 15))}...
                </div>
            `).join('')}
        ` : ''}
        <div class="context-menu-divider"></div>
        <div class="context-menu-item danger" data-action="delete">删除</div>
    `;

    menu.addEventListener('click', ev => {
        const item = ev.target.closest('.context-menu-item');
        if (!item) return;

        const action = item.dataset.action;
        if (action === 'top') {
            // 移到队列最前
            const idx = State.cards.findIndex(c => c.id === card.id);
            if (idx > 0) {
                State.cards.splice(idx, 1);
                State.cards.unshift(card);
                State.save();
                renderHand();
            }
        } else if (action === 'parent') {
            card.parentId = item.dataset.parentId;
            State.save();
            renderHand();
        } else if (action === 'delete') {
            State.cards = State.cards.filter(c => c.id !== card.id);
            State.save();
            renderHand();
        }

        menu.remove();
    });

    document.body.appendChild(menu);

    // 点击其他地方关闭
    setTimeout(() => {
        document.addEventListener('click', function closeMenu() {
            menu.remove();
            document.removeEventListener('click', closeMenu);
        }, { once: true });
    }, 0);
}
```

- [ ] **Step 2: 右键菜单样式**

```css
/* 右键菜单 */
.context-menu {
    position: absolute;
    background: var(--hs-bg-secondary);
    border: 1px solid var(--hs-border);
    border-radius: 6px;
    padding: 4px 0;
    min-width: 160px;
    z-index: 1000;
    box-shadow: 0 4px 12px rgba(0,0,0,0.4);
}

.context-menu-item {
    padding: 8px 16px;
    cursor: pointer;
    font-size: 13px;
    color: var(--hs-text);
    transition: background 0.15s;
}

.context-menu-item:hover {
    background: var(--hs-bg-dark);
}

.context-menu-item.danger {
    color: #ff4444;
}

.context-menu-item.danger:hover {
    background: rgba(255, 68, 68, 0.1);
}

.context-menu-divider {
    height: 1px;
    background: var(--hs-border);
    margin: 4px 0;
}

.context-menu-label {
    padding: 4px 16px;
    font-size: 11px;
    color: var(--hs-text-dim);
    text-transform: uppercase;
}
```

- [ ] **Step 3: Commit**

```bash
git add lvan/cmd/exporter/chat/chatui/app.js lvan/cmd/exporter/chat/chatui/style.css
git commit -m "feat(lvan/chat/ui): add right-click context menu for card relations"
```

---

## Self-Review

### Spec Coverage

| 设计文档需求 | 对应 Task |
|-------------|----------|
| 三栏布局 | Task 1 |
| 卡牌视觉样式 | Task 2 |
| 手牌区 pending 队列 | Task 3 |
| 战场区 answering/answered | Task 3 |
| 自动出牌机制 | Task 3 |
| 滚动无缝衔接 | Task 3 (handleScroll) |
| 墓地导航区 | Task 3 |
| 导图模式 | Task 4 |
| 关联标记 | Task 5 |
| 右键菜单 | Task 5 |

### Placeholder Scan

无 TBD/TODO。所有步骤包含完整代码。

### Type Consistency

- `Card` 类属性与 State 方法一致
- `status` 枚举值统一：'pending' | 'answering' | 'answered'
- `navMode` 枚举值统一：'graveyard' | 'mindmap'

---

## 执行方式选择

Plan complete and saved to `docs/superpowers/plans/2026-04-17-hearthstone-chat-ui.md`.

**Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
