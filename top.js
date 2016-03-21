/*document.body.onload = function(){
	showList("genre_all");
}*/
function changeGenre(e){
	showList(e.id);
}
function showList(id) {
	var ens = document.getElementById("works").getElementsByClassName("entry");
	if(ens.length === 0){console.log("not found: #works.entry");return;}
	var n = 0;
	for (var i = 0; i < ens.length; i++) {
		/*var col = n % 2 ? "#f8f8f8": "#ffffff";*/
		if (id === "genre_all" || ens[i].classList.contains(id)) {
			ens[i].style.display = "block";
			/*ens[i].style.backgroundColor = col;*/
			n++;
		} else {
			ens[i].style.display = "none";
		}
	}
}