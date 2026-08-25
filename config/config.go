package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Server         string `json:"server"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Scrobble       bool   `json:"scrobble"`
	Selector       string `json:"selector"`
	Player         string `json:"player"`
	Shuffle        bool   `json:"shuffle"`
	Loop           bool   `json:"loop"`
	AlbumFormat    string `json:"album_format"`
	SongFormat     string `json:"song_format"`
	ArtistFormat   string `json:"artist_format"`
	PlaylistFormat string `json:"playlist_format"`
	GenreFormat    string `json:"genre_format"`
}

const (
	defaultAlbumFormat    = `{{.Artist}} - {{.Name}} ({{.Year}}, {{.SongCount}} songs)`
	defaultSongFormat     = `{{.TrackStr}}. {{.Title}} ({{.DurationStr}})`
	defaultArtistFormat   = `{{.Name}} ({{.AlbumCount}} albums)`
	defaultPlaylistFormat = `{{.Name}} ({{.SongCount}} songs)`
	defaultGenreFormat    = `{{.Name}} ({{.SongCount}} songs, {{.AlbumCount}} albums)`
)

func Load() (*Config, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot find home directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}

	path := filepath.Join(configDir, "navifuzz", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	if cfg.Server == "" {
		return nil, fmt.Errorf("config: server is required")
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("config: username is required")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("config: password is required")
	}

	if cfg.Selector == "" {
		cfg.Selector = "fzf"
	}
	if cfg.Player == "" {
		cfg.Player = "mpv"
	}
	if cfg.AlbumFormat == "" {
		cfg.AlbumFormat = defaultAlbumFormat
	}
	if cfg.SongFormat == "" {
		cfg.SongFormat = defaultSongFormat
	}
	if cfg.ArtistFormat == "" {
		cfg.ArtistFormat = defaultArtistFormat
	}
	if cfg.PlaylistFormat == "" {
		cfg.PlaylistFormat = defaultPlaylistFormat
	}
	if cfg.GenreFormat == "" {
		cfg.GenreFormat = defaultGenreFormat
	}

	return &cfg, nil
}
