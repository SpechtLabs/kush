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
        # nixpkgs' default go lags behind; go.mod requires Go 1.27.
        default = pkgs.callPackage ./package.nix {
          buildGoModule = pkgs.buildGo127Module;
          go = pkgs.go_1_27;
          version = self.shortRev or self.dirtyShortRev or "dev";
          commit = self.rev or "none";
        };
      });

      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go_1_27
            gopls
            gotools
            golangci-lint
          ];
        };
      });
    };
}
