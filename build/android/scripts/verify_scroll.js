// 测各层 scrollWidth/scrollLeft,找横向可滚动元素(白边根因)
(function(){
  function info(sel){
    var el = document.querySelector(sel);
    if(!el) return null;
    var cs = getComputedStyle(el);
    return { sw: el.scrollWidth, cw: el.clientWidth, sl: el.scrollLeft, ofx: cs.overflowX, canScroll: el.scrollWidth > el.clientWidth + 1 };
  }
  return JSON.stringify({
    de: info('html'),
    body: info('body'),
    layout: info('#layout'),
    header: info('#layout-header'),
    content: info('#layout-content'),
    footer: info('#layout-footer'),
    wikiRoot: info('#HeroWikiResCheck'),
    vw: window.innerWidth
  });
})()