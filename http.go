// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"time"
)

type Index struct {
	Records Records
	Classes Classes
	Total   Duration
	Idle    Duration
	Zzz     bool
	Refresh time.Duration
}

type Record struct {
	Class string
	Name  string
	Spent Duration
	Idle  Duration
	Seen  time.Time
}

type Class struct {
	Class   string
	Spent   Duration
	Percent float64
}

type Records []Record

type Classes []Class

type Duration time.Duration

//go:embed static
var static embed.FS

func (r Records) Len() int           { return len(r) }
func (r Records) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Records) Less(i, j int) bool { return r[i].Spent < r[j].Spent }

func (c Classes) Len() int           { return len(c) }
func (c Classes) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Classes) Less(i, j int) bool { return c[i].Spent < c[j].Spent }

func (d Duration) String() string {
	return fmt.Sprint(time.Duration(d).Truncate(time.Second))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var idx Index
	idx.Zzz = zzz
	idx.Refresh = time.Minute // TODO use flag value
	class := r.URL.Path[1:]

	classes := make(map[string]time.Duration)

	for k, v := range tracks {
		classes[k.Class] += v.Spent
		idx.Total += Duration(v.Spent)
		idx.Idle += Duration(v.Idle)
		if class != "" && class != k.Class {
			continue
		}
		idx.Records = append(idx.Records, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: Duration(v.Spent),
			Idle:  Duration(v.Idle)})
	}
	for k, v := range classes {
		idx.Classes = append(idx.Classes, Class{
			Class:   k,
			Spent:   Duration(v),
			Percent: 100.0 * float64(v) / float64(idx.Total)})
	}
	sort.Sort(sort.Reverse(idx.Classes))
	sort.Sort(sort.Reverse(idx.Records))

	tmpl, err := template.ParseFS(static, "static/gone.tmpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.ExecuteTemplate(w, "root", idx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	tracks.RemoveSince(0)
	http.Redirect(w, r, "/", http.StatusFound)
}

func webReporter(port string) error {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/reset", resetHandler)
	return http.ListenAndServe(port, nil)
}
