(function() {
  
  function Uploads() {
    this.fileInput = $('#file-input');
    this.stopTimeout = null;
    this.uploading = false;
    this._registerEvents();
    
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
    if (this.uploading) {
      return;
    }
    e.preventDefault();
    this.delay(function() {
      window.app.circle.borderRegular();
    });
  };
  
  Uploads.prototype.handleDragOver = function(e) {
    if (this.uploading) {
      return;
    }
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
    if (this.uploading) {
      return;
    }
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
  
  Uploads.prototype.pickerDialog = function() {
    this.fileInput.click();
  };
  
  Uploads.prototype.resize = function() {
    $('#upload-view').css({height: $(window).height()});
  };
  
  Uploads.prototype.uploadFile = function(file) {
    this.uploading = true;
    var totalSize = file && file.size ? file.size : 0;

    var handleProgress = function(loaded, total) {
      if (total) {
        var percent = Math.min(loaded / total, 1);
        window.app.circle.animationInfo = percent;
        window.app.circle.draw();
      }
    };

    var handleResponse = function(value) {
      this.uploading = false;
      window.app.circle.borderRegular();
      if (value.error) {
        window.app.errorDialog('Failed to upload: ' + value.error);
      } else {
        window.location = '/files';
      }
    }.bind(this);

    var handleError = function(message) {
      this.uploading = false;
      window.app.circle.borderRegular();
      window.app.errorDialog(message);
    }.bind(this);

    // Create a multipart request body so we can report progress.
    var canStream = !!(window.ReadableStream && file && file.stream);
    if (canStream) {
      var boundary = '----f1le-boundary-' + Math.random().toString(16).slice(2);
      var safeName = file.name.replace(/"/g, '\\"');
      var contentType = file.type || 'application/octet-stream';
      var header = '--' + boundary + '\r\n' +
        'Content-Disposition: form-data; name="' + safeName + '"; filename="' +
        safeName + '"\r\n' +
        'Content-Type: ' + contentType + '\r\n\r\n';
      var footer = '\r\n--' + boundary + '--\r\n';
      var encoder = new TextEncoder();
      var headerBytes = encoder.encode(header);
      var footerBytes = encoder.encode(footer);
      var total = headerBytes.byteLength + file.size + footerBytes.byteLength;
      var loaded = headerBytes.byteLength;

      handleProgress(loaded, total);

      var stream = new ReadableStream({
        start: function(controller) {
          controller.enqueue(headerBytes);
          var reader = file.stream().getReader();
          var pump = function() {
            return reader.read().then(function(result) {
              if (result.done) {
                loaded += footerBytes.byteLength;
                handleProgress(loaded, total);
                controller.enqueue(footerBytes);
                controller.close();
                return;
              }
              loaded += result.value.byteLength;
              handleProgress(loaded, total);
              controller.enqueue(result.value);
              return pump();
            }).catch(function(err) {
              controller.error(err);
            });
          };
          return pump();
        }
      });

      fetch('/upload', {
        method: 'POST',
        headers: {'Content-Type': 'multipart/form-data; boundary=' + boundary},
        body: stream,
        duplex: 'half'
      }).then(function(response) {
        return response.json().catch(function() {
          throw new Error('Invalid JSON data.');
        });
      }).then(handleResponse).catch(function(err) {
        if (err && err.message === 'Invalid JSON data.') {
          handleError(err.message);
        } else {
          handleError('Failed to connect to the server.');
        }
      });
      return;
    }

    // Fallback: send FormData without progress.
    var formData = new FormData();
    formData.append(file.name, file);
    fetch('/upload', {
      method: 'POST',
      body: formData
    }).then(function(response) {
      return response.json().catch(function() {
        throw new Error('Invalid JSON data.');
      });
    }).then(handleResponse).catch(function(err) {
      if (err && err.message === 'Invalid JSON data.') {
        handleError(err.message);
      } else {
        handleError('Failed to connect to the server.');
      }
    });
  };
  
  Uploads.prototype._registerEvents = function() {
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
    this.fileInput.bind('change', function() {
      var file = this.fileInput[0].files[0];
      if (file) {
        this.uploadFile(file);
      }
    }.bind(this));
  };
  
  $(function() {
    window.app.uploads = new Uploads();
  });
  
  if (!window.app) {
    window.app = {};
  }
  
})();
