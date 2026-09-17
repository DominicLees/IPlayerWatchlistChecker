package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type IPlayerFilm struct {
	Title string
	Id    string
}

type ErrUserDoesNotExist struct {
	message string
}

func (e *ErrUserDoesNotExist) Error() string {
	return e.message
}

type ErrUserWatchlistPrivate struct {
	message string
}

func (e *ErrUserWatchlistPrivate) Error() string {
	return e.message
}

type ErrRateLimited struct {
	message string
}

func (e *ErrRateLimited) Error() string {
	return e.message
}

type ErrResponseCode struct {
	message string
}

func (e *ErrResponseCode) Error() string {
	return e.message
}

var itemNameRegExp = regexp.MustCompile(`data-item-name="([^"]*)"`)
var trailingYearRegExp = regexp.MustCompile(`\s\(\d{4}\)$`)

func getIPlayerFilmsOnWatchlist(watchlist []string) ([]IPlayerFilm, error) {
	var foundFilms []IPlayerFilm
	page := 1
	count := 0
	for {
		// Request next page of films
		resp, err := http.Get(fmt.Sprintf("https://ibl.api.bbci.co.uk/ibl/v1/categories/films/programmes?per_page=200&page=%d", page))
		if err != nil {
			return nil, err
		}
		switch {
		case resp.StatusCode == 429:
			errMsg := fmt.Sprintf("IPlayer rate limit reached. Retry after %ss", resp.Header.Get("Retry-After"))
			resp.Body.Close()
			return nil, &ErrRateLimited{message: errMsg}
		case resp.StatusCode != 200:
			errMsg := fmt.Sprintf("Unexpected %d response code", resp.StatusCode)
			resp.Body.Close()
			return nil, &ErrResponseCode{message: errMsg}
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
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

func getIPlayerFilms(page int) ([]IPlayerFilm, error) {
	resp, err := http.Get(fmt.Sprintf("https://ibl.api.bbci.co.uk/ibl/v1/categories/films/programmes?per_page=50&page=%d", page))
	if err != nil {
		return nil, err
	}
	switch {
	case resp.StatusCode == 429:
		errMsg := fmt.Sprintf("IPlayer rate limit reached. Retry after %ss", resp.Header.Get("Retry-After"))
		resp.Body.Close()
		return nil, &ErrRateLimited{message: errMsg}
	case resp.StatusCode != 200:
		errMsg := fmt.Sprintf("Unexpected %d response code", resp.StatusCode)
		resp.Body.Close()
		return nil, &ErrResponseCode{message: errMsg}
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
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

	// Convert JSON film data to IPlayerFilm array
	var foundFilms []IPlayerFilm
	for _, f := range films {
		filmObj := f.(map[string]interface{})
		title := filmObj["title"].(string)
		fmt.Println(title)
		id := filmObj["id"].(string)
		foundFilms = append(foundFilms, IPlayerFilm{Title: title, Id: id})
	}

	return foundFilms, nil
}

func getLetterboxdWatchlist(username string) ([]string, error) {
	var films []string
	page := 1

	for {
		// Request next page of user's watchlist
		url := fmt.Sprintf("https://letterboxd.com/%s/watchlist/page/%d/", username, page)

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}

		switch {
		case resp.StatusCode == 404:
			resp.Body.Close()
			return nil, &ErrUserDoesNotExist{message: "User does not exist"}
		case resp.StatusCode == 403:
			resp.Body.Close()
			return nil, &ErrUserWatchlistPrivate{message: "User's watchlist is private"}
		case resp.StatusCode == 429:
			errMsg := fmt.Sprintf("Letterboxd rate limit reached. Retry after %ss", resp.Header.Get("Retry-After"))
			resp.Body.Close()
			return nil, &ErrRateLimited{message: errMsg}
		case resp.StatusCode != 200:
			resp.Body.Close()
			return nil, &ErrResponseCode{message: "Unexpected response"}
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// Get films from html
		matches := itemNameRegExp.FindAllStringSubmatch(string(body), -1)
		if len(matches) == 0 {
			// Either we've gone past the last page, or the watchlist is empty.
			break
		}

		for _, m := range matches {
			// Remove year from end of film title
			title := trailingYearRegExp.ReplaceAllString(m[1], "")
			films = append(films, title)
		}

		page++
	}

	return films, nil
}
