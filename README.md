## LLM Disclaimer

The code was written with the assistance of [MiMo v2.5](https://mimo.xiaomi.com/mimo-v2-5) using the [opencode](https://opencode.ai/) free tier.

## Mirror

This repository is mirrored from my [forgejo](https://git.dymc.win/james/navifuzz) to [github](https://github.com/ryfrd/navifuzz).

## What is it?

navifuzz is a command line tool for browsing and playing the music on your [navidrome](https://navidrome.org/) server.

It fetches lists of albums, artists, songs, playlists, and genres from the server, presents them in a fuzzy picker, and hands your selection off to a media player.

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

## Nix

The repo ships a flake. Install with one command:

```sh
nix run github:ryfrd/navifuzz -- albums
```

Or add to your profile:

```sh
nix profile install github:ryfrd/navifuzz
```

A [home-manager](https://github.com/nix-community/home-manager) module provides
`programs.navifuzz` (options: `enable`, `package`, `settings`), which installs
the package and writes `~/.config/navifuzz/config.json`:

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";
    navifuzz.url = "github:ryfrd/navifuzz";
  };
  outputs = { navifuzz, ... }: {
    homeManagerModules.default = navifuzz.homeManagerModules.default;
  };
}
```

## Usage

### Dependencies

navifuzz needs the following:

- A [navidrome](https://navidrome.org/) server running somewhere.
- A media player: [mpv](https://mpv.io/) or [vlc](https://www.videolan.org/vlc/) currently supported.
- A fuzzy picker/launcher: [fzf](https://github.com/junegunn/fzf), [skim](https://github.com/lotabout/skim), [fuzzel](https://codeberg.org/dnkl/fuzzel), [rofi](https://github.com/davatorium/rofi), [tofi](https://github.com/philj56/tofi), [wofi](https://hg.sr.ht/~scoopta/wofi), [dmenu](https://tools.suckless.org/dmenu/), [bemenu](https://github.com/Cloudef/bemenu), and [noctalia](https://noctalia.dev/) currently supported.

### Configuration

Copy `config.example.json` to `~/.config/navifuzz/config.json` (or `$XDG_CONFIG_HOME/navifuzz/config.json` if set) and adjust to your liking.

`server`, `username`, and `password` are required. The rest are optional.

The picker and player can also be overridden per-invocation without touching the config file:

```sh
navifuzz --selector rofi albums
navifuzz --player vlc songs
navifuzz --config /path/to/other/config.json albums
```

The `--config`, `--selector`, and `--player` flags can appear anywhere before or after the command.
