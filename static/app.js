function clear() {
	localStorage.clear();
};

var stream = new EventSource("events");

stream.addEventListener("seen", function(e) {
	document.getElementById("seen").innerHTML += e.data + "<br>";
});

stream.addEventListener("idle", function(e) {
	document.getElementById("idle").innerHTML += e.data + "<br>";
});
