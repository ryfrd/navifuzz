self:
{
  lib,
  pkgs,
  config,
  ...
}:

let
  cfg = config.programs.navifuzz;
  format = pkgs.formats.json { };
in
{
  options.programs.navifuzz = {
    enable = lib.mkEnableOption "navifuzz";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      defaultText = lib.literalExpression "navifuzz";
      description = "The navifuzz package to use.";
    };

    settings = lib.mkOption {
      type = format.type;
      default = { };
      example = lib.literalExpression ''
        {
          server = "https://navidrome.example.com";
          username = "user";
          password = "pass";
          selector = "fzf";
          player = "mpv";
          shuffle = false;
          loop = false;
        }
      '';
      description = "Configuration written to config.json.";
    };
  };

  config = lib.mkIf cfg.enable {
    home.packages = [ cfg.package ];

    xdg.configFile."navifuzz/config.json".source = format.generate "navifuzz-config.json" cfg.settings;
  };
}
