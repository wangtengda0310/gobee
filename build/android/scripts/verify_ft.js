// 验证 function-test(当前页) sider 折叠 + content 宽 + 横向溢出
(function(){
  var sider = document.querySelector('.n-layout-sider');
  var content = document.querySelector('#Content');
  return JSON.stringify({
    siderCollapsed: sider ? (sider.className.indexOf('collapsed') >= 0) : null,
    siderW: sider ? Math.round(sider.getBoundingClientRect().width) : null,
    contentW: content ? Math.round(content.getBoundingClientRect().width) : null,
    vw: window.innerWidth,
    docScrollW: document.documentElement.scrollWidth,
    bodyOverflowX: getComputedStyle(document.body).overflowX
  });
})()