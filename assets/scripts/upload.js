(function() {
  
  function Uploads() {
    this.stopTimeout = null;
    
    var elements = [$(document.body), $('#upload-view')];
    var dragOver = function(e) {
      this.handleDragOver(e.originalEvent);
    }.bind(this);
    var dragLeave = function(e) {
      this.handleDragLeave(e.originalEvent);
    }.bind(this);
    var drop = function(e) {
      this.handleDrop(e.originalEvent);
    }.bind(this);
    for (var i = 0, len = elements.length; i < len; ++i) {
      elements[i].bind('dragover', dragOver);
      elements[i].bind('dragleave', dragLeave);
      elements[i].bind('drop', drop);
    }
    $(window).resize(this.resize.bind(this));
    this.resize();
  }
  
  Uploads.prototype.delay = function(fn) {
    if (this.stopTimeout !== null) {
      clearTimeout(this.stopTimeout);
    }
    this.stopTimeout = setTimeout(function() {
      this.stopTimeout = null;
      fn();
    }.bind(this), 10);
  };
  
  Uploads.prototype.handleDragLeave = function(e) {
    e.preventDefault();
    this.delay(function() {
      window.app.circle.borderRegular();
    });
  };
  
  Uploads.prototype.handleDragOver = function(e) {
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'copy';
    }
    e.stopPropagation();
    e.preventDefault();
    this.delay(function() {
      window.app.circle.borderAnts();
    });
  };
  
  Uploads.prototype.handleDrop = function(e) {
    if (this.stopTimeout !== null) {
      clearTimeout(this.stopTimeout);
      this.stopTimeout = null;
    }
    e.stopPropagation();
    e.preventDefault();
    var file = e.dataTransfer.files[0];
    if (!file) {
      window.app.circle.borderRegular();
    } else {
      window.app.circle.borderUploading();
      this.uploadFile(file);
    }
  };
  
  Uploads.prototype.resize = function() {
    $('#upload-view').css({height: $(window).height()});
  };
  
  Uploads.prototype.uploadFile = function(file) {
    var formData = new FormData();
    formData.append(file.name, file);
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/upload', true);
    xhr.addEventListener('load', function(e) {
      window.app.circle.borderRegular();
      var value
      try {
        var value = JSON.parse(xhr.response);
      } catch (e) {
        alert('error parsing json');
        return;
      }
      if (value.error) {
        // Well, there was an error.
        alert('it failed');
      } else {
        // It worked, apparently.
      }
    });
    xhr.addEventListener('error', function() {
      window.app.circle.borderRegular();
      alert('failed');
    });
    xhr.upload.addEventListener('progress', function(e) {
      if (e.lengthComputable) {
        var percent = e.loaded / e.total;
        window.app.circle.animationInfo = percent;
        window.app.circle.draw();
      }
    });
    xhr.send(formData);
  };
  
  $(function() {
    new Uploads();
  });
  
})();