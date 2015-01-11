(function() {
  
  function Files() {
    this.element = $('#file-list');
    this.background = $('#file-list-background');
    this.background.css({position: "fixed", "background-color": "#888",
      left: 0, width: "100%"});
    $(window).resize(this.resize.bind(this));
    $(window).scroll(this.scroll.bind(this));
    this.resize();
  }
  
  Files.prototype.resize = function() {
    this.element.css({top: $(window).height(),
      "min-height": $(window).height()});
    this.background.css({height: this.element.height()});
    this.scroll();
  };
  
  Files.prototype.scroll = function() {
    var topOffset = $(window).height() - $(window).scrollTop();
    topOffset = Math.max(topOffset, 0);
    console.log(topOffset);
    this.background.css({top: topOffset});
  };
  
  $(function() {
    window.app.files = new Files();
  });
  
  if (!window.app) {
    window.app = {};
  }
  
})();