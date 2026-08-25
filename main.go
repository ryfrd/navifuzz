package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/james/navifuzz/cmd"
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
	case "playlists":
		runPlaylists(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command `%s`\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func runAlbums(args []string) {
	fs := flag.NewFlagSet("albums", flag.ExitOnError)
	n := fs.Int("n", 0, "")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Browse and play albums from Navidrome.

Usage: navifuzz albums [OPTIONS] [TYPE]

Arguments:
  [TYPE]  Album list type [default: newest] [possible values: newest, recent, frequent, random, starred, alphabetical, played]

Options:
  -n <N>          Number of albums to fetch [default: all]
  -h, --help      Print help`)
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
		fmt.Fprintf(os.Stderr, "error: invalid album list type `%s`\n\n", listType)
		fmt.Fprintln(os.Stderr, "possible values: newest, recent, frequent, random, starred, alphabetical, played")
		os.Exit(1)
	}

	if err := cmd.Albums(listType, *n); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runArtists(args []string) {
	fs := flag.NewFlagSet("artists", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Browse artists, then albums, then songs.

Usage: navifuzz artists [OPTIONS]

Options:
  -h, --help      Print help`)
	}
	fs.Parse(args)

	if err := cmd.Artists(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runSongs(args []string) {
	fs := flag.NewFlagSet("songs", flag.ExitOnError)
	n := fs.Int("n", 0, "")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Play random songs from the library.

Usage: navifuzz songs [OPTIONS]

Options:
  -n <N>          Number of songs to fetch [default: all]
  -h, --help      Print help`)
	}
	fs.Parse(args)

	if err := cmd.Songs(*n); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runPlaylists(args []string) {
	fs := flag.NewFlagSet("playlists", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `Browse and play playlists from Navidrome.

Usage: navifuzz playlists [OPTIONS]

Options:
  -h, --help      Print help`)
	}
	fs.Parse(args)

	if err := cmd.Playlists(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: navifuzz [OPTIONS] <COMMAND>

Commands:
  albums      Browse and play albums
  artists     Browse artists, albums, and songs
  songs       Play random songs
  playlists   Browse and play playlists

Options:
  -h, --help      Print help

Config: ~/.config/navifuzz/config.json`)
}
