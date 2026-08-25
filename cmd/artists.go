package cmd

import (
	"fmt"
	"os"

	"github.com/ryfrd/navifuzz/api"
)

func Artists(configPath string) error {
	sess, err := newSession(configPath)
	if err != nil {
		return err
	}
	client := sess.client
	cfg := sess.cfg

	fmt.Fprintln(os.Stderr, "Fetching artists...")
	artists, err := client.GetArtists()
	if err != nil {
		return fmt.Errorf("cannot fetch artists: %w", err)
	}

	if len(artists) == 0 {
		fmt.Fprintln(os.Stderr, "No artists found.")
		return nil
	}

	artistRender, err := api.ArtistRenderer(cfg.ArtistFormat)
	if err != nil {
		return err
	}
	res, err := pick(artists, artistRender, fmt.Sprintf("Play all (%d artists)", len(artists)), "Select artist", cfg.Selector)
	if err != nil {
		return err
	}
	if !res.ok {
		return nil
	}

	if res.playAll {
		allSongs, err := fetchSongsForArtists(client, artists)
		if err != nil {
			return err
		}
		if len(allSongs) == 0 {
			fmt.Fprintln(os.Stderr, "No songs found.")
			return nil
		}
		return playSongs(client, allSongs, cfg)
	}

	fmt.Fprintln(os.Stderr, "Fetching albums...")
	albums, err := client.GetArtist(res.item.ID)
	if err != nil {
		return fmt.Errorf("cannot fetch artist: %w", err)
	}

	if len(albums) == 0 {
		fmt.Fprintln(os.Stderr, "No albums by this artist.")
		return nil
	}

	albumRender, err := api.AlbumRenderer(cfg.AlbumFormat)
	if err != nil {
		return err
	}
	res2, err := pick(albums, albumRender, fmt.Sprintf("Play all (%d albums)", len(albums)), "Select album", cfg.Selector)
	if err != nil {
		return err
	}
	if !res2.ok {
		return nil
	}

	if res2.playAll {
		allSongs, err := fetchSongsForAlbums(client, albums)
		if err != nil {
			return err
		}
		if len(allSongs) == 0 {
			fmt.Fprintln(os.Stderr, "No songs by this artist.")
			return nil
		}
		return playSongs(client, allSongs, cfg)
	}

	fmt.Fprintln(os.Stderr, "Fetching songs...")
	songs, err := client.GetAlbum(res2.item.ID)
	if err != nil {
		return fmt.Errorf("cannot fetch album: %w", err)
	}

	if len(songs) == 0 {
		fmt.Fprintln(os.Stderr, "No songs in this album.")
		return nil
	}

	return selectSongs(client, songs, cfg)
}
