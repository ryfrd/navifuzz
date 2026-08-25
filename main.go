package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ryfrd/navifuzz/cmd"
)

var version = "0.1.0"

func main() {
	configPath, versionOnly, args := parseGlobals(os.Args[1:])
	if versionOnly {
		fmt.Printf("navifuzz %s\n", version)
		return
	}

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "help", "--help", "-h":
		printUsage()
		return
	case "version":
		fmt.Printf("navifuzz %s\n", version)
		return
	}

	runCommand(args[0], args[1:], configPath)
}

func parseGlobals(args []string) (configPath string, versionOnly bool, rest []string) {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--config":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --config requires a path")
				os.Exit(1)
			}
			configPath = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--config="):
			configPath = strings.TrimPrefix(args[i], "--config=")
		case args[i] == "--version" || args[i] == "-V" || args[i] == "-v":
			versionOnly = true
		default:
			rest = append(rest, args[i])
		}
	}
	return configPath, versionOnly, rest
}

func runCommand(name string, args []string, configPath string) {
	switch name {
	case "albums":
		runAlbums(args, configPath)
	case "artists":
		runArtists(args, configPath)
	case "songs":
		runSongs(args, configPath)
	case "playlists":
		runPlaylists(args, configPath)
	case "genres":
		runGenres(args, configPath)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command `%s`\n\n", name)
		printUsage()
		os.Exit(1)
	}
}

func runAlbums(args []string, configPath string) {
	fs := flag.NewFlagSet("albums", flag.ExitOnError)
	n := fs.Int("n", 0, "")
	help := fs.Bool("h", false, "")
	fs.BoolVar(help, "help", false, "")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), `Browse and play albums from Navidrome.

Usage: navifuzz albums [OPTIONS] [TYPE]

Arguments:
  [TYPE]  Album list type [default: newest] [possible values: newest, recent, frequent, random, starred, alphabetical, played]

Options:
  -n <N>          Number of albums to fetch [default: all]
  -h, --help      Print help`)
	}
	fs.Parse(args)
	if *help {
		fs.SetOutput(os.Stdout)
		fs.Usage()
		os.Exit(0)
	}

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

	if err := cmd.Albums(listType, *n, configPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runArtists(args []string, configPath string) {
	fs := flag.NewFlagSet("artists", flag.ExitOnError)
	help := fs.Bool("h", false, "")
	fs.BoolVar(help, "help", false, "")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), `Browse artists, then albums, then songs.

Usage: navifuzz artists [OPTIONS]

Options:
  -h, --help      Print help`)
	}
	fs.Parse(args)
	if *help {
		fs.SetOutput(os.Stdout)
		fs.Usage()
		os.Exit(0)
	}

	if err := cmd.Artists(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runSongs(args []string, configPath string) {
	fs := flag.NewFlagSet("songs", flag.ExitOnError)
	n := fs.Int("n", 0, "")
	help := fs.Bool("h", false, "")
	fs.BoolVar(help, "help", false, "")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), `Play random songs from the library.

Usage: navifuzz songs [OPTIONS]

Options:
  -n <N>          Number of songs to fetch [default: all]
  -h, --help      Print help`)
	}
	fs.Parse(args)
	if *help {
		fs.SetOutput(os.Stdout)
		fs.Usage()
		os.Exit(0)
	}

	if err := cmd.Songs(*n, configPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runPlaylists(args []string, configPath string) {
	fs := flag.NewFlagSet("playlists", flag.ExitOnError)
	help := fs.Bool("h", false, "")
	fs.BoolVar(help, "help", false, "")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), `Browse and play playlists from Navidrome.

Usage: navifuzz playlists [OPTIONS]

Options:
  -h, --help      Print help`)
	}
	fs.Parse(args)
	if *help {
		fs.SetOutput(os.Stdout)
		fs.Usage()
		os.Exit(0)
	}

	if err := cmd.Playlists(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runGenres(args []string, configPath string) {
	fs := flag.NewFlagSet("genres", flag.ExitOnError)
	n := fs.Int("n", 0, "")
	help := fs.Bool("h", false, "")
	fs.BoolVar(help, "help", false, "")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), `Browse genres and play songs from a selected genre.

Usage: navifuzz genres [OPTIONS]

Options:
  -n <N>          Number of songs to fetch [default: all]
  -h, --help      Print help`)
	}
	fs.Parse(args)
	if *help {
		fs.SetOutput(os.Stdout)
		fs.Usage()
		os.Exit(0)
	}

	if err := cmd.Genres(*n, configPath); err != nil {
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
  genres      Browse genres and play songs

Options:
      --config <PATH>     Path to config file [default: ~/.config/navifuzz/config.json]
  -h, --help              Print help
  -v, -V, --version       Print version`)
}
