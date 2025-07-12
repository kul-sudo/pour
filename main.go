package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"os"
)

func main() {
	// config, err := bootstrap.ConfigGen()
	// if err != nil {
	// 	return
	// }

	// ffmpeg -i 19.mp4 -f segment -segment_time 10 -reset_timestamps 1 -c copy -segment_format_options movflags=frag_keyframe+empty_moov pour/chunks/%d.mp4
	go func() {
		lastNum := -1

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			files, _ := ioutil.ReadDir("chunks/")

			maxNum := -1
			maxFile := ""
			for _, f := range files {
				filename := strings.TrimSuffix(f.Name(), ".mp4")
				if num, err := strconv.Atoi(filename); err == nil {
					if num > maxNum {
						maxNum = num
						maxFile = f.Name()
					}
				}
			}

				fmt.Println(maxNum)

			if lastNum == maxNum {
				fmt.Println(5)
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				fmt.Println(maxNum)
				lastNum = maxNum
				fullName := filepath.Join("chunks", maxFile);
				file, err := os.Open(fullName)
				if err != nil {
					http.Error(w, "File not found", http.StatusNotFound)
					return
				}
				defer file.Close()

				fileInfo, err := file.Stat()
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
			}
		})
		http.ListenAndServe("localhost:8080", nil)
	}()

	go func() {
		http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			t, err := template.ParseFiles("html/node.html")
			if err != nil {
				fmt.Printf("failed to parse node dashboard html")
				return
			}
			t.Execute(w, nil)
		})
		// fs := http.FileServer(http.Dir("/home/user/pour/chunks"))
		// http.Handle("/chunks/", http.StripPrefix("/chunks", fs))
		http.ListenAndServe("localhost:8082", nil)

	}()

	for {

	}

	// switch config.Mode {
	// case "seeder":
	// 	seeder.Setup(&config)
	// case "node":
	// 	node.Setup(&config)
	// }
}
