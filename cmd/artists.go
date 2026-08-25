package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/james/navifuzz/api"
	"github.com/james/navifuzz/config"
)

func Artists() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.Server, cfg.Username, cfg.Password)

	fmt.Fprintln(os.Stderr, "Connecting to Navidrome...")
	if err := client.Ping(); err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Fetching artists...")
	artists, err := client.GetArtists()
	if err != nil {
		return fmt.Errorf("cannot fetch artists: %w", err)
	}

	if len(artists) == 0 {
		fmt.Fprintln(os.Stderr, "No artists found.")
		return nil
	}

	artistIDs := make([]string, 0, len(artists))
	artistDisplay := make([]string, 0, len(artists)+1)
	artistIDs = append(artistIDs, "__PLAY_ALL__")
	artistDisplay = append(artistDisplay, fmt.Sprintf("Play all (%d artists)", len(artists)))
	for _, a := range artists {
		artistIDs = append(artistIDs, a.ID)
		display, err := api.RenderArtist(a, cfg.ArtistFormat)
		if err != nil {
			return err
		}
		artistDisplay = append(artistDisplay, display)
	}

	selected, err := runSelector(cfg.Selector, strings.Join(artistDisplay, "\n"), "Select artist")
	if err != nil {
		if err.Error() == "exit status 1" {
			return nil
		}
		return fmt.Errorf("selector failed: %w", err)
	}

	idx := indexOf(artistDisplay, selected)
	if idx < 0 {
		return fmt.Errorf("selection not found")
	}

	if artistIDs[idx] == "__PLAY_ALL__" {
		var allSongs []api.Song
		for _, a := range artists {
			albums, err := client.GetArtist(a.ID)
			if err != nil {
				return fmt.Errorf("cannot fetch artist: %w", err)
			}
			for _, al := range albums {
				songs, err := client.GetAlbum(al.ID)
				if err != nil {
					return fmt.Errorf("cannot fetch album: %w", err)
				}
				allSongs = append(allSongs, songs...)
			}
		}
		if len(allSongs) == 0 {
			fmt.Fprintln(os.Stderr, "No songs found.")
			return nil
		}
		return playSongs(client, allSongs, cfg)
	}

	fmt.Fprintln(os.Stderr, "Fetching albums...")
	albums, err := client.GetArtist(artistIDs[idx])
	if err != nil {
		return fmt.Errorf("cannot fetch artist: %w", err)
	}

	if len(albums) == 0 {
		fmt.Fprintln(os.Stderr, "No albums by this artist.")
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

	selectedAlbum, err := runSelector(cfg.Selector, strings.Join(albumDisplay, "\n"), "Select album")
	if err != nil {
		if err.Error() == "exit status 1" {
			return nil
		}
		return fmt.Errorf("selector failed: %w", err)
	}

	albumIdx := indexOf(albumDisplay, selectedAlbum)
	if albumIdx < 0 {
		return fmt.Errorf("selection not found")
	}

	if albumIDs[albumIdx] == "__PLAY_ALL__" {
		var allSongs []api.Song
		for _, a := range albums {
			songs, err := client.GetAlbum(a.ID)
			if err != nil {
				return fmt.Errorf("cannot fetch album: %w", err)
			}
			allSongs = append(allSongs, songs...)
		}
		if len(allSongs) == 0 {
			fmt.Fprintln(os.Stderr, "No songs by this artist.")
			return nil
		}
		return playSongs(client, allSongs, cfg)
	}

	fmt.Fprintln(os.Stderr, "Fetching songs...")
	songs, err := client.GetAlbum(albumIDs[albumIdx])
	if err != nil {
		return fmt.Errorf("cannot fetch album: %w", err)
	}

	if len(songs) == 0 {
		fmt.Fprintln(os.Stderr, "No songs in this album.")
		return nil
	}

	return selectSongs(client, songs, cfg)
}
