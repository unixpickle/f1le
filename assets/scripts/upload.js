(function() {
  
  var stopTimeout = null;
  
  function delayedFunction(fn) {
    if (stopTimeout !== null) {
      clearTimeout(stopTimeout);
    }
    stopTimeout = setTimeout(function() {
      stopTimeout = null;
      fn();
    }, 10);
  }
  
  function handleDragOver(e) {
    e.preventDefault();
    delayedFunction(function() {
      window.app.circle.borderAnts();
    });
  }
  
  function handleDragLeave(e) {
    e.preventDefault();
    delayedFunction(function() {
      window.app.circle.borderRegular();
    });
  }
  
  function handleResize() {
    $('#upload-view').css({height: $(window).height()});
  }
  
  $(function() {
    var elements = [$(document.body), $('#upload-view')];
    for (var i = 0, len = elements.length; i < len; ++i) {
      elements[i].bind('dragover', handleDragOver);
      elements[i].bind('dragleave', handleDragLeave);
    }
    $(window).resize(handleResize);
    handleResize();
  });
  
})();