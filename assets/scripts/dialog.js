(function() {
  
  if (!window.app) {
    window.app = {};
  }
  
  window.app.errorDialog = function(e) {
    alert(e);
  };
  
  window.app.confirm = function(msg, callback) {
    callback(confirm(msg));
  };
  
})();