function clearStorage() {
	localStorage.clear();
};

var tracks = new Array();

document.addEventListener('DOMContentLoaded', function() {
	tracks = JSON.parse(localStorage.getItem("tracks"));
});

function update(data) {
	for (var i in tracks) {
		if (tracks[i].Class == data.Class && tracks[i].Name == data.Name) {
			tracks[i].Active += data.Active;
			localStorage.setItem("tracks", JSON.stringify(tracks));
			return;
		}
	}
	tracks.push(data);
	localStorage.setItem("tracks", JSON.stringify(tracks));
}

var stream = new EventSource("events");

stream.addEventListener("seen", function(e) {
	var v = JSON.parse(e.data);
	update(v);
	tracks.sort(function(a, b) {
		return a.Active - b.Active;
	});
	var tab = document.getElementById("table");
	tab.innerHTML = "";
	for (var i in tracks) {
		var row = tab.insertRow(0);
		row.insertCell(0).innerHTML = tracks[i].Class;
		row.insertCell(1).innerHTML = tracks[i].Name;
		row.insertCell(2).innerHTML = tracks[i].Active;
		row.insertCell(3).innerHTML = tracks[i].Seen;
	}

	//document.getElementById("seen").innerHTML = v.Class + " | " + v.Name + " : " + v.Active;
	//document.getElementById("storage").innerHTML = localStorage.getItem("tracks");
});

stream.addEventListener("idle", function(e) {
	document.getElementById("idle").innerHTML = e.data;
});
