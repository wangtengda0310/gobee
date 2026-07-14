// 点"下载并安装"(cards[4] 第2个按钮;第1个=检查更新)
(function(){
  var card = document.querySelectorAll('.setting-card')[4];
  var btns = card ? card.querySelectorAll('.n-button') : [];
  if(btns.length >= 2) btns[1].click();
  return JSON.stringify({ btnCount: btns.length, clickedInstall: btns.length >= 2 });
})()