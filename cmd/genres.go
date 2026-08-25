package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/james/navifuzz/api"
	"github.com/james/navifuzz/config"
)

func Genres(n int) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.Server, cfg.Username, cfg.Password)

	fmt.Fprintln(os.Stderr, "Connecting to Navidrome...")
	if err := client.Ping(); err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Fetching genres...")
	genres, err := client.GetGenres()
	if err != nil {
		return fmt.Errorf("cannot fetch genres: %w", err)
	}

	if len(genres) == 0 {
		fmt.Fprintln(os.Stderr, "No genres found.")
		return nil
	}

	genreNames := make([]string, 0, len(genres))
	for _, g := range genres {
		display, err := api.RenderGenre(g, cfg.GenreFormat)
		if err != nil {
			return err
		}
		genreNames = append(genreNames, display)
	}

	selected, err := runSelector(cfg.Selector, strings.Join(genreNames, "\n"), "Select genre")
	if err != nil {
		if err.Error() == "exit status 1" {
			return nil
		}
		return fmt.Errorf("selector failed: %w", err)
	}

	idx := indexOf(genreNames, selected)
	if idx < 0 {
		return fmt.Errorf("selection not found")
	}

	fmt.Fprintf(os.Stderr, "Fetching songs in %s...\n", genres[idx].Name)
	songs, err := client.GetSongsByGenre(genres[idx].Name, n)
	if err != nil {
		return fmt.Errorf("cannot fetch songs by genre: %w", err)
	}

	if len(songs) == 0 {
		fmt.Fprintln(os.Stderr, "No songs found in this genre.")
		return nil
	}

	return selectSongs(client, songs, cfg)
}
