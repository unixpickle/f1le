(function() {
  
  function Files() {
    this.element = $('#files');
  }
  
  Files.prototype.delete = function(id) {
    window.app.confirm('Delete this file?', function(doIt) {
      if (!doIt) {
        return;
      }
      
      // Remove the file from the DOM
      var elements = $('.file');
      for (var i = 0, len = elements.length; i < len; ++i) {
        if ($(elements[i]).find('.file-id').val() == id) {
          elements[i].remove();
          break;
        }
      }
      if (elements.length == 1) {
        $(document.body).append($('<div class="no-files">No files.</div>'))
      }
      
      $.ajax('/delete/' + id);
    })
  };
  
  Files.prototype.download = function(e) {
    
  };
  
  $(function() {
    window.app.files = new Files();
  });
  
  if (!window.app) {
    window.app = {};
  }
  
})();