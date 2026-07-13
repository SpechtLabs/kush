{ lib
, buildGoModule
}:

buildGoModule rec {
  pname = "kush";
  version = "1.0.0";

  src = ./.;

  # We use a fake hash first; Nix will fail the build and print the correct hash.
  vendorHash = "sha256-j9888cQvkpiH/uBRjMGOa3s9qJ6Sa0MkX7NNucHiMtU=";

  subPackages = [ "cmd/kush" ];

  # Lower the required Go version in go.mod to match nixpkgs' Go version if needed.
  postPatch = ''
    substituteInPlace go.mod --replace-warn "go 1.26.5" "go 1.26.4"
  '';

  ldflags = [
    "-s"
    "-w"
    "-X main.Version=${version}"
    "-X main.Commit=none"
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
