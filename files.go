package main

import (
	"encoding/csv"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
)

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
