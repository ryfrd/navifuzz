package cmd

import (
	"fmt"
	"os"

	"github.com/ryfrd/navifuzz/api"
)

func Playlists(opts Options) error {
	sess, err := newSession(opts)
	if err != nil {
		return err
	}
	client := sess.client
	cfg := sess.cfg

	fmt.Fprintln(os.Stderr, "Fetching playlists...")
	playlists, err := client.GetPlaylists()
	if err != nil {
		return fmt.Errorf("cannot fetch playlists: %w", err)
	}

	if len(playlists) == 0 {
		fmt.Fprintln(os.Stderr, "No playlists found.")
		return nil
	}

	render, err := api.PlaylistRenderer(cfg.PlaylistFormat)
	if err != nil {
		return err
	}
	res, err := pick(playlists, render, "", "Select playlist", cfg.Selector)
	if err != nil {
		return err
	}
	if !res.ok {
		return nil
	}

	fmt.Fprintln(os.Stderr, "Fetching songs...")
	songs, err := client.GetPlaylist(res.item.ID)
	if err != nil {
		return fmt.Errorf("cannot fetch playlist: %w", err)
	}

	if len(songs) == 0 {
		fmt.Fprintln(os.Stderr, "No songs in this playlist.")
		return nil
	}

	return selectSongs(client, songs, cfg)
}
