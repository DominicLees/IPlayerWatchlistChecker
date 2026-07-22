package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

const port int = 8000

func results(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Read() // skip header

	// Read titles from watchlist
	var watchlist []string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		watchlist = append(watchlist, row[1])
	}

	var foundFilms []string
	page := 1
	count := 0
	for {
		// Request next page of films
		resp, err := http.Get(fmt.Sprintf("https://ibl.api.bbci.co.uk/ibl/v1/categories/films/programmes?per_page=200&page=%d", page))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		// Decode JSON
		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}

		// Reduce to film data
		data := result["category_programmes"].(map[string]interface{})
		films := data["elements"].([]interface{})

		// Check for films on watchlist
		set := make(map[string]struct{}, len(films))
		for _, f := range films {
			filmObj := f.(map[string]interface{})
			set[filmObj["title"].(string)] = struct{}{}
		}

		for _, s := range watchlist {
			if _, found := set[s]; found {
				foundFilms = append(foundFilms, s)
			}
		}

		count += len(films)
		if count >= int(data["count"].(float64)) {
			break
		}
		page++
	}
	fmt.Fprint(w, foundFilms)
}

func index() http.HandlerFunc {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, r.URL.Query().Get("err"))
	}
}

func main() {
	http.HandleFunc("/", index())
	http.HandleFunc("/results", results)
	fmt.Printf("Server listening on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
