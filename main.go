package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

const port int = 8000

type IPlayerFilm struct {
	Title string
	Id    string
}

func results() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("templates/results.html"))

	return func(w http.ResponseWriter, r *http.Request) {
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

		// Check csv file was sent
		if strings.ToLower(filepath.Ext(header.Filename)) != ".csv" {
			http.Redirect(w, r, "/?err=file", 301)
			return
		}

		reader := csv.NewReader(file)
		reader.Read() // Skip header

		// Read film titles from watchlist files
		var watchlist []string
		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Redirect(w, r, "/?err=read", 301)
				return
			}
			watchlist = append(watchlist, row[1])
		}

		var foundFilms []IPlayerFilm
		page := 1
		count := 0
		for {
			// Request next page of films
			resp, err := http.Get(fmt.Sprintf("https://ibl.api.bbci.co.uk/ibl/v1/categories/films/programmes?per_page=200&page=%d", page))
			if err != nil {
				http.Redirect(w, r, "/?err=bbc", 301)
				return
			}
			defer resp.Body.Close()

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				http.Redirect(w, r, "/?err=bbc", 301)
				return
			}

			// Decode JSON
			var result map[string]interface{}
			err = json.Unmarshal(body, &result)
			if err != nil {
				http.Redirect(w, r, "/?err=bbc", 301)
				return
			}

			// Reduce to film data
			data := result["category_programmes"].(map[string]interface{})
			films := data["elements"].([]interface{})

			// Check for films on watchlist
			watchSet := make(map[string]struct{}, len(watchlist))
			for _, w := range watchlist {
				watchSet[w] = struct{}{}
			}

			for _, f := range films {
				filmObj := f.(map[string]interface{})
				title := filmObj["title"].(string)
				if _, found := watchSet[title]; !found {
					continue
				}
				id := filmObj["id"].(string)
				foundFilms = append(foundFilms, IPlayerFilm{Title: title, Id: id})
			}

			count += len(films)
			if count >= int(data["count"].(float64)) {
				break
			}
			page++
		}

		tmpl.Execute(w, foundFilms)
	}
}

func index() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, r.URL.Query().Get("err"))
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", index())
	http.HandleFunc("/results", results())
	fmt.Printf("Server listening on http://localhost:%d/\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
