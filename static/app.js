function clear() {
	localStorage.clear();
};

var stream = new EventSource("events");

stream.addEventListener("seen", function(e) {
	var v = JSON.parse(e.data);
	var k = v.Class + "|" + v.Name;
	if (localStorage.getItem(k) === null) {
		localStorage.setItem(k, +v.Active);
	} else {
		localStorage.setItem(k, +localStorage.getItem(k) + v.Active);
	}
	document.getElementById("seen").innerHTML = v.Class + " | " + v.Name + " : " + v.Active;
});

stream.addEventListener("idle", function(e) {
	document.getElementById("idle").innerHTML = e.data;
});
