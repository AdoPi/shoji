{
  description = "Manage SSH keys with Nix";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/release-23.11";
  inputs.flake-utils.url = "github:numtide/flake-utils";
  outputs = {
    self,
    nixpkgs,
    flake-utils
  }: let
    systems = [
      "x86_64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
      "aarch64-linux"
    ];
  in flake-utils.lib.eachSystem systems (system:
    let
      pkgs = import nixpkgs { inherit system; };
    in {
      packages.default = pkgs.callPackage ./default.nix { };
      devShells.default = pkgs.mkShell {
        buildInputs = self.packages.${system}.default.nativeBuildInputs;
      };
    });
}
