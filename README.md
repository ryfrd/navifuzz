# navifuzz

## LLM Disclaimer

The code was written with the assistance of [MiMo v2.5](https://mimo.xiaomi.com/mimo-v2-5) using the [opencode](https://opencode.ai/) free tier.

## Mirror

This repository is mirrored from my [forgejo](https://git.dymc.win/james/navifuzz) to [github](https://github.com/ryfrd/navifuzz).

## What is it?

A command line tool for browsing and playing music from a [navidrome](https://navidrome.org/) server.

## What is it not?

A media player.

A fuzzy finder/picker/launcher.

It is not for writing anything to a navidrome server (eg. adding or removing playlists/music).

## Building

Requires [Go](https://go.dev/dl/) 1.26 or newer.

```sh
go build -o navifuzz .
```

Or install directly to your `GOBIN`:

```sh
go install .
```

## Usage

### Dependencies

navifuzz needs the following:

- A [navidrome](https://navidrome.org/) server running somewhere.
- A media player: [mpv](https://mpv.io/) or [vlc](https://www.videolan.org/vlc/) currently supported.
- A fuzzy picker/launcher: [fzf](https://github.com/junegunn/fzf), [dmenu](https://tools.suckless.org/dmenu/), [fuzzel](https://codeberg.org/dnkl/fuzzel), and [rofi](https://github.com/davatorium/rofi) currenlty supported.

### Configuration

Copy `config.example.json` to `~/.config/navifuzz/config.json` (or `$XDG_CONFIG_HOME/navifuzz/config.json` if set) and adjust to your liking.

`server`, `username`, and `password` are required. The rest are optional.
