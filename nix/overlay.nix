final: prev: {
  go = final.go_1_17;
  protobuf = final.protobuf3_19;

  protoc-gen-go = prev.callPackage ./protoc-gen-go.nix { };
  protoc-gen-go-grpc = prev.callPackage ./protoc-gen-go-grpc.nix { };

  devShell = final.callPackage ./spiffe_fog.nix { };
}
