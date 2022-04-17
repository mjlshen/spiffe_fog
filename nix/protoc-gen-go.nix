{ buildGoModule, fetchFromGitHub }:

buildGoModule rec {
  pname = "protoc-gen-go";
  version = "1.28.0";

  src = fetchFromGitHub {
    owner = "protocolbuffers";
    repo = "protobuf-go";
    rev = "v${version}";
    sha256 = "sha256-GXXhBXzeFgS4DjY3jDRaJu/9PbfUnmWJiE4BwDBh7Ns=";
  };

  vendorSha256 = "sha256-yb8l4ooZwqfvenlxDRg95rqiL+hmsn0weS/dPv/oD2Y=";

  subPackages = [ "cmd/protoc-gen-go" ];
}
