package cmd

import (
	"fmt"
	"os"

	"github.com/james/navifuzz/api"
	"github.com/james/navifuzz/config"
)

func Songs(size int) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	client := api.NewClient(cfg.Server, cfg.Username, cfg.Password)

	fmt.Fprintln(os.Stderr, "Connecting to Navidrome...")
	if err := client.Ping(); err != nil {
		return fmt.Errorf("cannot connect to server: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Fetching songs...")
	songs, err := client.GetRandomSongs(size)
	if err != nil {
		return fmt.Errorf("cannot fetch songs: %w", err)
	}

	if len(songs) == 0 {
		fmt.Fprintln(os.Stderr, "No songs found.")
		return nil
	}

	return selectSongs(client, songs, cfg)
}
