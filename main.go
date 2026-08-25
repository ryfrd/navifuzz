package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/james/navifzf/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "albums":
		runAlbums(os.Args[2:])
	case "artists":
		runArtists(os.Args[2:])
	case "songs":
		runSongs(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runAlbums(args []string) {
	fs := flag.NewFlagSet("albums", flag.ExitOnError)
	n := fs.Int("n", 0, "number of albums to fetch (default: all)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: navifzf albums [options] [type]

Browse and play albums from Navidrome.

Arguments:
  type    Album list type (default: newest)

Types:
  newest         Recently added
  recent         Recently played
  frequent       Most frequently played
  random         Random albums
  starred        Starred/favorited albums
  alphabetical   Alphabetical order
  played         Most recently played

Options:`)
		fs.PrintDefaults()
	}
	fs.Parse(args)

	listType := "newest"
	if fs.NArg() > 0 {
		listType = fs.Arg(0)
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

	if err := cmd.Albums(listType, *n); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runArtists(args []string) {
	fs := flag.NewFlagSet("artists", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: navifzf artists

Browse artists, then albums, then songs.`)
	}
	fs.Parse(args)

	if err := cmd.Artists(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runSongs(args []string) {
	fs := flag.NewFlagSet("songs", flag.ExitOnError)
	n := fs.Int("n", 0, "number of songs to fetch (default: all)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: navifzf songs [options]

Play random songs from the library.

Options:`)
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if err := cmd.Songs(*n); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `navifzf - Navidrome CLI with fzf + mpv

Usage:
  navifzf <command> [options]

Commands:
  albums      Browse and play albums
  artists     Browse artists, albums, and songs
  songs       Play random songs

Run 'navifzf <command> --help' for command-specific help.

Config: ~/.config/navifzf/config.toml`)
}
