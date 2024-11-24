{ description = "Go Anime Nix Flake";

  inputs = {
    nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.2405.*.tar.gz";
    nixos-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";
  };
  outputs = { self, nixpkgs, nixos-unstable }:
    let
      allSystems = [
        "x86_64-linux" 
        "aarch64-linux" 
        "x86_64-darwin" 
        "aarch64-darwin" 
      ];

      
      forAllSystems = f: nixpkgs.lib.genAttrs allSystems (system: f {
        
        pkgs = import nixos-unstable { inherit system; };
      });
    in
    {
      packages = forAllSystems ({ pkgs }: {
        default = pkgs.buildGoModule {
          name = "GoAnime";
          go = pkgs.go_1_23; 
          src = self;
          vendorHash = "sha256-dqfgiBMcEhq5hr524BKIbP0ulByWJa7gkoxSy4598v8=";
          subPackages = [ "cmd/goanime" ];
          propagatedBuildInputs = with pkgs;[ mpv yt-dlp ]; 
        };
      });
      devShell = forAllSystems ({ pkgs }: pkgs.mkShell {
        buildInputs = with pkgs;[ mpv yt-dlp];
      });
    };
}

