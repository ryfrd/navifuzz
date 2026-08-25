package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/james/navifuzz/api"
	"github.com/james/navifuzz/config"
)

func Playlists() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.Server, cfg.Username, cfg.Password)

	fmt.Fprintln(os.Stderr, "Connecting to Navidrome...")
	if err := client.Ping(); err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Fetching playlists...")
	playlists, err := client.GetPlaylists()
	if err != nil {
		return fmt.Errorf("cannot fetch playlists: %w", err)
	}

	if len(playlists) == 0 {
		fmt.Fprintln(os.Stderr, "No playlists found.")
		return nil
	}

	playlistIDs := make([]string, 0, len(playlists))
	playlistDisplay := make([]string, 0, len(playlists))
	for _, p := range playlists {
		playlistIDs = append(playlistIDs, p.ID)
		display, err := api.RenderPlaylist(p, cfg.PlaylistFormat)
		if err != nil {
			return err
		}
		playlistDisplay = append(playlistDisplay, display)
	}

	selected, err := runSelector(cfg.Selector, strings.Join(playlistDisplay, "\n"), "Select playlist")
	if err != nil {
		if err.Error() == "exit status 1" {
			return nil
		}
		return fmt.Errorf("selector failed: %w", err)
	}

	idx := indexOf(playlistDisplay, selected)
	if idx < 0 {
		return fmt.Errorf("selection not found")
	}

	fmt.Fprintln(os.Stderr, "Fetching songs...")
	songs, err := client.GetPlaylist(playlistIDs[idx])
	if err != nil {
		return fmt.Errorf("cannot fetch playlist: %w", err)
	}

	if len(songs) == 0 {
		fmt.Fprintln(os.Stderr, "No songs in this playlist.")
		return nil
	}

	return selectSongs(client, songs, cfg)
}
