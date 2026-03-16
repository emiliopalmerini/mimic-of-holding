{
  description = "Mimic of Holding — CLI and MCP server for a Johnny Decimal Obsidian vault";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-fM19kutxYJBJXs+nfBX49p+HL4cY/Mvxw6heE9Hzjxc=";
      in
      {
        packages = {
          mimic = pkgs.buildGoModule {
            pname = "mimic";
            inherit version src vendorHash;
            subPackages = [ "cmd/mimic" ];
            meta.description = "CLI for a Johnny Decimal Obsidian vault";
          };

          mimic-mcp = pkgs.buildGoModule {
            pname = "mimic-mcp";
            inherit version src vendorHash;
            subPackages = [ "cmd/mimic-mcp" ];
            meta.description = "MCP server for a Johnny Decimal Obsidian vault";
          };

          default = self.packages.${system}.mimic;
        };

        devShells.default = pkgs.mkShell {
          packages = [ pkgs.go ];
        };
      }
    );
}
