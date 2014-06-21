// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"
)

var tmpl *template.Template

func initTemplate(fname string) {
	tmpl = template.Must(template.ParseFiles(fname))
}

type Index struct {
	Title   string
	Records Records
	Classes Records
	Total   Duration
	Zzz     bool
}

type Records []Record
type Duration time.Duration

type Record struct {
	Class   string
	Name    string
	Spent   Duration
	Idle    Duration
	Seen    time.Time
	Percent float64
}

func (r Records) Len() int           { return len(r) }
func (r Records) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Records) Less(i, j int) bool { return r[i].Spent < r[j].Spent }

func (d Duration) String() string {
	h := int(time.Duration(d).Hours())
	m := int(time.Duration(d).Minutes()) % 60
	s := int(time.Duration(d).Seconds()) % 60
	var ret string
	if h > 0 {
		ret += fmt.Sprintf("%dh", h)
	}
	if m > 0 {
		ret += fmt.Sprintf("%dm", m)
	}
	return ret + fmt.Sprintf("%ds", s)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var idx Index
	idx.Title = "Gone Time Tracker"
	idx.Zzz = zzz
	class := r.URL.Path[1:]

	classtotal := make(map[string]time.Duration)

	for k, v := range tracks {
		classtotal[k.Class] += v.Spent
		idx.Total += Duration(v.Spent)
		if class != "" && class != k.Class {
			continue
		}
		idx.Records = append(idx.Records, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: Duration(v.Spent),
			Idle:  Duration(v.Idle)})
	}
	for k, v := range classtotal {
		idx.Classes = append(idx.Classes, Record{
			Class:   k,
			Spent:   Duration(v),
			Percent: 100.0 * float64(v) / float64(idx.Total)})
	}
	sort.Sort(sort.Reverse(idx.Classes))
	sort.Sort(sort.Reverse(idx.Records))
	err := tmpl.Execute(w, idx)
	if err != nil {
		log.Println(err)
	}
}

func dumpHandler(w http.ResponseWriter, r *http.Request) {
	var rec Records

	for k, v := range tracks {
		rec = append(rec, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: Duration(v.Spent),
			Seen:  v.Seen})
	}

	data, err := json.MarshalIndent(rec, "", "\t")
	if err != nil {
		log.Println("dump:", err)
	}
	w.Write(data)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	tracks.Remove(0)
	http.Redirect(w, r, "/", http.StatusFound)
}

func webReporter(port string) {
	log.Println("listen on", port)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/gone.json", dumpHandler)
	http.HandleFunc("/reset", resetHandler)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
