package main

import (
	"flag"
)

func main() {
	port := flag.Int("port", 8000, "The port the server will listen on")
	filePath := flag.String("file", "", "The path to the watchlist csv file. Passing a file will run the check and output the result instead of spinning up the server.")
	flag.Parse()

	if *filePath == "" {
		server(*port)
	} else {
		cli(*filePath)
	}
}
