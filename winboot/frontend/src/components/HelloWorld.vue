<script setup>
import {ref, onMounted, useTemplateRef} from 'vue'
import {Events, WML} from "@wailsio/runtime";
import {CmdService, GreetService} from "../../bindings/github.com/wangtengda0310/gobee/winboot";
import AppFooter from './AppFooter.vue';

const name = ref('');

const titleNameRef = useTemplateRef('titleNameRef');

// --- 浮动弹窗 (原 toast 重构): 响应式状态驱动, 居中弹出, 点击/超时关闭 ---
const toastMsg = ref('')
const toastVisible = ref(false)
let toastTimer

function showToast(message) {
  toastMsg.value = message
  toastVisible.value = true
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toastVisible.value = false }, 6000)
}

function closeToast() {
  clearTimeout(toastTimer)
  toastVisible.value = false
}

// Crossfade the framework word in the heading ("Wails + Vue") to the name the
// user entered ("Wails + <name>"): the old word fades out while the new one
// fades in over the same spot.
function swapTitleName(newName) {
  const titleNameElement = titleNameRef.value;
  if (!titleNameElement) {
    return;
  }
  const current = titleNameElement.querySelector('.title-name-text:not(.is-outgoing)');
  if (!current || current.textContent === newName) {
    return;
  }
  const incoming = document.createElement('span');
  incoming.className = 'title-name-text is-entering';
  incoming.textContent = newName;
  current.classList.add('is-outgoing');
  titleNameElement.appendChild(incoming);
  // Force a reflow so the transitions run from the starting state.
  void incoming.offsetWidth;
  incoming.classList.remove('is-entering');
  current.classList.add('is-leaving');
  current.addEventListener('transitionend', () => current.remove(), {once: true});
}

const doGreet = () => {
  let n = name.value || 'anonymous';
  swapTitleName(n);
  GreetService.SetServerID(n).then(showToast).catch(console.error);
}

// 统一入口: 命令名是常量, 不复用输入框的值 (输入框语义是 server_id)
const runCmd = (cmd, args = []) => {
  swapTitleName(cmd);
  // 立即反馈: 长命令(install 数分钟)若只等完成回调, 用户会以为点了没反应
  showToast(`⏳ ${cmd} 执行中, 完成后显示结果...`);
  CmdService.Exec(cmd, args).then(showToast).catch(console.error);
}

// GUI 调用必须 -Yes(跳过脚本内 Read-Host, 隐藏控制台无输入会卡死/误判取消) + -NoPause(跳过结尾 ReadKey)
const install = () => runCmd('install', ['-Yes', '-NoPause'])

const uninstall = () => {
  // 双保险: 前端确认 + 脚本侧 -Yes 跳过 Read-Host (隐藏控制台无输入)
  if (!window.confirm('将卸载 Docker Desktop 并清空所有镜像/容器/卷, 确认继续?')) return
  runCmd('uninstall', ['-Yes', '-NoPause'])
}

const up = () => {
  // -Detach: compose 后台运行, 脚本自然结束返回 (否则 Exec 永久卡在 attached 模式)
  // -Mode 1: 免交互选模式 (无控制台的 Read-Host 会阻塞)
  // -NoPause: 跳过"按任意键退出" (隐藏控制台的 ReadKey 会阻塞)
  const args = ['-Detach', '-Mode', '1', '-NoPause']
  // 输入框作为 server_id 透传 (仅新机器 .server_id 缺失时生效, 已配置机器忽略)
  if (name.value.trim()) args.push('-ServerId', name.value.trim())
  runCmd('up', args)
}

const stop = () => runCmd('stop', [])
const restart = () => runCmd('restart', [])
const doctor = () => runCmd('doctor', [])

// --- 容器日志: app 启动即自动跟随 (无需按钮), 事件流渲染到控制台面板 ---
const logLines = ref([])

onMounted(() => {
  Events.On('boot:log', (e) => {
    logLines.value.push(e.data)
    if (logLines.value.length > 500) logLines.value.splice(0, logLines.value.length - 500)
  })
  // 启动即监听: 容器未运行时 Go 侧安静等待, 容器起来自动连接; 停止跟随由 app 退出钩子处理
  CmdService.StartLogs().catch(console.error)
  // 打开时预填当前 server_id (经 serverRootPath 定位, 与启动目录无关)
  GreetService.GetServerID().then((id) => { if (id) name.value = id })
  // Wire up data-wml-openURL links (logos + footer "Docs" link).
  WML.Reload();
})
</script>

<template>
  <main class="container">
    <header class="brand">
      <a class="brand-mark" data-wml-openURL="https://v3.wails.io" aria-label="Wails website">
        <img src="/wails.png" class="brand-logo" alt="Wails logo"/>
      </a>
      <a class="brand-badge" data-wml-openURL="https://vuejs.org/" aria-label="Vue">
        <img src="/vue.svg" alt="Vue logo"/>
      </a>
    </header>

    <h1 class="title"><span class="title-accent">运行命令 +</span> <span class="title-name" ref="titleNameRef"><span class="title-name-text">Vue</span></span></h1>
    <p class="subtitle">Build beautiful cross-platform apps with Go and Vue.</p>

    <div class="greet">
      <div class="input-box" id="input">
        <svg class="input-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        <input aria-label="input" class="input" id="name" v-model="name" type="text" placeholder="Your name" autocomplete="off"/>
        <button aria-label="greet-btn" class="btn" @click="doGreet">设置服务器名称
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg>
        </button>
      </div>

      <button aria-label="install-btn" class="btn" @click="install">一键安装环境</button>
      <button aria-label="uninstall-btn" class="btn" @click="uninstall">一键卸载环境</button>
      <button aria-label="up-btn" class="btn" @click="up">启动服务器</button>
      <button aria-label="restart-btn" class="btn" @click="restart">重启服务器</button>
      <button aria-label="stop-btn" class="btn" @click="stop">停止服务器</button>
      <button aria-label="doctor-btn" class="btn" @click="doctor">检查环境</button>
    </div>
  </main>

  <!-- 底栏 (分隔线/时钟/文档链接): 独立组件, 见 AppFooter.vue -->
  <AppFooter/>

  <!-- 控制台面板: 常驻右上角; 命令输出实时流入([命令名] 前缀), 容器日志经「跟随日志」流入 -->
  <pre aria-label="console-logs" class="logs"><template v-if="logLines.length">{{ logLines.join('\n') }}</template><span v-else class="logs-empty">暂无日志 — 执行任意命令实时输出; 容器日志点击「跟随日志」</span></pre>

  <!-- 浮动弹窗: 命令执行结果, 居中弹出, 点击任意处/超时自动关闭 -->
  <Transition name="popup">
    <div v-if="toastVisible" class="popup-wrap" role="status" aria-live="polite">
      <div class="popup" @click="closeToast">
        <div class="popup-header">
          <span class="popup-label">执行结果</span>
          <button aria-label="关闭弹窗" class="popup-close" type="button" @click.stop="closeToast">&times;</button>
        </div>
        <div aria-label="result" class="popup-msg">{{ toastMsg }}</div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
/* ----- 容器日志面板: 常驻右上角 (原 toast 位置), 玻璃拟态与全局风格一致 ----- */
.logs {
  position: fixed;
  top: max(var(--s-3), env(safe-area-inset-top));
  right: max(var(--s-3), env(safe-area-inset-right));
  z-index: 15;
  width: min(46vw, 520px);
  max-height: 62vh;
  overflow: auto;
  margin: 0;
  padding: 12px 16px;
  border: 1px solid var(--glass-border);
  border-radius: 14px;
  background: rgba(13, 17, 28, 0.72);
  -webkit-backdrop-filter: blur(16px);
  backdrop-filter: blur(16px);
  box-shadow: 0 16px 44px rgba(0, 0, 0, 0.45);
  color: #9fef9f;
  font-family: ui-monospace, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  /* 面板内容允许选中文本复制排障 */
  -webkit-user-select: text;
  user-select: text;
}
/* 空态引导文案: 弱化, 与真实日志区分 */
.logs-empty {
  color: var(--text-muted, rgba(255, 255, 255, 0.45));
  font-family: inherit;
}
/* 窄窗口: 面板改为贴右全宽, 避免挤占内容 */
@media (max-width: 640px) {
  .logs {
    left: max(var(--s-3), env(safe-area-inset-left));
    width: auto;
    max-height: 40vh;
  }
}

/* ----- 浮动弹窗: 居中弹出, 缩放淡入, 点击/超时关闭 ----- */
.popup-wrap {
  position: fixed;
  inset: 0;
  z-index: 30;
  display: flex;
  align-items: center;
  justify-content: center;
  /* 无遮罩层: 不挡背景操作, 仅卡片本体接收点击 */
  pointer-events: none;
  padding: max(var(--s-4), env(safe-area-inset-top)) max(var(--s-4), env(safe-area-inset-right))
           max(var(--s-4), env(safe-area-inset-bottom)) max(var(--s-4), env(safe-area-inset-left));
}
.popup {
  pointer-events: auto;
  cursor: pointer; /* 点击卡片任意处关闭 */
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: min(86vw, 560px);
  max-height: 70vh;
  padding: 14px 18px;
  border: 1px solid var(--glass-border);
  border-radius: 14px;
  background: rgba(13, 17, 28, 0.85);
  -webkit-backdrop-filter: blur(16px);
  backdrop-filter: blur(16px);
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.55);
  color: var(--text, #eef1f8);
}
.popup-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: default;
}
.popup-label {
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-muted, rgba(255, 255, 255, 0.55));
}
.popup-close {
  border: none;
  background: none;
  color: var(--text-muted, rgba(255, 255, 255, 0.55));
  font-size: 18px;
  line-height: 1;
  padding: 2px 6px;
  border-radius: 6px;
  cursor: pointer;
}
.popup-close:hover {
  color: var(--text, #fff);
  background: rgba(255, 255, 255, 0.08);
}
.popup-msg {
  overflow: auto;
  font-size: 13px;
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-all;
  cursor: default;
  -webkit-user-select: text;
  user-select: text;
}

/* 弹入: 从 92% 缩放 + 淡入, 轻微上浮; 弹出反向 */
.popup-enter-active,
.popup-leave-active {
  transition: opacity 0.28s ease, transform 0.32s cubic-bezier(0.2, 0.8, 0.2, 1);
}
.popup-enter-from,
.popup-leave-to {
  opacity: 0;
  transform: scale(0.92);
}
</style>
