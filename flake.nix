{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      version = "0.1.0";
      defaultPackage = pkgs.buildGoModule {
        pname = "navifuzz";
        version = version;
        src = self;
        vendorHash = null;
        ldflags = [
          "-s"
          "-w"
          "-X main.version=${version}"
        ];
      };
    in
    {
      packages.${system}.default = defaultPackage;

      apps.${system}.default = {
        type = "app";
        program = "${defaultPackage}/bin/navifuzz";
      };

      devShells.${system}.default = pkgs.mkShell {
        packages = [
          pkgs.go
          pkgs.gotools
          pkgs.gopls
        ];
      };

      homeManagerModules.default = import ./nix/home-module.nix self;
    };
}
