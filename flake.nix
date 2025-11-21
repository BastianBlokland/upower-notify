{
  description = "upower-notify";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-25.05";
  };

  outputs =
    { nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      tmpDir = "/tmp/upower-notify";
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        packages = [
          pkgs.nixfmt-rfc-style

          pkgs.go
          pkgs.gopls
          pkgs.go-tools
          pkgs.delve
        ];

        GOPATH = "${tmpDir}/go";
        GOENV = "${tmpDir}/go/env";
        GOCACHE ="${tmpDir}/go/cache";
        GOMODCACHE = "${tmpDir}/go/pkg/mod";
        GOTELEMETRYDIR = "${tmpDir}/go/telemetry";
      };
    };
}
