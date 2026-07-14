// 验证 wiki config-card + PathConfigInput 双输入框 + scrollW(拖出白边根因)
(function(){
  var card = document.querySelector('.config-card');
  var inputs = document.querySelectorAll('.config-card .n-input');
  var btns = document.querySelectorAll('.config-card .n-button');
  return JSON.stringify({
    onWiki: !!document.querySelector('#HeroWikiResCheck'),
    cardW: card ? Math.round(card.getBoundingClientRect().width) : null,
    inputCount: inputs.length,
    input1W: inputs.length > 0 ? Math.round(inputs[0].getBoundingClientRect().width) : null,
    input2W: inputs.length > 1 ? Math.round(inputs[1].getBoundingClientRect().width) : null,
    btnCount: btns.length,
    scrollW: document.documentElement.scrollWidth,
    vw: window.innerWidth
  });
})()