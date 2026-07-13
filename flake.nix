{
  description = "Ephemeral, isolated kube-context subshells";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs { inherit system; };
      });
    in
    {
      packages = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.callPackage ./package.nix {
          version = self.shortRev or self.dirtyShortRev or "dev";
          commit = self.rev or "none";
        };
      });

      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            golangci-lint
          ];
        };
      });
    };
}
