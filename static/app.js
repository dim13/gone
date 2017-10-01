var stream = new EventSource("events");

stream.addEventListener("update", function(e) {
	// ... e.data
});
