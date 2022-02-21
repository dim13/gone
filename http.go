// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/exp/slices"
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

var tmpl *template.Template

//go:embed static
var static embed.FS

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
			Idle:  Duration(v.Idle),
		})
	}
	slices.SortFunc(idx.Records, func(a, b Record) bool { return a.Spent > b.Spent })

	for k, v := range classes {
		idx.Classes = append(idx.Classes, Class{
			Class:   k,
			Spent:   Duration(v),
			Percent: 100.0 * float64(v) / float64(idx.Total),
		})
	}
	slices.SortFunc(idx.Classes, func(a, b Class) bool { return a.Spent > b.Spent })

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
