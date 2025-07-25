package dashboard

import (
	"html/template"
	"net/http"
	"fmt"
	"time"
	"bytes"
)

type Page struct {
	Nodes *[]string
	LatestChunk *[]byte
	Dashboard string
}

func ShowSeederInfo(page *Page) {
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Connection", "keep-alive")

		rc := http.NewResponseController(w)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				fmt.Fprintf(w, "event:update\ndata:%s\n\n", fmt.Sprint(*page.Nodes))

				rc.Flush()
			}
		}
	})
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("html/seeder.html")
		if err != nil {
			fmt.Printf("failed to parse seeder dashboard html")
			return
		}
		t.Execute(w, page)
	})

	http.ListenAndServe(page.Dashboard, nil)
}

func ShowNodeInfo(page *Page) {
	http.HandleFunc("/chunk", func(w http.ResponseWriter, r *http.Request) {
		reader := bytes.NewReader(*page.LatestChunk)
		http.ServeContent(w, r, "video.mp4", time.Now(), reader)
	})
	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("html/node.html")
		if err != nil {
			fmt.Printf("failed to parse node dashboard html")
			return
		}
		t.Execute(w, page)
	})
	http.ListenAndServe(page.Dashboard, nil)
}
