package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/ryfrd/navifuzz/api"
	"github.com/ryfrd/navifuzz/config"
)

func Albums(listType string, size int, opts Options) error {
	sess, err := newSession(opts)
	if err != nil {
		return err
	}
	client := sess.client
	cfg := sess.cfg

	fmt.Fprintln(os.Stderr, "Fetching albums...")
	albums, err := client.GetAlbumList2(listType, size)
	if err != nil {
		return fmt.Errorf("cannot fetch albums: %w", err)
	}

	if len(albums) == 0 {
		fmt.Fprintln(os.Stderr, "No albums found.")
		return nil
	}

	render, err := api.AlbumRenderer(cfg.AlbumFormat)
	if err != nil {
		return err
	}
	res, err := pick(albums, render, fmt.Sprintf("Play all (%d albums)", len(albums)), "Select album", cfg.Selector)
	if err != nil {
		return err
	}
	if !res.ok {
		return nil
	}

	if res.playAll {
		allSongs, err := fetchSongsForAlbums(client, albums)
		if err != nil {
			return err
		}
		if len(allSongs) == 0 {
			fmt.Fprintln(os.Stderr, "No songs found.")
			return nil
		}
		return playSongs(client, allSongs, cfg)
	}

	fmt.Fprintln(os.Stderr, "Fetching songs...")
	songs, err := client.GetAlbum(res.item.ID)
	if err != nil {
		return fmt.Errorf("cannot fetch album: %w", err)
	}

	if len(songs) == 0 {
		fmt.Fprintln(os.Stderr, "No songs in this album.")
		return nil
	}

	return selectSongs(client, songs, cfg)
}

func playSongs(client *api.Client, songs []api.Song, cfg *config.Config) error {
	binary := playerBinary(cfg.Player)
	if _, err := exec.LookPath(binary); err != nil {
		return fmt.Errorf("player %q not found: %w", binary, err)
	}

	var urls []string
	for _, s := range songs {
		urls = append(urls, client.StreamURL(s.ID))
	}

	tmpFile, err := os.CreateTemp("", "navifuzz-playlist-*.m3u")
	if err != nil {
		return fmt.Errorf("cannot create playlist: %w", err)
	}
	playlistFile := tmpFile.Name()
	tmpFile.Close()
	if err := os.WriteFile(playlistFile, []byte(strings.Join(urls, "\n")), 0600); err != nil {
		os.Remove(playlistFile)
		return fmt.Errorf("cannot write playlist: %w", err)
	}
	defer os.Remove(playlistFile)

	args := playerArgs(cfg.Player, playlistFile, cfg.Shuffle && len(songs) > 1, cfg.Loop && len(songs) > 1)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("player failed: %w", err)
	}

	return nil
}

func selectSongs(client *api.Client, songs []api.Song, cfg *config.Config) error {
	render, err := api.SongRenderer(cfg.SongFormat)
	if err != nil {
		return err
	}
	res, err := pick(songs, render, fmt.Sprintf("Play all (%d songs)", len(songs)), "Select song", cfg.Selector)
	if err != nil {
		return err
	}
	if !res.ok {
		return nil
	}

	if res.playAll {
		return playSongs(client, songs, cfg)
	}

	return playSongs(client, []api.Song{res.item}, cfg)
}

const maxConcurrency = 16

func fetchSongsForAlbums(client *api.Client, albums []api.Album) ([]api.Song, error) {
	sem := make(chan struct{}, maxConcurrency)
	results := make([][]api.Song, len(albums))
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	for i, a := range albums {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			songs, err := client.GetAlbum(a.ID)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("cannot fetch album: %w", err)
				}
				errMu.Unlock()
				return
			}
			results[i] = songs
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	var allSongs []api.Song
	for _, songs := range results {
		allSongs = append(allSongs, songs...)
	}
	return allSongs, nil
}

func fetchSongsForArtists(client *api.Client, artists []api.Artist) ([]api.Song, error) {
	sem := make(chan struct{}, maxConcurrency)
	albumByArtist := make([][]api.Album, len(artists))
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	for i, a := range artists {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			albums, err := client.GetArtist(a.ID)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("cannot fetch artist: %w", err)
				}
				errMu.Unlock()
				return
			}
			albumByArtist[i] = albums
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	var allAlbums []api.Album
	for _, albums := range albumByArtist {
		allAlbums = append(allAlbums, albums...)
	}
	return fetchSongsForAlbums(client, allAlbums)
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func runSelector(name, input, prompt string) (string, error) {
	if !isTerminal() && needsTTY(name) {
		return "", fmt.Errorf("%s requires a terminal: run navifuzz from a terminal, or use a GUI selector (dmenu, fuzzel, rofi)", name)
	}

	args := selectorArgs(name, prompt)
	if _, err := exec.LookPath(args[0]); err != nil {
		return "", fmt.Errorf("selector %q not found: %w", args[0], err)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && (exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130) {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// TUI selectors (fzf, sk) need a real TTY to render. GUI selectors (dmenu,
// fuzzel, rofi, tofi, wofi, bemenu) work fine when navifuzz itself has no TTY,
// eg. launched from a keybinding.
func needsTTY(name string) bool {
	switch name {
	case "dmenu", "fuzzel", "rofi", "tofi", "wofi", "bemenu":
		return false
	default: // fzf, sk, and unknown names, which fall back to fzf
		return true
	}
}

func indexOf(list []string, item string) int {
	for i, v := range list {
		if v == item {
			return i
		}
	}
	return -1
}

func makeUniqueDisplay(display []string) {
	used := make(map[string]bool, len(display))
	for i, d := range display {
		candidate := d
		for n := 2; used[candidate]; n++ {
			candidate = fmt.Sprintf("%s (%d)", d, n)
		}
		used[candidate] = true
		display[i] = candidate
	}
}

func playerBinary(name string) string {
	switch name {
	case "vlc", "cvlc":
		return name
	default:
		return "mpv"
	}
}

func playerArgs(name, playlist string, shuffle, loop bool) []string {
	switch name {
	case "vlc", "cvlc":
		args := []string{name}
		if shuffle {
			args = append(args, "--random")
		}
		if loop {
			args = append(args, "--loop")
		}
		args = append(args, playlist)
		return args
	default: // mpv
		args := []string{"mpv"}
		if shuffle {
			args = append(args, "--shuffle")
		}
		if loop {
			args = append(args, "--loop-playlist=yes")
		}
		args = append(args, "--playlist="+playlist)
		return args
	}
}

func selectorArgs(name, prompt string) []string {
	switch name {
	case "dmenu":
		return []string{"dmenu", "-p", prompt + " > "}
	case "fuzzel":
		return []string{"fuzzel", "--dmenu", "-p", prompt + " > "}
	case "rofi":
		return []string{"rofi", "-dmenu", "-p", prompt}
	case "tofi":
		return []string{"tofi", "--prompt-text=" + prompt + " > "}
	case "wofi":
		return []string{"wofi", "--dmenu", "-p", prompt}
	case "bemenu":
		return []string{"bemenu", "-p", prompt + " > "}
	case "sk":
		return []string{"sk", "--prompt", prompt + " > "}
	default: // fzf
		return []string{"fzf", "--prompt", prompt + " > "}
	}
}
