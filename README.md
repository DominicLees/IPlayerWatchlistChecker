# BBC iPlayer Film Availability Checker

This tool takes your Letterboxd watchlist and checks if any of the films are currently available on the BBC iPlayer. The checker can be used through the command line interface or by hosting the website locally. Your watchlist can either by passed in as a CSV file downloaded from Letterboxd or take your username and scrape your watchlist from your profile.

## Local Site

By default, the program will run a web server on port 8000. The port can be changed by setting the -port flag.

## CLI

By setting either the -username or -file flag, the program will instead run the check, output the result and then exit. If both flags are set, the check will only be performed using the passed file. 

## Flags

| Flag     | Type   | Default Value | Use                                               |
|----------|--------|---------------|---------------------------------------------------|
| port     | int    | 8000          | Set the port for the web server to listen to      |
| username | string | nil           | Letterboxd username to scrape watchlist from      |
| file     | string | nil           | Path to .csv file containing Letterboxd watchlist |