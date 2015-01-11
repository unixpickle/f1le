(function() {
  
  function Files() {
    this.element = $('#file-list');
    $(window).resize(this.resize.bind(this));
    this.resize();
  }
  
  Files.prototype.resize = function() {
    this.element.css({top: $(window).height(),
      "min-height": $(window).height()});
  };
  
  $(function() {
    window.app.files = new Files();
  });
  
  if (!window.app) {
    window.app = {};
  }
  
})();