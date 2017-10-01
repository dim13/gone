function clear() {
	localStorage.clear();
};

var stream = new EventSource("events");

stream.addEventListener("update", function(e) {
	document.getElementById("data").innerHTML = e.data;
});
