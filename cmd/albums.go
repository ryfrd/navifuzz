package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/james/navifzf/api"
	"github.com/james/navifzf/config"
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

	albumIDs := make([]string, len(albums))
	albumDisplay := make([]string, len(albums))
	for i, a := range albums {
		albumIDs[i] = a.ID
		albumDisplay[i] = api.DisplayAlbum(a)
	}

	selected, err := runSelector(cfg.Selector, strings.Join(albumDisplay, "\n"), "Select album")
	if err != nil {
		if err.Error() == "exit status 1" {
			return nil
		}
		return fmt.Errorf("selector failed: %w", err)
	}

	idx := indexOf(albumDisplay, selected)
	if idx < 0 {
		return fmt.Errorf("selection not found")
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

	songIDs := make([]string, 0, len(songs)+1)
	songDisplay := make([]string, 0, len(songs)+1)
	songIDs = append(songIDs, "__PLAY_ALL__")
	songDisplay = append(songDisplay, fmt.Sprintf("Play all (%d songs)", len(songs)))
	for _, s := range songs {
		songIDs = append(songIDs, s.ID)
		songDisplay = append(songDisplay, api.DisplaySong(s))
	}

	selectedSong, err := runSelector(cfg.Selector, strings.Join(songDisplay, "\n"), "Select song")
	if err != nil {
		if err.Error() == "exit status 1" {
			return nil
		}
		return fmt.Errorf("selector failed: %w", err)
	}

	songIdx := indexOf(songDisplay, selectedSong)
	if songIdx < 0 {
		return fmt.Errorf("selection not found")
	}

	if songIDs[songIdx] == "__PLAY_ALL__" {
		return playSongs(client, songs, cfg)
	}

	return playSongs(client, []api.Song{songs[songIdx-1]}, cfg)
}

func playSongs(client *api.Client, songs []api.Song, cfg *config.Config) error {
	var urls []string
	for _, s := range songs {
		urls = append(urls, client.StreamURL(s.ID))
	}

	tmpDir := os.TempDir()
	playlistFile := filepath.Join(tmpDir, "navifzf-playlist.m3u")
	if err := os.WriteFile(playlistFile, []byte(strings.Join(urls, "\n")), 0644); err != nil {
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

func runSelector(name, input, prompt string) (string, error) {
	args := selectorArgs(name, prompt)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
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
		return []string{"fuzzel", "-p", prompt + " > "}
	case "rofi":
		return []string{"rofi", "-dmenu", "-p", prompt}
	default: // fzf
		return []string{"fzf", "--prompt", prompt + " > "}
	}
}
