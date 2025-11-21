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
      };
    };
}
