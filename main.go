package main

import (
	"flag"
)

func main() {
	port := flag.Int("port", 8000, "The port the server will listen on")
	flag.Parse()

	server(*port)
}
