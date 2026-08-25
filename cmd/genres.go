package cmd

import (
	"fmt"
	"os"

	"github.com/ryfrd/navifuzz/api"
)

func Genres(n int, configPath string) error {
	sess, err := newSession(configPath)
	if err != nil {
		return err
	}
	client := sess.client
	cfg := sess.cfg

	fmt.Fprintln(os.Stderr, "Fetching genres...")
	genres, err := client.GetGenres()
	if err != nil {
		return fmt.Errorf("cannot fetch genres: %w", err)
	}

	if len(genres) == 0 {
		fmt.Fprintln(os.Stderr, "No genres found.")
		return nil
	}

	render, err := api.GenreRenderer(cfg.GenreFormat)
	if err != nil {
		return err
	}
	res, err := pick(genres, render, "", "Select genre", cfg.Selector)
	if err != nil {
		return err
	}
	if !res.ok {
		return nil
	}

	fmt.Fprintf(os.Stderr, "Fetching songs in %s...\n", res.item.Name)
	songs, err := client.GetSongsByGenre(res.item.Name, n)
	if err != nil {
		return fmt.Errorf("cannot fetch songs by genre: %w", err)
	}

	if len(songs) == 0 {
		fmt.Fprintln(os.Stderr, "No songs found in this genre.")
		return nil
	}

	return selectSongs(client, songs, cfg)
}
