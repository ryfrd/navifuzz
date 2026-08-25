package cmd

import (
	"fmt"
	"os"
)

func Songs(size int, configPath string) error {
	sess, err := newSession(configPath)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "Fetching songs...")
	songs, err := sess.client.GetRandomSongs(size)
	if err != nil {
		return fmt.Errorf("cannot fetch songs: %w", err)
	}

	if len(songs) == 0 {
		fmt.Fprintln(os.Stderr, "No songs found.")
		return nil
	}

	return selectSongs(sess.client, songs, sess.cfg)
}
