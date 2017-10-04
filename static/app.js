function duration(ns) {
	var x = Math.floor(ns/1e9);
	var s = x % 60;
	x = Math.floor(x/60);
	var m = x % 60;
	x = Math.floor(x/60);
	var h = x % 24;
	var d = Math.floor(x/24);
	return ((d > 0) ? d + "d" : "") + ((h > 0) ? h + "h" : "") + ((m > 0) ? m + "m" : "") + s + "s";
}

function loadTracks() {
	var tracks = JSON.parse(localStorage.getItem("tracks"));
	if (tracks == null) {
		return new Array();
	}
	return removeOld(tracks, 8);
}

function storeTracks(tracks) {
	tracks.sort(function(a, b) {
		return b.Active - a.Active;
	});
	localStorage.setItem("tracks", JSON.stringify(tracks));
	return tracks;
}

function removeOld(tracks, h) {
	var t = Date.now() - (h * 60 * 60 * 1000);
	return tracks.filter(function(item) {
		return Date.parse(item.Seen) > t;
	});
}

function update(data) {
	var tracks = loadTracks()
	var seen = false;
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
	return storeTracks(tracks);
}

function display(tracks) {
	var table = document.createElement("table");
	var head = table.createTHead().insertRow(0);
	head.insertCell(0).innerHTML = "Class";
	head.insertCell(1).innerHTML = "Name";
	head.insertCell(2).innerHTML = "Spent";
	tracks.map(function(item) {
		var row = table.insertRow(-1);
		row.insertCell(0).innerHTML = item.Class;
		row.insertCell(1).innerHTML = item.Name;
		row.insertCell(2).innerHTML = duration(item.Active);
	});
	document.getElementById("table").innerHTML = table.innerHTML;
}

function clearStorage() {
	localStorage.clear();
}

document.addEventListener('DOMContentLoaded', function() {
	var tracks = loadTracks();
	display(tracks);
});

var stream = new EventSource("events");

stream.addEventListener("seen", function(e) {
	var tracks = update(JSON.parse(e.data));
	display(tracks);
});
