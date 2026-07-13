{ lib
, buildGoModule
, go
, version ? "dev"
, commit ? "none"
}:

buildGoModule {
  pname = "kush";
  inherit version;

  src = ./.;

  # Stale after a go.sum change (renovate gomod bumps) → the `nix` CI job
  # goes red. Fix with: mise run nix:update-hash
  vendorHash = "sha256-j9888cQvkpiH/uBRjMGOa3s9qJ6Sa0MkX7NNucHiMtU=";

  subPackages = [ "cmd/kush" ];

  # Dynamically adjust the required Go version in go.mod to match the Nix compiler version.
  # This prevents compiler version mismatch failures when upstreaming or upgrading.
  postPatch = ''
    sed -i 's/^go [0-9.]*/go ${go.version}/' go.mod
  '';

  ldflags = [
    "-s"
    "-w"
    "-X main.Version=${version}"
    "-X main.Commit=${commit}"
    "-X main.Date=unknown"
  ];

  meta = with lib; {
    description = "Ephemeral, isolated kube-context subshells";
    homepage = "https://github.com/spechtlabs/kush";
    license = licenses.asl20;
    mainProgram = "kush";
    platforms = platforms.unix;
  };
}
