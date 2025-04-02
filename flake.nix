{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    devshell.url = "github:numtide/devshell";
    git-hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs = inputs:
    inputs.flake-parts.lib.mkFlake {inherit inputs;} {
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-darwin"
        "x86_64-linux"
      ];

      imports = [
        inputs.devshell.flakeModule
        inputs.git-hooks.flakeModule
      ];

      perSystem = {pkgs, ...}: {
        pre-commit.settings.hooks = {
          alejandra.enable = true;
          deadnix.enable = true;
          gofmt.enable = true;
        };

        devshells.default = let
          version = "1.23.4";
          go = pkgs.go.overrideAttrs {
            name = "go-${version}";
            inherit version;
            src = pkgs.fetchFromGitHub {
              owner = "golang";
              repo = "go";
              rev = "refs/tags/go${version}";
              hash = "sha256-rRlln7DluZ0CCmjoHWrXwHxrIIHzd7X6AnEIbSu8Kio=";
            };
          };
        in {
          packages = with pkgs; [
            deadnix
            alejandra
            go
          ];
          env = [
            {
              name = "NIX_CONFIG";
              value = "experimental-features = nix-command flakes";
            }
          ];
          commands = [
            {
              name = "fmt";
              help = "format nix code (using alejandra)";
              command = "nix fmt";
              category = "nix";
            }
            {
              name = "lint";
              help = "lint go code";
              command = "gofmt";
              category = "golang";
            }
          ];
        };
      };
    };
}
