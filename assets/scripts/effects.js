(function() {
  
  $(function() {
    var canvas = $('#pentagons')[0];
    var pents = new window.app.Pentagons(canvas);
    pents.begin();
    
    var scrollFunc = function() {
      pents.grayHeight = $(window).scrollTop();
      pents.draw();
    };
    
    $(window).scroll(scrollFunc);
    
    var resizeFunc = function() {
      var width = $(window).width();
      var height = $(window).height();
      canvas.width = width;
      canvas.height = height;
      scrollFunc();
    };
    
    $(window).resize(resizeFunc);
    resizeFunc();
    
    scrollFunc();
  });
  
})();