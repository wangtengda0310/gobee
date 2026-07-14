// 验证当前页(应为 wiki)+ main-content 宽 + 横向溢出
(function(){
  var wiki = document.querySelector('#HeroWikiResCheck');
  var test = document.querySelector('#Test');
  var anchor = document.querySelector('.anchor-container');
  var main = document.querySelector('.main-content');
  return JSON.stringify({
    onWiki: !!wiki,
    onTest: !!test,
    anchorExists: !!anchor,
    anchorDisplay: anchor ? getComputedStyle(anchor).display : 'no-anchor(needs-runCheck)',
    mainW: main ? Math.round(main.getBoundingClientRect().width) : null,
    scrollW: document.documentElement.scrollWidth,
    vw: window.innerWidth
  });
})()