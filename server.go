package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func returnToIndex(w http.ResponseWriter, r *http.Request, err error, cause string) {
	http.Redirect(w, r, fmt.Sprintf("/?err=%s", cause), 303)
	fmt.Println(err)
}

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
		returnToIndex(w, r, err, cause)
		return
	}
	defer file.Close()

	watchlist, err := readWatchlistFile(file, header.Filename)
	if err != nil {
		returnToIndex(w, r, err, "read")
		return
	}

	foundFilms, err := getIPlayerFilms(watchlist)
	if err != nil {
		returnToIndex(w, r, err, "bbc")
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
