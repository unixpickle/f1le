(function () {

  function Uploads() {
    this.fileInput = $('#file-input');
    this.stopTimeout = null;
    this.uploading = false;
    this._registerEvents();

    $(window).resize(this.resize.bind(this));
    this.resize();
  }

  Uploads.prototype.delay = function (fn) {
    if (this.stopTimeout !== null) {
      clearTimeout(this.stopTimeout);
    }
    this.stopTimeout = setTimeout(function () {
      this.stopTimeout = null;
      fn();
    }.bind(this), 10);
  };

  Uploads.prototype.handleDragLeave = function (e) {
    if (this.uploading) {
      return;
    }
    e.preventDefault();
    this.delay(function () {
      window.app.circle.borderRegular();
    });
  };

  Uploads.prototype.handleDragOver = function (e) {
    if (this.uploading) {
      return;
    }
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'copy';
    }
    e.stopPropagation();
    e.preventDefault();
    this.delay(function () {
      window.app.circle.borderAnts();
    });
  };

  Uploads.prototype.handleDrop = function (e) {
    if (this.uploading) {
      return;
    }
    if (this.stopTimeout !== null) {
      clearTimeout(this.stopTimeout);
      this.stopTimeout = null;
    }
    e.stopPropagation();
    e.preventDefault();
    var files = e.dataTransfer.files;
    if (!files || !files.length) {
      window.app.circle.borderRegular();
    } else {
      window.app.circle.borderUploading();
      this.uploadFiles(files);
    }
  };

  Uploads.prototype.pickerDialog = function () {
    this.fileInput.click();
  };

  Uploads.prototype.resize = function () {
    $('#upload-view').css({ height: $(window).height() });
  };

  Uploads.prototype.uploadFiles = function (files) {
    this.uploading = true;
    var totalSize = 0;

    // Create the request.
    var formData = new FormData();
    for (var i = 0; i < files.length; ++i) {
      var file = files[i];
      totalSize += file.size || 0;
      formData.append('file-input', file, file.name);
    }

    var xhr = new XMLHttpRequest();

    // Show the progress around the circle.
    var handleProgress = function (e) {
      var total = e.total;
      if (!e.lengthComputable || !total) {
        total = totalSize;
      }
      if (total) {
        var percent = Math.min(e.loaded / total, 1);
        window.app.circle.animationInfo = percent;
        window.app.circle.draw();
      }
    };
    xhr.upload.addEventListener('progress', handleProgress, false);
    xhr.open('POST', '/upload', true);

    // When it's done, we need to make sure the upload succeeded.
    xhr.addEventListener('load', function (e) {
      this.uploading = false;
      window.app.circle.borderRegular();
      var value;
      try {
        var value = JSON.parse(xhr.response);
      } catch (e) {
        window.app.errorDialog('Invalid JSON data.');
        return;
      }
      if (value.error) {
        window.app.errorDialog('Failed to upload: ' + value.error);
      } else {
        window.location = '/files';
      }
    }.bind(this));

    // Handle basic connection errors.
    xhr.addEventListener('error', function () {
      this.uploading = false;
      window.app.circle.borderRegular();
      window.app.errorDialog('Failed to connect to the server.');
    }.bind(this));

    xhr.send(formData);
  };

  Uploads.prototype._registerEvents = function () {
    var elements = [$(document.body), $('#upload-view')];
    var dragOver = function (e) {
      this.handleDragOver(e.originalEvent);
    }.bind(this);
    var dragLeave = function (e) {
      this.handleDragLeave(e.originalEvent);
    }.bind(this);
    var drop = function (e) {
      this.handleDrop(e.originalEvent);
    }.bind(this);
    for (var i = 0, len = elements.length; i < len; ++i) {
      elements[i].bind('dragover', dragOver);
      elements[i].bind('dragleave', dragLeave);
      elements[i].bind('drop', drop);
    }
    this.fileInput.bind('change', function () {
      var files = this.fileInput[0].files;
      if (files && files.length) {
        window.app.circle.borderUploading();
        this.uploadFiles(files);
      }
    }.bind(this));
  };

  $(function () {
    window.app.uploads = new Uploads();
  });

  if (!window.app) {
    window.app = {};
  }

})();
