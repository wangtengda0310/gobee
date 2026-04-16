// 全局状态
let sessionID = localStorage.getItem('chat_session_id') || generateUUID();
let currentModel = '';
let abortController = null;
let isSending = false;

localStorage.setItem('chat_session_id', sessionID);

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    loadConfig();
    bindEvents();
});

function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
        const r = Math.random() * 16 | 0;
        return (c === 'x' ? r : (r & 0x3 | 0x8)).toString(16);
    });
}

// 加载配置
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
        currentModel = config.current;
    } catch(e) {
        showError('加载配置失败: ' + e.message);
    }
}

// 绑定事件
function bindEvents() {
    const input = document.getElementById('input');
    const btnSend = document.getElementById('btn-send');
    const btnClear = document.getElementById('btn-clear');

    input.addEventListener('keydown', e => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });
    btnSend.addEventListener('click', () => {
        if (isSending) stopSending();
        else sendMessage();
    });
    btnClear.addEventListener('click', clearHistory);
    document.getElementById('model-select').addEventListener('change', e => {
        switchModel(e.target.value);
    });
}

// 发送消息（核心 SSE 逻辑）
async function sendMessage() {
    const input = document.getElementById('input');
    const content = input.value.trim();
    if (!content || isSending) return;

    input.value = '';
    input.style.height = 'auto';
    appendMessage('user', content);

    setSending(true);
    abortController = new AbortController();

    try {
        const resp = await fetch('/chat/api/message', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ content, model: currentModel, session_id: sessionID }),
            signal: abortController.signal,
        });

        if (!resp.ok) {
            throw new Error(`HTTP ${resp.status}: ${resp.statusText}`);
        }

        const reader = resp.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';
        let aiDiv = null;

        while (true) {
            const { done, value } = await reader.read();
            if (done) break;

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');
            buffer = lines.pop(); // 保留不完整的行

            let currentEvent = '';
            for (const line of lines) {
                if (line.startsWith('event: ')) {
                    currentEvent = line.slice(7);
                } else if (line.startsWith('data: ')) {
                    try {
                        const data = JSON.parse(line.slice(6));
                        handleSSEEvent(currentEvent, data, div => { aiDiv = div; });
                    } catch(e) {
                        console.error('解析 SSE 数据失败:', line, e);
                    }
                }
            }
        }
    } catch(e) {
        if (e.name !== 'AbortError') showError('连接错误: ' + e.message);
    } finally {
        setSending(false);
        abortController = null;
    }
}

function handleSSEEvent(event, data, setAiDiv) {
    switch(event) {
        case 'content':
            // 找到或创建 AI 消息 div
            let aiDiv = document.querySelector('.message.assistant:last-child .content');
            if (!aiDiv) {
                const wrapper = appendMessage('assistant', '');
                aiDiv = wrapper.querySelector('.content');
                setAiDiv(wrapper);
            }
            aiDiv.innerHTML += renderMarkdown(data.text);
            scrollToBottom();
            break;
        case 'tool_use':
            appendToolCall(data.name, data.args);
            break;
        case 'tool_result':
            appendToolResult(data.name, data.result, data.success);
            break;
        case 'error':
            showError(data.message);
            break;
        case 'done':
            // 流结束
            break;
    }
}

// 停止发送
function stopSending() {
    if (abortController) abortController.abort();
    setSending(false);
}

// 追加消息到界面
function appendMessage(role, content) {
    const div = document.createElement('div');
    div.className = `message ${role}`;
    div.innerHTML = `<div class="content">${renderMarkdown(content)}</div>`;
    document.getElementById('messages').appendChild(div);
    scrollToBottom();
    return div;
}

// 追加工具调用
function appendToolCall(name, args) {
    const div = document.createElement('div');
    div.className = 'tool-call';
    div.innerHTML = `<details open><summary>🔧 调用 ${name}</summary><pre>${JSON.stringify(args, null, 2)}</pre></details>`;
    document.getElementById('messages').appendChild(div);
    scrollToBottom();
}

// 追加工具结果
function appendToolResult(name, result, success) {
    const icon = success ? '✅' : '❌';
    const div = document.createElement('div');
    div.className = 'tool-call';
    const resultStr = typeof result === 'string' ? result : JSON.stringify(result, null, 2);
    div.innerHTML = `<details><summary>${icon} ${name} 结果</summary><pre>${escapeHtml(resultStr)}</pre></details>`;
    document.getElementById('messages').appendChild(div);
    scrollToBottom();
}

// HTML 转义
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 简易 Markdown 渲染
function renderMarkdown(text) {
    if (!text) return '';
    // 转义 HTML
    text = escapeHtml(text);
    // 代码块
    text = text.replace(/```(\w*)\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>');
    // 行内代码
    text = text.replace(/`([^`]+)`/g, '<code>$1</code>');
    // 粗体
    text = text.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    // 无序列表
    text = text.replace(/^- (.+)$/gm, '<li>$1</li>');
    text = text.replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>');
    // 换行
    text = text.replace(/\n/g, '<br>');
    return text;
}

// 切换模型
async function switchModel(model) {
    try {
        const resp = await fetch('/chat/api/config', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ model }),
        });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        currentModel = model;
    } catch(e) {
        showError('切换模型失败: ' + e.message);
        // 恢复选择
        const select = document.getElementById('model-select');
        select.value = currentModel;
    }
}

// 清空历史
async function clearHistory() {
    try {
        const resp = await fetch('/chat/api/history?session_id=' + sessionID, { method: 'DELETE' });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        document.getElementById('messages').innerHTML = '';
    } catch(e) {
        showError('清空历史失败: ' + e.message);
    }
}

// 设置发送状态
function setSending(sending) {
    isSending = sending;
    const btn = document.getElementById('btn-send');
    if (sending) {
        btn.textContent = '停止';
        btn.classList.add('sending');
    } else {
        btn.textContent = '发送';
        btn.classList.remove('sending');
    }
    btn.disabled = false;
}

// 显示错误
function showError(msg) {
    const toast = document.createElement('div');
    toast.className = 'error-toast';
    toast.textContent = msg;
    document.body.appendChild(toast);
    setTimeout(() => toast.remove(), 5000);
}

// 滚动到底部
function scrollToBottom() {
    const el = document.getElementById('messages');
    el.scrollTop = el.scrollHeight;
}

// textarea 自适应高度
document.addEventListener('input', e => {
    if (e.target.id === 'input') {
        e.target.style.height = 'auto';
        e.target.style.height = Math.min(e.target.scrollHeight, 120) + 'px';
    }
});
