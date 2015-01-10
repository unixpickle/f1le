(function() {
  
  function Circle() {
    this.border = $('#circle-border')[0];
    this.circle = $('#circle');
    this.logo = $('#circle-logo');
    this.resize();
    $(window).resize(this.resize.bind(this));
  }
  
  Circle.prototype.draw = function() {
    var ctx = this.border.getContext('2d');
    var size = this.border.width;
    var scale = this.border.width / $(this.border).width();
    
    // Draw a nice circle.
    var width = 10*scale*(size/500);
    ctx.clearRect(0, 0, size, size);
    ctx.beginPath();
    ctx.arc(size/2, size/2, size/2-width/2, 0, 2*Math.PI, false);
    ctx.lineWidth = width;
    ctx.strokeStyle = 'white';
    ctx.stroke();
  }
  
  Circle.prototype.resize = function() {
    var smallerDim = Math.min($(window).width(), $(window).height());
    var size = Math.max(Math.min(smallerDim-100, 500), 300);
    var circleTop = ($(window).height()-size) / 2;
    var circleLeft = ($(window).width()-size) / 2;
    this.circle.css({width: size, height: size, top: circleTop,
      left: circleLeft, "font-size": size/20});
  
    this.border.width = size;
    this.border.height = size;
    this.draw();
  };
  
  $(function() {
    new Circle();
  });
  
})();