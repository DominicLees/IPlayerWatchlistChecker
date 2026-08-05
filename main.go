package main

import (
	"flag"
)

func main() {
	port := flag.Int("port", 8000, "The port the server will listen on")
	username := flag.String("username", "", "Your letterboxd username. Passing a username will run the check and output the result instead of spinning up the server.")
	filePath := flag.String("file", "", "The path to the watchlist csv file. Passing a file will run the check and output the result instead of spinning up the server or using your username to get your watchlist.")
	flag.Parse()

	if *filePath != "" {
		cliFromFile(*filePath)
	} else if *username != "" {
		cliFromUsername(*username)
	} else {
		server(*port)
	}
}
