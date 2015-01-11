(function() {
  
  var StateNormal = 'normal';
  var StateAnts = 'ants';
  var StateUploading = 'uploading';
  
  function Circle() {
    window.app.BaseCircle.call(this);
    this.state = StateNormal;
    this.animationInfo = null;
    this.animateInterval = null;
    this.draw();
  }
  
  Circle.prototype = Object.create(window.app.BaseCircle.prototype);
  
  Circle.prototype.borderAnts = function() {
    if (this.state === StateAnts) {
      return;
    }
    this.state = StateAnts;
    this.animationInfo = (new Date()).getTime();
    this.animateInterval = setInterval(function() {
      this.draw();
    }.bind(this), 1000/24);
    this.draw();
  };
  
  Circle.prototype.borderRegular = function() {
    if (this.state === StateNormal) {
      return;
    }
    this.state = StateNormal;
    this.animationInfo = null;
    if (this.animationInterval !== null) {
      clearInterval(this.animationInterval);
      this.animationInterval = null;
    }
    this.draw();
  };
  
  Circle.prototype.borderUploading = function() {
    this.state = StateUploading;
    this.animationInfo = 0;
    this.draw();
  };
  
  Circle.prototype.draw = function() {
    if (this.state === StateNormal) {
      window.app.BaseCircle.prototype.draw.call(this);
      return;
    }
    
    var ctx = this.border.getContext('2d');
    var size = this.border.width;
    var thickness = this.borderThickness();
    
    ctx.clearRect(0, 0, size, size);
    
    ctx.beginPath();
    ctx.arc(size/2, size/2, size/2-thickness/2, 0, 2*Math.PI, false);
    ctx.lineWidth = thickness;
    ctx.fillStyle = 'rgba(101, 188, 212, 0.8)';
    ctx.fill();
    
    ctx.strokeStyle = '#d7d7d7';
    if (this.state === StateAnts) {
      var runningTime = (new Date()).getTime() - this.animationInfo;
      var angle = (runningTime/3000) % (Math.PI*2);
      var antCount = 20;
      var antAngle = Math.PI*2/(antCount*2);
      for (var i = 0; i < antCount; ++i) {
        ctx.beginPath();
        var startAngle = angle + antAngle*2*i;
        var endAngle = angle + antAngle*(2*i+1);
        ctx.arc(size/2, size/2, size/2-thickness/2, startAngle, endAngle,
          false);
        ctx.stroke();
      }
    } else if (this.state === StateUploading) {
      // TODO: this
    }
  };
  
  $(function() {
    window.app.circle = new Circle();
  });
  
})();