package dashboard

import (
	"html/template"
	"net/http"
	"fmt"
	"time"
)

type Page struct {
	Contributors *[]string
	Dashboard string
}

func ShowSeederInfo(page *Page) {
		http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Connection", "keep-alive")

			w.Header().Set("Access-Control-Allow-Origin", "*")

			rc := http.NewResponseController(w)
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					fmt.Fprintf(w, "event:update\ndata:%s\n\n", fmt.Sprint(*page.Contributors))

					rc.Flush()
				}
			}
		})
		http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			t, _ := template.ParseFiles("html/seeder.html")
			t.Execute(w, page)
		})
	http.ListenAndServe(page.Dashboard, nil)
}

func ShowNodeInfo(page *Page) {
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("html/node.html")
		t.Execute(w, page)
	})
	http.ListenAndServe(page.Dashboard, nil)
}
