// Gone Time Tracker -or- Where has my time gone?
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

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
	Class string
	Name  string
	Spent Duration
	Seen  time.Time
	Odd   bool `json:"-"`
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

func (d Duration) Seconds() int {
	return int(time.Duration(d).Seconds())
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var idx Index
	idx.Title = "Gone Time Tracker"
	idx.Zzz = zzz
	class := r.URL.Path[1:]

	classtotal := make(map[string]time.Duration)

	m.Lock()
	for k, v := range tracks {
		classtotal[k.Class] += v.Spent
		idx.Total += Duration(v.Spent)
		if class != "" && class != k.Class {
			continue
		}
		idx.Records = append(idx.Records, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: Duration(v.Spent)})
	}
	m.Unlock()
	for k, v := range classtotal {
		idx.Classes = append(idx.Classes, Record{Class: k, Spent: Duration(v)})
	}
	sort.Sort(sort.Reverse(idx.Classes))
	sort.Sort(sort.Reverse(idx.Records))
	for j := range idx.Records {
		idx.Records[j].Odd = j%2 == 0
	}
	err := tmpl.Execute(w, idx)
	if err != nil {
		log.Println(err)
	}
}

func dumpHandler(w http.ResponseWriter, r *http.Request) {
	var rec Records

	m.Lock()
	for k, v := range tracks {
		rec = append(rec, Record{
			Class: k.Class,
			Name:  k.Name,
			Spent: Duration(v.Spent),
			Seen:  v.Seen})
	}
	m.Unlock()

	data, err := json.MarshalIndent(rec, "", "\t")
	if err != nil {
		log.Println("dump:", err)
	}
	w.Write(data)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	tracks.cleanup(0)
	http.Redirect(w, r, "/", http.StatusFound)
}
