function duration(ns) {
	var x = Math.floor(ns/1e9);
	var s = x % 60;
	x = Math.floor(x/60);
	var m = x % 60;
	x = Math.floor(x/60);
	var h = x % 24;
	var d = Math.floor(x/24);
	return ((d > 0) ? d + "d" : "") + ((h > 0) ? h + "h" : "") + ((m > 0) ? m + "m" : "") + s + "s";
};

function clearStorage() {
	tracks = new Array();
	localStorage.removeItem("tracks");
};

var tracks = new Array();

document.addEventListener('DOMContentLoaded', function() {
	tracks = JSON.parse(localStorage.getItem("tracks"));
	display(tracks);
});

function update(data) {
	var seen = false;
	var then = Date.now() - (8*60*60*1000);
	tracks = tracks.filter(function(item) {
		var lastSeen = Date.parse(item.Seen);
		return lastSeen > then;
	});
	tracks.map(function(item) {
		if (item.Class == data.Class && item.Name == data.Name) {
			seen = true;
			item.Active += data.Active;
			item.Seen = data.Seen;
		}
	});
	if (!seen) {
		tracks.push(data);
	}
	tracks.sort(function(a, b) {
		return a.Active - b.Active;
	});
	localStorage.setItem("tracks", JSON.stringify(tracks));
}

var stream = new EventSource("events");

function display(tracks) {
	var tab = document.getElementById("table");
	tab.innerHTML = "";
	tracks.map(function(item) {
		var row = tab.insertRow(0);
		row.insertCell(0).innerHTML = item.Class;
		row.insertCell(1).innerHTML = item.Name;
		row.insertCell(2).innerHTML = duration(item.Active);
		row.insertCell(3).innerHTML = item.Seen;
	});
};

stream.addEventListener("seen", function(e) {
	update(JSON.parse(e.data));
	display(tracks);
});

stream.addEventListener("idle", function(e) {
	document.getElementById("idle").innerHTML = e.data;
});
