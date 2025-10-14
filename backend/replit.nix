{ pkgs }: {
  deps = [
    pkgs.go
    pkgs.postgresql
    pkgs.docker
  ];
}
