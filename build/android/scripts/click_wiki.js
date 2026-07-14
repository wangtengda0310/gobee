// 点导航按钮 index 5(武将Wiki检查)。全 ASCII 避免 PowerShell GBK 中文乱码
(function(){
  var b = document.querySelectorAll('.idea-icon-button')[5];
  if(b){ b.click(); return 'idx5-clicked'; }
  return 'no-idx5';
})()