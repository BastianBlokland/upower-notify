{
  description = "upower-notify";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-26.05";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
      tmpDir = "/tmp/upower-notify";
    in
    {
      packages.${system}.default = pkgs.buildGoModule {
        pname = "upower-notify";
        version = "0.1.0";
        src = ./.;
        vendorHash = "sha256-oPiRr9x0wEbEjwrB8KsO1vHCSV+f/dn4VTUI3NI4UpE=";
        proxyVendor = true;

        meta = {
          description = "Fork of: https://github.com/omeid/upower-notify for personal use.";
          mainProgram = "upower-notify";
        };
      };

      defaultPackage.${system} = self.packages.${system};

      devShells.${system}.default = pkgs.mkShell {
        packages = [
          pkgs.nixfmt

          pkgs.go
          pkgs.gopls
          pkgs.go-tools
          pkgs.delve
        ];

        # Avoid polluting our home directory.
        GOPATH = "${tmpDir}/go";
        GOENV = "${tmpDir}/go/env";
        GOCACHE ="${tmpDir}/go/cache";
        GOMODCACHE = "${tmpDir}/go/pkg/mod";
        GOTELEMETRYDIR = "${tmpDir}/go/telemetry";
      };
    };
}
