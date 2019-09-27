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

function removeOld(tracks, h) {
	var t = Date.now() - (h * 60 * 60 * 1000);
	return tracks.filter(function(item) {
		return Date.parse(item.Seen) > t;
	});
}

function update(tracks, data) {
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
}

function replace(element, content) {
	while (element.hasChildNodes()) {
		element.removeChild(element.lastChild);
	}
	element.appendChild(content);
}

function records(tracks) {
	var table = document.createElement("table");
	table.className = "table table-striped";
	var head = table.createTHead().insertRow();
	head.insertCell().innerHTML = "Class";
	head.insertCell().innerHTML = "Name";
	head.insertCell().innerHTML = "Spent";
	tracks.map(function(item) {
		var row = table.insertRow(-1);
		row.insertCell().innerHTML = item.Class;
		row.insertCell().innerHTML = item.Name;
		row.insertCell().innerHTML = duration(item.Active);
	});
	replace(document.getElementById("records"), table);
}

function classes(tracks) {
	var table = document.createElement("table");
	var total = 0;
	var m = new Map();
	tracks.map(function(item) {
		var v = m.get(item.Class);
		v = v ? v : 0;
		m.set(item.Class, v + item.Active);
		total += item.Active;
	});
	m.forEach(function(value, key) {
		var row = table.insertRow(-1);
		row.insertCell().innerHTML = key;
		row.insertCell().innerHTML = duration(value);
	});
	var totalRow = table.insertRow(-1);
	totalRow.insertCell().innerHTML = "&Sigma;";
	var d = totalRow.insertCell(1);
	d.id = "total";
	d.innerHTML = duration(total);
	replace(document.getElementById("classes"), table);
}

function display(tracks) {
	records(tracks);
	classes(tracks);
}

var stream = new EventSource("events");
var tracks = new Array();

stream.addEventListener("seen", function(e) {
	update(tracks, JSON.parse(e.data));
	display(tracks);
});

document.addEventListener('DOMContentLoaded', function() {
	display(tracks);
});
