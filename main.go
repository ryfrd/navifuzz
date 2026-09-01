package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ryfrd/navifuzz/cmd"
)

var version = "0.1.0"

func main() {
	opts, versionOnly, args := parseGlobals(os.Args[1:])
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

	runCommand(args[0], args[1:], opts)
}

func parseGlobals(args []string) (opts cmd.Options, versionOnly bool, rest []string) {
	var shuffle, loop *bool
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--config":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --config requires a path")
				os.Exit(1)
			}
			opts.ConfigPath = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--config="):
			opts.ConfigPath = strings.TrimPrefix(args[i], "--config=")
		case args[i] == "--selector":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --selector requires a value")
				os.Exit(1)
			}
			opts.Selector = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--selector="):
			opts.Selector = strings.TrimPrefix(args[i], "--selector=")
		case args[i] == "--player":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --player requires a value")
				os.Exit(1)
			}
			opts.Player = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--player="):
			opts.Player = strings.TrimPrefix(args[i], "--player=")
		case args[i] == "--shuffle":
			v := true
			shuffle = &v
		case args[i] == "--no-shuffle":
			v := false
			shuffle = &v
		case strings.HasPrefix(args[i], "--shuffle="):
			v, err := strconv.ParseBool(strings.TrimPrefix(args[i], "--shuffle="))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: invalid value for --shuffle: %s\n", strings.TrimPrefix(args[i], "--shuffle="))
				os.Exit(1)
			}
			shuffle = &v
		case args[i] == "--loop":
			v := true
			loop = &v
		case args[i] == "--no-loop":
			v := false
			loop = &v
		case strings.HasPrefix(args[i], "--loop="):
			v, err := strconv.ParseBool(strings.TrimPrefix(args[i], "--loop="))
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: invalid value for --loop: %s\n", strings.TrimPrefix(args[i], "--loop="))
				os.Exit(1)
			}
			loop = &v
		case args[i] == "--version" || args[i] == "-V" || args[i] == "-v":
			versionOnly = true
		default:
			rest = append(rest, args[i])
		}
	}
	opts.Shuffle = shuffle
	opts.Loop = loop
	return opts, versionOnly, rest
}

func runCommand(name string, args []string, opts cmd.Options) {
	switch name {
	case "albums":
		runAlbums(args, opts)
	case "artists":
		runArtists(args, opts)
	case "songs":
		runSongs(args, opts)
	case "playlists":
		runPlaylists(args, opts)
	case "genres":
		runGenres(args, opts)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown command `%s`\n\n", name)
		printUsage()
		os.Exit(1)
	}
}

func runAlbums(args []string, opts cmd.Options) {
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

	if err := cmd.Albums(listType, *n, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runArtists(args []string, opts cmd.Options) {
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

	if err := cmd.Artists(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runSongs(args []string, opts cmd.Options) {
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

	if err := cmd.Songs(*n, opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runPlaylists(args []string, opts cmd.Options) {
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

	if err := cmd.Playlists(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runGenres(args []string, opts cmd.Options) {
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

	if err := cmd.Genres(*n, opts); err != nil {
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
      --selector <NAME>   Override the selector from the config file (fzf, dmenu, rofi, fuzzel, tofi, wofi, bemenu, sk)
      --player <NAME>     Override the player from the config file (mpv, vlc)
      --shuffle[=BOOL]    Override shuffle from the config file (also --no-shuffle)
      --loop[=BOOL]       Override loop from the config file (also --no-loop)
  -h, --help              Print help
  -v, -V, --version       Print version`)
}
