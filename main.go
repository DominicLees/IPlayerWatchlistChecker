package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
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

		for _, film := range films {
			filmData := film.(map[string]interface{})
			count += 1
			fmt.Printf("%d: %s\n", count, filmData["title"].(string))
		}

		// count += len(films)
		if count >= int(data["count"].(float64)) {
			break
		}
		page++
	}
}
