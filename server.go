package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func index() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, r.URL.Query().Get("err"))
	}
}

var resultsTmpl = template.Must(template.ParseFiles("templates/results.html"))

func resultsFromFile(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		cause := "read"
		if err == http.ErrMissingFile || err == http.ErrNotMultipart {
			cause = "file"
		}
		http.Redirect(w, r, fmt.Sprintf("/?err=%s", cause), 301)
		fmt.Println(err)
		return
	}
	defer file.Close()

	watchlist, err := readWatchlistFile(file, header.Filename)
	if err != nil {
		http.Redirect(w, r, "/?err=read", 301)
		fmt.Println(err)
		return
	}

	foundFilms, err := getIPlayerFilms(watchlist)
	if err != nil {
		http.Redirect(w, r, "/?err=bbc", 301)
		fmt.Println(err)
		return
	}

	resultsTmpl.Execute(w, foundFilms)
}

func server(port int) {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", index())
	http.HandleFunc("/results/file", resultsFromFile)
	// http.HandleFunc("/results/username", resultsFromUsername)

	fmt.Printf("Server listening on http://localhost:%d/\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	fmt.Println(err)
}
