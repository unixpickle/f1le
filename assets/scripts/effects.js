(function() {
  
  $(function() {
    var canvas = $('#pentagons')[0];
    var pents = new window.app.Pentagons(canvas);
    pents.begin();
    
    var resizeFunc = function() {
      var width = $(window).width();
      var height = $(window).height();
      canvas.width = width;
      canvas.height = height;
      pents.draw();
    };
    
    $(window).resize(resizeFunc);
    resizeFunc();
  });
  
})();