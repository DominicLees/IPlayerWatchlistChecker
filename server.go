package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
)

//go:embed templates/*.html
var templateFS embed.FS
var resultsTmpl = template.Must(template.ParseFS(templateFS, "templates/results.html"))

//go:embed static
var staticFS embed.FS

func returnToIndex(w http.ResponseWriter, r *http.Request, err error, cause string) {
	http.Redirect(w, r, fmt.Sprintf("/?err=%s", cause), 303)
	fmt.Println(err)
}

func index() http.HandlerFunc {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/index.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, r.URL.Query().Get("err"))
	}
}

func openInBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		fmt.Println(err)
	}
}

func resultsFromFile(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err == http.ErrMissingFile || err == http.ErrNotMultipart {
		returnToIndex(w, r, err, "file")
		return
	} else if err != nil {
		returnToIndex(w, r, err, "read")
		return
	}
	defer file.Close()

	watchlist, err := readWatchlistFile(file, header.Filename)
	if err != nil {
		returnToIndex(w, r, err, "read")
		return
	}

	foundFilms, err := getIPlayerFilmsOnWatchlist(watchlist)
	if err != nil {
		returnToIndex(w, r, err, "bbc")
		return
	}

	resultsTmpl.Execute(w, foundFilms)
}

func resultsFromUsername(w http.ResponseWriter, r *http.Request) {
	watchlist, err := getLetterboxdWatchlist(r.FormValue("username"))
	if err != nil {
		if _, ok := errors.AsType[*ErrUserDoesNotExist](err); ok {
			returnToIndex(w, r, err, "noUser")
		} else if _, ok := errors.AsType[*ErrUserWatchlistPrivate](err); ok {
			returnToIndex(w, r, err, "privateList")
		} else {
			returnToIndex(w, r, err, "letterboxd")
		}
		return
	}

	foundFilms, err := getIPlayerFilmsOnWatchlist(watchlist)
	if err != nil {
		returnToIndex(w, r, err, "bbc")
		return
	}

	resultsTmpl.Execute(w, foundFilms)
}

func server(port int) {
	fs := http.FileServer(http.FS(staticFS))
	http.Handle("/static/", fs)
	http.HandleFunc("/", index())
	http.HandleFunc("/results/file", resultsFromFile)
	http.HandleFunc("/results/username", resultsFromUsername)

	url := fmt.Sprintf("http://localhost:%d/", port)
	go openInBrowser(url)

	fmt.Printf("Server listening on %s\n", url)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	fmt.Println(err)
}
