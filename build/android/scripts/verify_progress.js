// 查下载进度 + 状态(btnText 含百分比 = 下载中;变化 = installApk 触发后前端状态)
(function(){
  var card = document.querySelectorAll('.setting-card')[4];
  var prog = card ? card.querySelector('.n-progress') : null;
  var btns = card ? card.querySelectorAll('.n-button') : [];
  return JSON.stringify({
    btnCount: btns.length,
    lastBtnText: btns.length ? (btns[btns.length-1].textContent||'').replace(/[^a-zA-Z0-9 %]/g,'?') : '',
    hasProgress: !!prog,
    hasRedError: !!(card && (card.innerHTML||'').indexOf('ff6b6b')>=0)
  });
})()