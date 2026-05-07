{
  description = "Mimic of Holding - CLI for a Johnny Decimal Obsidian vault";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "0.4.0";
        src = ./.;
        vendorHash = "sha256-7K17JaXFsjf163g5PXCb5ng2gYdotnZ2IDKk8KFjNj0=";
      in
      {
        packages = {
          mimic = pkgs.buildGoModule {
            pname = "mimic";
            inherit version src vendorHash;
            subPackages = [ "cmd/mimic" ];
            meta.description = "CLI for a Johnny Decimal Obsidian vault";
          };

          default = self.packages.${system}.mimic;
        };

        devShells.default = pkgs.mkShell {
          packages = [ pkgs.go ];
        };
      }
    );
}
