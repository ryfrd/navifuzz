package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ryfrd/navifuzz/api"
	"github.com/ryfrd/navifuzz/config"
)

type Options struct {
	ConfigPath string
	Selector   string
	Player     string
}

type session struct {
	cfg    *config.Config
	client *api.Client
}

func newSession(opts Options) (*session, error) {
	cfg, err := config.LoadPath(opts.ConfigPath)
	if err != nil {
		return nil, err
	}
	if opts.Selector != "" {
		cfg.Selector = opts.Selector
	}
	if opts.Player != "" {
		cfg.Player = opts.Player
	}
	client := api.NewClient(cfg.Server, cfg.Username, cfg.Password)
	fmt.Fprintln(os.Stderr, "Connecting to Navidrome...")
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("cannot connect to server: %w", err)
	}
	return &session{cfg: cfg, client: client}, nil
}

type pickResult[T any] struct {
	item    T
	playAll bool
	ok      bool
}

func pick[T any](items []T, render func(T) (string, error), playAllHeader string, prompt string, selector string) (pickResult[T], error) {
	var display []string
	hasPlayAll := playAllHeader != ""
	if hasPlayAll {
		display = append(display, playAllHeader)
	}
	for _, it := range items {
		d, err := render(it)
		if err != nil {
			return pickResult[T]{}, err
		}
		display = append(display, d)
	}
	makeUniqueDisplay(display)

	selected, err := runSelector(selector, strings.Join(display, "\n"), prompt)
	if err != nil {
		return pickResult[T]{}, fmt.Errorf("selector failed: %w", err)
	}
	if selected == "" {
		return pickResult[T]{}, nil
	}

	idx := indexOf(display, selected)
	if idx < 0 {
		return pickResult[T]{}, fmt.Errorf("selection not found")
	}
	if hasPlayAll && idx == 0 {
		return pickResult[T]{ok: true, playAll: true}, nil
	}
	itemIdx := idx
	if hasPlayAll {
		itemIdx = idx - 1
	}
	return pickResult[T]{ok: true, item: items[itemIdx]}, nil
}
