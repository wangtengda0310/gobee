// 验证 Android 适配:全局判定 + 当前页锚点(用 IIFE 避免 var 作用域问题)
JSON.stringify({
  isMobile: document.documentElement.classList.contains('is-mobile'),
  coarse: matchMedia('(pointer: coarse)').matches,
  cls: document.documentElement.className,
  path: location.hash,
  anchorDisplay: (function(){var a=document.querySelector('.anchor-container');return a?getComputedStyle(a).display:'no-anchor-el';})(),
  vw: window.innerWidth
})