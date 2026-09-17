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
var tmpl = template.Must(template.ParseFS(templateFS, "templates/*.html"))

//go:embed static
var staticFS embed.FS

type indexPageErrorMsg string // Values /?error= can be set to that will cause a relevant error message to appear on the index page

const (
	File        indexPageErrorMsg = "file"
	Read        indexPageErrorMsg = "read"
	BBC         indexPageErrorMsg = "bbc"
	NoUser      indexPageErrorMsg = "noUser"
	PrivateList indexPageErrorMsg = "privateList"
	Letterboxd  indexPageErrorMsg = "letterboxd"
)

func returnToIndex(w http.ResponseWriter, r *http.Request, err error, cause indexPageErrorMsg) {
	http.Redirect(w, r, fmt.Sprintf("/?err=%s", cause), 303)
	fmt.Println(err)
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

func index(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "index.html", r.URL.Query().Get("err"))
	if err != nil {
		fmt.Println(err)
	}
}

func browse(w http.ResponseWriter, r *http.Request) {
	films, err := getIPlayerFilms(1, Recent)
	if err != nil {
		returnToIndex(w, r, err, "bbc")
	}

	err = tmpl.ExecuteTemplate(w, "browse.html", films)
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

	err = tmpl.ExecuteTemplate(w, "results.html", foundFilms)
	if err != nil {
		fmt.Println(err)
	}
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

	err = tmpl.ExecuteTemplate(w, "results.html", foundFilms)
	if err != nil {
		fmt.Println(err)
	}
}

func server(port int) {
	// Server static files
	fs := http.FileServer(http.FS(staticFS))
	http.Handle("/static/", fs)

	// Routes
	http.HandleFunc("/", index)
	http.HandleFunc("/browse", browse)
	http.HandleFunc("/results/file", resultsFromFile)
	http.HandleFunc("/results/username", resultsFromUsername)

	url := fmt.Sprintf("http://localhost:%d/", port)
	go openInBrowser(url)

	fmt.Printf("Server listening on %s\n", url)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	fmt.Println(err)
}
