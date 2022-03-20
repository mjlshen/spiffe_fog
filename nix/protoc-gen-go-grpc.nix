{ buildGoModule, fetchFromGitHub }:

buildGoModule rec {
  pname = "protoc-gen-go-grpc";
  version = "1.2.0";
  modRoot = "cmd/protoc-gen-go-grpc";

  src = fetchFromGitHub {
    owner = "grpc";
    repo = "grpc-go";
    rev = "cmd/protoc-gen-go-grpc/v${version}";
    sha256 = "sha256-pIHMykwwyu052rqwSV5Js0+JCKCNLrFVN6XQ3xjlEOI=";
  };

  vendorSha256 = "sha256-yxOfgTA5IIczehpWMM1kreMqJYKgRT5HEGbJ3SeQ/Lg=";
}
