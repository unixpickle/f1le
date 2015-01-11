(function() {
  
  function Files() {
    this.element = $('#file-list');
    this.showList(window.app.loadedFiles);
    window.app.loadedFiles = null;
    $(window).resize(this.resize.bind(this));
    this.resize();
  }
  
  Files.prototype.resize = function() {
    this.element.css({top: $(window).height(),
      "min-height": $(window).height()});
  };
  
  Files.prototype.showList = function(list) {
    this.element.html('');
    for (var i = 0, len = list.length; i < len; ++i) {
      var fileDiv = document.createElement('div');
      fileDiv.className = 'file';
      
      var infoBox = document.createElement('div');
      infoBox.className = 'file-info';
      
      var nameField = document.createElement('label');
      nameField.className = 'file-name';
      $(nameField).text(list[i].name);
      
      var sizeField = document.createElement('label');
      sizeField.className = 'file-size';
      $(sizeField).text(list[i].size + 'bytes');
      
      var dateField = document.createElement('label');
      dateField.className = 'file-date';
      var date = new Date(0);
      console.log(list[i].uploaded);
      date.setUTCSeconds(list[i].uploaded);
      var month = date.getUTCMonth() + 1; //months from 1-12
      var day = date.getUTCDate();
      var year = date.getUTCFullYear();
      $(dateField).text(month + "/" + day + "/" + year);
      
      infoBox.appendChild(nameField);
      infoBox.appendChild(sizeField);
      infoBox.appendChild(dateField);
      fileDiv.appendChild(infoBox);
      
      this.element.append(fileDiv);
    }
  };
  
  $(function() {
    window.app.files = new Files();
  });
  
  if (!window.app) {
    window.app = {};
  }
  
})();