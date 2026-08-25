package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/james/navifzf/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "albums":
		args := os.Args[2:]
		listType := "newest"
		size := 0

		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "-n":
				if i+1 < len(args) {
					n, err := strconv.Atoi(args[i+1])
					if err != nil || n < 1 {
						fmt.Fprintf(os.Stderr, "Invalid size: %s\n", args[i+1])
						os.Exit(1)
					}
					size = n
					i++
				} else {
					fmt.Fprintln(os.Stderr, "-n requires a number")
					os.Exit(1)
				}
			default:
				listType = args[i]
			}
		}

		validTypes := map[string]bool{
			"newest": true, "recent": true, "frequent": true,
			"random": true, "starred": true, "alphabetical": true,
			"played": true,
		}
		if !validTypes[listType] {
			fmt.Fprintf(os.Stderr, "Invalid album list type: %s\n", listType)
			fmt.Fprintf(os.Stderr, "Valid types: newest, recent, frequent, random, starred, alphabetical, played\n")
			os.Exit(1)
		}

		if err := cmd.Albums(listType, size); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `navifzf - Navidrome CLI with fzf + mpv

Usage:
  navifzf albums [type] [-n size]    Browse and play albums

Options:
  -n <number>  Number of albums to fetch (default: server default)

Album types:
  newest         Recently added (default)
  recent         Recently played
  frequent       Most frequently played
  random         Random albums
  starred        Starred/favorited albums
  alphabetical   Alphabetical order
  played         Most recently played

Config: ~/.config/navifzf/config.toml

Example config:
  server   = "https://navidrome.example.com"
  username = "user"
  password = "pass"
  scrobble = true
  player   = "mpv"
  selector = "fzf"
  shuffle  = true
  loop     = true

Accepted players: mpv, vlc, cvlc
Accepted selectors: fzf, dmenu, rofi, fuzzel`)
}
