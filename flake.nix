{
  description = "Navidrome CLI with fzf + mpv";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
  };

  outputs =
    { self, nixpkgs }:
    {
      packages.x86_64-linux.default = nixpkgs.legacyPackages.x86_64-linux.buildGoModule {
        pname = "navifuzz";
        version = "unstable";
        src = ./.;
        vendorHash = null;
      };
    };
}
