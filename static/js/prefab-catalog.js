function toggleHide() {
    var x = document.getElementById("hide");
    if (x.style.display === "none") {
      x.style.display = "block";
    } else { 
      x.style.display = "none";
    }
}

$('textarea').each(function () {
  this.setAttribute('style', 'height:' + (this.scrollHeight) + 'px;overflow-y:hidden;');
}).on('input', function () {
  this.style.height = 'auto';
  this.style.height = (this.scrollHeight) + 'px';
});