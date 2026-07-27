package main

import (
	"fmt"
	"os"
)

func cli(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	watchlist, err := readWatchlistFile(file, path)
	if err != nil {
		panic(err)
	}

	foundFilms, err := getFilms(watchlist)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d films found\n", len(foundFilms))
	for _, film := range foundFilms {
		fmt.Printf("%s: https://www.bbc.co.uk/iplayer/episodes/%s\n", film.Title, film.Id)
	}
}
