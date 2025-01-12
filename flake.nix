{
  description = "NAIS CLI";

  inputs.nixpkgs.url = "nixpkgs/nixos-unstable";

  outputs = {
    self,
    nixpkgs,
  }: let
    version = builtins.substring 0 8 (self.lastModifiedDate or self.lastModified or "19700101");
    # goOverlay = final: prev: {
    #   go = prev.go.overrideAttrs (old: {
    #     version = "1.23.2";
    #     src = prev.fetchurl {
    #       url = "https://go.dev/dl/go1.23.2.src.tar.gz";
    #       hash = "sha256-NpMBYqk99BfZC9IsbhTa/0cFuqwrAkGO3aZxzfqc0H8=";
    #     };
    #   });
    # };
    withSystem = nixpkgs.lib.genAttrs [
      "x86_64-linux"
      "x86_64-darwin"
      "aarch64-linux"
      "aarch64-darwin"
    ];
    withPkgs = callback:
      withSystem (
        system:
          callback (
            import nixpkgs {
              inherit system;
              # overlays = [goOverlay];
            }
          )
      );
  in {
    packages = withPkgs (pkgs: rec {
      nais = pkgs.buildGoModule {
        pname = "nais-cli";
        inherit version;
        src = ./.;
        vendorHash = "sha256-WE5WSOOAUBEfxF/NAaVjTFBZeMRA3APmOJGbBueKObo=";
        postInstall = ''
          mv $out/bin/cli $out/bin/nais
        '';
      };
      default = nais;
    });

    devShells = withPkgs (pkgs: {
      default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          gopls
          gotools
          go-tools
          nodejs_20
          nodePackages.prettier
        ];
      };
    });

    formatter = withPkgs (pkgs: pkgs.nixfmt-rfc-style);
  };
}
