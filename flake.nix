{
  description = "Manage SSH keys with Nix";
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/release-23.11";
  outputs = {
    self,
    nixpkgs
  }: let
    systems = [
      "x86_64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
      "aarch64-linux"
    ];

    forAllSystems = f: builtins.listToAttrs (map (system: { name = system; value = f system; }) systems);
  in {
    devShells = forAllSystems (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in {
        default = pkgs.mkShell {
          buildInputs = with pkgs; [ go ];
        };
      });
  };
}
