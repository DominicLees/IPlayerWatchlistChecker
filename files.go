package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

type IPlayerFilm struct {
	Title string
	Id    string
}

type ErrNotCSV struct {
	message string
}

func (e *ErrNotCSV) Error() string {
	return e.message
}

func readWatchlistFile(file multipart.File, header *multipart.FileHeader) ([]string, error) {
	// Check csv file was passed
	if strings.ToLower(filepath.Ext(header.Filename)) != ".csv" {
		return nil, &ErrNotCSV{"File does not have .csv extension"}
	}

	reader := csv.NewReader(file)
	reader.Read() // Skip header

	// Read film titles from watchlist file
	var watchlist []string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		watchlist = append(watchlist, row[1])
	}

	return watchlist, nil
}

func getFilms(watchlist []string) ([]IPlayerFilm, error) {
	var foundFilms []IPlayerFilm
	page := 1
	count := 0
	for {
		// Request next page of films
		resp, err := http.Get(fmt.Sprintf("https://ibl.api.bbci.co.uk/ibl/v1/categories/films/programmes?per_page=200&page=%d", page))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// Decode JSON
		var result map[string]interface{}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return nil, err
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

	return foundFilms, nil
}
