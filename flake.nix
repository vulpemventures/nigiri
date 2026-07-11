{
  description = "Nigiri - one-click bitcoin development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        version = self.shortRev or self.dirtyShortRev or "dev";
      in
      {
        packages = {
          nigiri = pkgs.buildGoModule {
            pname = "nigiri";
            inherit version;

            src = self;

            vendorHash = "sha256-Ob4E/Q/6QmYzqT8hD4GICvAe2X6V/BkLXzfJMmPDR+4=";

            subPackages = [ "cmd/nigiri" ];

            env.CGO_ENABLED = "0";

            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
              "-X main.commit=${self.rev or "dirty"}"
            ];

            # Needs docker CLI (for `docker inspect` and `docker compose` fallback)
            # and docker-compose at runtime
            nativeBuildInputs = [ pkgs.makeWrapper ];
            postInstall = ''
              wrapProgram $out/bin/nigiri \
                --prefix PATH : ${pkgs.lib.makeBinPath [ pkgs.docker-client pkgs.docker-compose ]}
            '';

            meta = {
              description = "One-click bitcoin development environment";
              homepage = "https://github.com/vulpemventures/nigiri";
              license = pkgs.lib.licenses.mit;
              mainProgram = "nigiri";
            };
          };

          default = self.packages.${system}.nigiri;
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            docker-compose
            goreleaser
          ];
        };
      }
    );
}
