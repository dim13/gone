// Gone Time Tracker -or- Where has my time gone?
package main

import (
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
	ID    int
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

func init() {
	t := FSMustString(false, "/static/gone.tmpl")
	tmpl = template.Must(template.New("").Parse(t))
}

func (r Records) Len() int           { return len(r) }
func (r Records) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Records) Less(i, j int) bool { return r[i].Spent < r[j].Spent }

func (c Classes) Len() int           { return len(c) }
func (c Classes) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Classes) Less(i, j int) bool { return c[i].Spent < c[j].Spent }

func (d Duration) String() string {
	return fmt.Sprint(time.Duration(d).Truncate(time.Second))
}

func (t Tracks) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var idx Index
	idx.Zzz = t.zzz
	idx.Refresh = t.interval
	class := r.URL.Path[1:]

	classes := make(map[string]time.Duration)

	for k, v := range t.tracks {
		classes[k.Class] += v.Spent
		idx.Total += Duration(v.Spent)
		idx.Idle += Duration(v.Idle)
		if class != "" && class != k.Class {
			continue
		}
		idx.Records = append(idx.Records, Record{
			ID:    k.ID,
			Class: k.Class,
			Name:  k.Name,
			Spent: Duration(v.Spent),
			Idle:  Duration(v.Idle),
		})
	}
	for k, v := range classes {
		total := idx.Total
		if total == 0 {
			total = 1
		}
		idx.Classes = append(idx.Classes, Class{
			Class:   k,
			Spent:   Duration(v),
			Percent: 100.0 * float64(v) / float64(total),
		})
	}
	sort.Sort(sort.Reverse(idx.Classes))
	sort.Sort(sort.Reverse(idx.Records))
	err := tmpl.ExecuteTemplate(w, "root", idx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func webReporter(t *Tracks, port string) error {
	http.Handle("/", t)
	return http.ListenAndServe(port, nil)
}
