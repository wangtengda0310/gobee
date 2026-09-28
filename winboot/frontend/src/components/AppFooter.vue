<script setup>
import {ref, onMounted} from 'vue'
import {Events} from '@wailsio/runtime'

// 实时时钟: 订阅 Go 侧每秒发出的 "time" 事件 (作为 center 插槽的默认内容).
// 窄屏只显示时分秒 (完整 RFC1123 太宽, 对齐全局 640px 断点).
const time = ref('Listening for Time event...')

onMounted(() => {
  Events.On('time', (timeValue) => {
    const full = timeValue.data;
    const compact = (full.match(/\d{1,2}:\d{2}:\d{2}/) || [full])[0];
    time.value = window.matchMedia('(max-width: 640px)').matches ? compact : full;
  })
})
</script>

<template>
  <hr class="footer-divider"/>
  <footer class="footer">
    <!-- 通用三栏: left / center / right 具名插槽, 均带默认内容;
         使用方按需覆盖任意一栏, 不传则保持默认 (版本 | 时钟 | 文档链接).
         ⚠ 覆盖插槽时: 若新内容是链接/按钮, 必须带 --wails-draggable: no-drag
         (或直接复用 .footer-docs 类) —— 底栏处于 wails 窗口标题栏拖拽区域内,
         不声明则点击会被当作拖拽窗口, 链接点不了. -->
    <span class="footer-left">
      <slot name="left">0.0.1</slot>
    </span>
    <span class="footer-center">
      <slot name="center">
        <span class="footer-time">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          <span id="time">{{ time }}</span>
        </span>
      </slot>
    </span>
    <span class="footer-right">
      <slot name="right">
        <a class="footer-docs" data-wml-openURL="https://v3.wails.io" aria-label="Wails documentation">Docs
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><line x1="7" y1="17" x2="17" y2="7"/><polyline points="7 7 17 7 17 17"/></svg>
        </a>
      </slot>
    </span>
  </footer>
</template>

<style scoped>
/* 底栏: 细分隔线 + 左中右三栏 (插槽化, 内容可按需替换). */
.footer-divider {
  width: 100%;
  height: 0;
  border: none;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  margin: 0;
}
.footer {
  display: grid;
  grid-template-columns: 1fr auto 1fr;   /* left | center | right */
  align-items: center;
  width: 100%;
  /* 离底边留空隙, 避开 home indicator */
  padding: 11px max(0.5rem, env(safe-area-inset-right))
           max(11px, env(safe-area-inset-bottom)) 0;
  color: var(--muted);
  font-size: 12.5px;
}
/* 三栏容器共用形态, 仅对齐方向不同 */
.footer-left,
.footer-center,
.footer-right {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  white-space: nowrap;
  min-width: 0;
}
.footer-left   { justify-self: start; }
.footer-center { justify-self: center; }
.footer-right  { justify-self: end; }

/* --- 以下为各插槽的默认内容样式 --- */
.footer-time svg { width: 19px; height: 19px; opacity: 0.85; }
.footer-docs {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: var(--muted);
  text-decoration: none;
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.2s ease;
  /* 链接可点击 (窗口标题栏拖拽区域例外) */
  --wails-draggable: no-drag;
}
.footer-docs:hover { color: var(--text); }
.footer-docs svg { width: 17px; height: 17px; }

@media (max-width: 640px) {
  .footer { font-size: 11px; gap: 4px; }
  .footer-left, .footer-center, .footer-right { gap: 5px; }
  .footer-time svg { width: 13px; height: 13px; }
}
</style>
