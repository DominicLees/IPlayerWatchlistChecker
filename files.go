package main

import (
	"encoding/csv"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
)

type ErrFileNotCSV struct {
	message string
}

func (e *ErrFileNotCSV) Error() string {
	return e.message
}

type ErrFileNotListOfFilms struct {
	message string
}

func (e *ErrFileNotListOfFilms) Error() string {
	return e.message
}

func readWatchlistFile(file multipart.File, filename string) ([]string, error) {
	// Check csv file was passed
	if strings.ToLower(filepath.Ext(filename)) != ".csv" {
		return nil, &ErrFileNotCSV{"File does not have .csv extension"}
	}

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	// Check header layout matches that of a CSV file provided by Letterboxd
	if len(header) != 4 || header[3] != "Letterboxd URI" {
		return nil, &ErrFileNotListOfFilms{"CSV contents not formatted as a list of films"}
	}

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
