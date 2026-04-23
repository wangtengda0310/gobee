// ============================================
// 炉石传说风格聊天界面 - 状态机驱动
// 文件: lvan/cmd/exporter/chat/chatui/app.js
// ============================================

// ---------- Card 类 ----------
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
        this.cardNumber = 0;
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

// ---------- 全局状态 ----------
const State = {
    sessionID: localStorage.getItem('chat_session_id') || generateUUID(),
    currentModel: '',
    abortController: null,
    isSending: false,
    cards: [],
    currentCardId: null,
    nextCardNumber: 1,
    navMode: 'graveyard',

    getPendingCards() {
        return this.cards.filter(c => c.status === 'pending');
    },

    getAnsweringCard() {
        return this.cards.find(c => c.status === 'answering');
    },

    getAnsweredCards() {
        return this.cards.filter(c => c.status === 'answered')
            .sort((a, b) => a.answeredAt - b.answeredAt);
    },

    addCard(content) {
        const card = new Card(content);
        card.cardNumber = this.nextCardNumber++;
        this.cards.push(card);
        this.save();
        return card;
    },

    // pending → answering
    playCard(cardId) {
        const card = this.cards.find(c => c.id === cardId);
        if (!card || card.status !== 'pending') return null;

        card.setAnswering();
        this.currentCardId = card.id;
        this.save();
        return card;
    },

    // 完成当前 answering 卡牌
    completeCurrentCard() {
        const card = this.getAnsweringCard();
        if (card) {
            card.setAnswered();
            this.currentCardId = null;
            this.save();
        }
        return card;
    },

    // 自动出牌：手牌区最左侧
    autoPlay() {
        const pending = this.getPendingCards();
        if (pending.length === 0) return null;
        return pending[0];
    },

    save() {
        localStorage.setItem('hearthstone_chat_state', JSON.stringify({
            sessionID: this.sessionID,
            cards: this.cards,
            nextCardNumber: this.nextCardNumber,
            navMode: this.navMode,
        }));
    },

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

// ---------- DOM 渲染 ----------

// 渲染手牌区
function renderHand() {
    const container = document.getElementById('hand-cards');
    container.innerHTML = '';

    const pending = State.getPendingCards();
    pending.forEach(card => {
        const el = createCardElement(card);
        // 左键点击：插队到最前并出牌
        el.addEventListener('click', () => {
            const idx = State.cards.findIndex(c => c.id === card.id);
            if (idx > 0) {
                State.cards.splice(idx, 1);
                State.cards.unshift(card);
                State.save();
                renderHand();
            }
            playCard(card.id);
        });
        // 右键菜单
        el.addEventListener('contextmenu', e => {
            e.preventDefault();
            showContextMenu(e, card);
        });
        container.appendChild(el);
    });
}

// 创建小卡牌 DOM
function createCardElement(card) {
    const el = document.createElement('div');
    el.className = `hs-card ${card.status}`;
    el.dataset.cardId = card.id;

    const content = card.content.length > 20
        ? card.content.substring(0, 20) + '...'
        : card.content;
    const time = new Date(card.timestamp)
        .toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });

    el.innerHTML = `
        <div class="card-title">${escapeHtml(card.content.substring(0, 15))}</div>
        <div class="card-content">${escapeHtml(content)}</div>
        <div class="card-footer">#${card.cardNumber} ${time}</div>
    `;
    return el;
}

// 渲染战场区
function renderBattlefield() {
    const container = document.getElementById('battlefield-cards');
    container.innerHTML = '';

    // answered 卡牌（按时间排列）
    const answered = State.getAnsweredCards();
    answered.forEach(card => {
        container.appendChild(createBattlefieldCard(card));
    });

    // answering 卡牌
    const answering = State.getAnsweringCard();
    if (answering) {
        container.appendChild(createBattlefieldCard(answering));
    }

    // 滚动到 answering 或底部
    if (answering) {
        const el = container.parentElement.querySelector('.battlefield-card.answering');
        if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
    } else {
        scrollToBottom();
    }
}

// 创建战场区大卡片
function createBattlefieldCard(card) {
    const el = document.createElement('div');
    el.className = `battlefield-card ${card.status}`;
    el.dataset.cardId = card.id;

    const statusText = card.status === 'answering' ? '回答中...'
                     : card.status === 'answered' ? '已完成' : '等待中';

    let html = `
        <div class="card-header">
            <div class="question">${escapeHtml(card.content)}</div>
            <span class="status-badge">${statusText}</span>
        </div>
    `;

    if (card.answer || card.toolCalls.length > 0) {
        html += `<div class="card-body">${renderMarkdown(card.answer)}</div>`;

        // 工具调用区域
        html += '<div class="card-tools">';
        card.toolCalls.forEach(tool => {
            const icon = tool.success === false ? '❌' : '✅';
            const resultStr = typeof tool.result === 'string'
                ? tool.result : JSON.stringify(tool.result, null, 2);
            html += `
                <div class="tool-call">
                    <details ${tool.open ? 'open' : ''}>
                        <summary>${icon} ${tool.name}</summary>
                        <pre>${escapeHtml(resultStr)}</pre>
                    </details>
                </div>
            `;
        });
        html += '</div>';
    }

    el.innerHTML = html;
    return el;
}

// 更新战场区单个卡牌（SSE 流式更新时用，避免全量重渲染）
function updateBattlefieldCard(card) {
    const el = document.querySelector(`.battlefield-card[data-card-id="${card.id}"]`);
    if (!el) return;

    const bodyEl = el.querySelector('.card-body');
    if (bodyEl) {
        bodyEl.innerHTML = renderMarkdown(card.answer);
    }

    // 工具调用区域更新
    let toolsEl = el.querySelector('.card-tools');
    if (card.toolCalls.length > 0) {
        const lastTool = card.toolCalls[card.toolCalls.length - 1];
        if (!toolsEl) {
            toolsEl = document.createElement('div');
            toolsEl.className = 'card-tools';
            el.appendChild(toolsEl);
        }
        // 检查是否需要新增工具调用
        const existingTools = toolsEl.querySelectorAll('.tool-call');
        if (existingTools.length < card.toolCalls.length) {
            const icon = lastTool.success === false ? '❌' : '✅';
            const resultStr = typeof lastTool.result === 'string'
                ? lastTool.result : JSON.stringify(lastTool.result, null, 2);
            const toolDiv = document.createElement('div');
            toolDiv.className = 'tool-call';
            toolDiv.innerHTML = `
                <details ${lastTool.open ? 'open' : ''}>
                    <summary>${icon} ${lastTool.name}</summary>
                    <pre>${escapeHtml(resultStr)}</pre>
                </details>
            `;
            toolsEl.appendChild(toolDiv);
        } else if (existingTools.length > 0) {
            // 更新最后一个工具调用的结果
            const lastToolEl = existingTools[existingTools.length - 1];
            const summary = lastToolEl.querySelector('summary');
            const pre = lastToolEl.querySelector('pre');
            const icon = lastTool.success === false ? '❌' : '✅';
            if (summary) summary.textContent = `${icon} ${lastTool.name}`;
            if (pre) pre.textContent = typeof lastTool.result === 'string'
                ? lastTool.result : JSON.stringify(lastTool.result, null, 2);
        }
    }

    // 滚动到底部（如果用户在底部附近）
    const bf = document.getElementById('battlefield');
    const isNearBottom = bf.scrollHeight - bf.scrollTop - bf.clientHeight < 100;
    if (isNearBottom) {
        scrollToBottom();
    }
}

// 渲染墓地导航区
function renderGraveyard() {
    const container = document.getElementById('graveyard-timeline');
    if (!container) return;
    container.innerHTML = '';

    const answered = State.getAnsweredCards();
    answered.forEach(card => {
        const el = document.createElement('div');
        el.className = 'graveyard-item';
        el.dataset.cardId = card.id;

        const answerPreview = card.answer.length > 30
            ? card.answer.substring(0, 30) + '...'
            : card.answer;
        const time = new Date(card.answeredAt)
            .toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });

        el.innerHTML = `
            <div class="item-time">${time}</div>
            <div class="item-text">#${card.cardNumber} ${escapeHtml(card.content.substring(0, 20))}</div>
        `;

        el.addEventListener('click', () => {
            const bfCard = document.querySelector(
                `.battlefield-card[data-card-id="${card.id}"]`
            );
            if (bfCard) bfCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
        });

        container.appendChild(el);
    });
}

// ---------- 导图模式 ----------
function renderMindmap() {
    const canvas = document.getElementById('mindmap-canvas');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    const dpr = window.devicePixelRatio || 1;

    const rect = canvas.parentElement.getBoundingClientRect();
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    canvas.style.width = rect.width + 'px';
    canvas.style.height = rect.height + 'px';
    ctx.scale(dpr, dpr);

    const width = rect.width;
    const height = rect.height;

    const cards = State.getAnsweredCards();
    if (cards.length === 0) {
        ctx.fillStyle = '#666';
        ctx.font = '14px sans-serif';
        ctx.textAlign = 'center';
        ctx.fillText('暂无历史问答', width / 2, height / 2);
        return;
    }

    // 网格布局
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

    // 连线
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

    // 节点
    nodes.forEach(node => {
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

        // 问题预览
        ctx.fillStyle = '#aaa';
        ctx.font = '10px sans-serif';
        const preview = node.card.content.substring(0, 8) + '...';
        ctx.fillText(preview, node.x, node.y + nodeRadius + 12);
    });

    // 点击跳转
    canvas.onclick = e => {
        const canvasRect = canvas.getBoundingClientRect();
        const x = e.clientX - canvasRect.left;
        const y = e.clientY - canvasRect.top;

        nodes.forEach(node => {
            const dx = x - node.x;
            const dy = y - node.y;
            if (dx * dx + dy * dy < nodeRadius * nodeRadius) {
                const bfCard = document.querySelector(
                    `.battlefield-card[data-card-id="${node.id}"]`
                );
                if (bfCard) bfCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }
        });
    };
}

// ---------- 右键菜单 ----------
function showContextMenu(e, card) {
    document.querySelectorAll('.context-menu').forEach(m => m.remove());

    const menu = document.createElement('div');
    menu.className = 'context-menu';
    menu.style.left = e.pageX + 'px';
    menu.style.top = e.pageY + 'px';

    const answeredCards = State.getAnsweredCards();
    const canMarkParent = answeredCards.length > 0 && card.status === 'pending';

    let menuHtml = `
        <div class="context-menu-item" data-action="top">置顶</div>
    `;
    if (canMarkParent) {
        menuHtml += `
            <div class="context-menu-divider"></div>
            <div class="context-menu-label">标记为追问：</div>
            ${answeredCards.map(c => `
                <div class="context-menu-item" data-action="parent" data-parent-id="${c.id}">
                    #${c.cardNumber} ${escapeHtml(c.content.substring(0, 15))}...
                </div>
            `).join('')}
        `;
    }
    menuHtml += `
        <div class="context-menu-divider"></div>
        <div class="context-menu-item danger" data-action="delete">删除</div>
    `;
    menu.innerHTML = menuHtml;

    menu.addEventListener('click', ev => {
        const item = ev.target.closest('.context-menu-item');
        if (!item) return;

        const action = item.dataset.action;
        if (action === 'top') {
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
    setTimeout(() => {
        document.addEventListener('click', function closeMenu() {
            menu.remove();
            document.removeEventListener('click', closeMenu);
        }, { once: true });
    }, 0);
}

// ---------- 出牌和 SSE ----------

// 出牌并开始回答
async function playCard(cardId) {
    if (State.isSending) return; // 防止并发

    const card = State.playCard(cardId);
    if (!card) return;

    renderHand();
    renderBattlefield();

    await sendCardAnswer(card);
}

// SSE 请求
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
                    } catch(err) {
                        console.error('解析 SSE 失败:', line, err);
                    }
                }
            }
        }
    } catch(err) {
        if (err.name !== 'AbortError') showError('连接错误: ' + err.message);
    } finally {
        State.completeCurrentCard();
        State.isSending = false;
        State.abortController = null;
        updateSendButton();

        renderBattlefield();
        renderGraveyard();

        // 自动出下一张
        setTimeout(() => {
            const next = State.autoPlay();
            if (next) playCard(next.id);
        }, 500);
    }
}

// SSE 事件处理
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

// ---------- 事件绑定 ----------

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

    // 侧边栏 tab 切换
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            setNavMode(btn.dataset.panel === 'mindmap-panel' ? 'mindmap' : 'graveyard');
        });
    });

    // 战场区滚动检测
    document.getElementById('battlefield').addEventListener('scroll', handleScroll);
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

    // 没有正在回答的卡牌时自动出牌
    if (!State.getAnsweringCard() && !State.isSending) {
        playCard(card.id);
    }
}

// 停止发送
function stopSending() {
    if (State.abortController) {
        State.abortController.abort();
    }
}

// 清空
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

    document.getElementById('tab-graveyard').classList.toggle('active', mode === 'graveyard');
    document.getElementById('tab-mindmap').classList.toggle('active', mode === 'mindmap');
    document.getElementById('graveyard-panel').classList.toggle('active', mode === 'graveyard');
    document.getElementById('mindmap-panel').classList.toggle('active', mode === 'mindmap');

    if (mode === 'mindmap') {
        renderMindmap();
    }
}

// 滚动检测
function handleScroll() {
    const bf = document.getElementById('battlefield');
    const isAtBottom = bf.scrollHeight - bf.scrollTop - bf.clientHeight < 50;

    if (isAtBottom) {
        const answering = State.getAnsweringCard();
        if (answering) {
            const el = document.querySelector('.battlefield-card.answering');
            if (el) el.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }
    }
}

// 更新发送按钮
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

// ---------- 配置加载 ----------

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

// ---------- 工具函数 ----------

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

function scrollToBottom() {
    const el = document.getElementById('battlefield');
    el.scrollTop = el.scrollHeight;
}

// textarea 自适应高度
document.addEventListener('input', e => {
    if (e.target.id === 'input') {
        e.target.style.height = 'auto';
        e.target.style.height = Math.min(e.target.scrollHeight, 120) + 'px';
    }
});

// ---------- 初始化 ----------
document.addEventListener('DOMContentLoaded', () => {
    loadConfig();
    bindEvents();
    renderHand();
    renderBattlefield();
    renderGraveyard();
    setNavMode(State.navMode);
});
