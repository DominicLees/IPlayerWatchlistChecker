package main

import (
	"fmt"
	"os"
)

func printFoundFilms(watchlist []string) {
	films, err := getIPlayerFilms(watchlist)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d films found\n", len(films))
	for _, film := range films {
		fmt.Printf("%s: https://www.bbc.co.uk/iplayer/episodes/%s\n", film.Title, film.Id)
	}
}

func cliFromFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	watchlist, err := readWatchlistFile(file, path)
	if err != nil {
		panic(err)
	}

	printFoundFilms(watchlist)
}

func cliFromUsername(username string) {
	watchlist, err := getLetterboxdWatchlist(username)
	if err != nil {
		panic(err)
	}

	printFoundFilms(watchlist)
}
