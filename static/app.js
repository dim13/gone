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
	tracks = tracks ? tracks : new Array();
	return removeOld(tracks, 8);
}

function storeTracks(tracks) {
	localStorage.setItem("tracks", JSON.stringify(tracks));
	return tracks;
}

function removeOld(tracks, h) {
	var t = Date.now() - (h * 60 * 60 * 1000);
	return tracks.filter(function(item) {
		return Date.parse(item.Seen) > t;
	});
}

function overview(tracks) {
	var m = new Map();
	tracks.map(function(item) {
		var v = m.get(item.Class);
		v = v ? v : 0;
		m.set(item.Class, v + item.Active);
	});
	return m;
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
	tracks.sort(function(a, b) {
		return b.Active - a.Active;
	});
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
	var records = document.getElementById("records");
	while (records.hasChildNodes()) {
		records.removeChild(records.lastChild);
	}
	records.appendChild(table);

	var table = document.createElement("table");
	var classMap = overview(tracks);
	var total = 0;
	classMap.forEach(function(value, key) {
		var row = table.insertRow(-1);
		row.insertCell(0).innerHTML = key;
		row.insertCell(1).innerHTML = duration(value);
		total += value;
	});
	var totalRow = table.insertRow(-1);
	totalRow.insertCell(0).innerHTML = "Total";
	var d = totalRow.insertCell(1);
	d.id = "total";
	d.innerHTML = duration(total);
	var classes = document.getElementById("classes");
	while (classes.hasChildNodes()) {
		classes.removeChild(classes.lastChild);
	}
	classes.appendChild(table);
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
