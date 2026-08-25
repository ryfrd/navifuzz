package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/james/navifuzz/api"
	"github.com/james/navifuzz/config"
)

func Albums(listType string, size int) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.Server, cfg.Username, cfg.Password)

	fmt.Fprintln(os.Stderr, "Connecting to Navidrome...")
	if err := client.Ping(); err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Fetching albums...")
	albums, err := client.GetAlbumList2(listType, size)
	if err != nil {
		return fmt.Errorf("cannot fetch albums: %w", err)
	}

	if len(albums) == 0 {
		fmt.Fprintln(os.Stderr, "No albums found.")
		return nil
	}

	albumIDs := make([]string, 0, len(albums))
	albumDisplay := make([]string, 0, len(albums)+1)
	albumIDs = append(albumIDs, "__PLAY_ALL__")
	albumDisplay = append(albumDisplay, fmt.Sprintf("Play all (%d albums)", len(albums)))
	for _, a := range albums {
		albumIDs = append(albumIDs, a.ID)
		display, err := api.RenderAlbum(a, cfg.AlbumFormat)
		if err != nil {
			return err
		}
		albumDisplay = append(albumDisplay, display)
	}

	selected, err := runSelector(cfg.Selector, strings.Join(albumDisplay, "\n"), "Select album")
	if err != nil {
		return fmt.Errorf("selector failed: %w", err)
	}
	if selected == "" {
		return nil
	}

	idx := indexOf(albumDisplay, selected)
	if idx < 0 {
		return fmt.Errorf("selection not found")
	}

	if albumIDs[idx] == "__PLAY_ALL__" {
		var allSongs []api.Song
		for _, a := range albums {
			songs, err := client.GetAlbum(a.ID)
			if err != nil {
				return fmt.Errorf("cannot fetch album: %w", err)
			}
			allSongs = append(allSongs, songs...)
		}
		if len(allSongs) == 0 {
			fmt.Fprintln(os.Stderr, "No songs found.")
			return nil
		}
		return playSongs(client, allSongs, cfg)
	}

	fmt.Fprintln(os.Stderr, "Fetching songs...")
	songs, err := client.GetAlbum(albumIDs[idx])
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

	if cfg.Scrobble {
		for _, s := range songs {
			if err := client.Scrobble(s.ID); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: scrobble failed for %s: %v\n", s.Title, err)
			}
		}
	}

	return nil
}

func selectSongs(client *api.Client, songs []api.Song, cfg *config.Config) error {
	songIDs := make([]string, 0, len(songs)+1)
	songDisplay := make([]string, 0, len(songs)+1)
	songIDs = append(songIDs, "__PLAY_ALL__")
	songDisplay = append(songDisplay, fmt.Sprintf("Play all (%d songs)", len(songs)))
	for _, s := range songs {
		songIDs = append(songIDs, s.ID)
		display, err := api.RenderSong(s, cfg.SongFormat)
		if err != nil {
			return err
		}
		songDisplay = append(songDisplay, display)
	}

	selected, err := runSelector(cfg.Selector, strings.Join(songDisplay, "\n"), "Select song")
	if err != nil {
		return fmt.Errorf("selector failed: %w", err)
	}
	if selected == "" {
		return nil
	}

	idx := indexOf(songDisplay, selected)
	if idx < 0 {
		return fmt.Errorf("selection not found")
	}

	if songIDs[idx] == "__PLAY_ALL__" {
		return playSongs(client, songs, cfg)
	}

	return playSongs(client, []api.Song{songs[idx-1]}, cfg)
}

func runSelector(name, input, prompt string) (string, error) {
	args := selectorArgs(name, prompt)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", nil
		}
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func indexOf(list []string, item string) int {
	for i, v := range list {
		if v == item {
			return i
		}
	}
	return -1
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
	default: // fzf
		return []string{"fzf", "--prompt", prompt + " > "}
	}
}
