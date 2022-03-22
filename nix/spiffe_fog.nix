{ mkShell
, grpcurl
, protobuf
, protoc-gen-go
, protoc-gen-go-grpc
}:

mkShell rec {
  packages = [
    grpcurl
    protobuf
    protoc-gen-go
    protoc-gen-go-grpc
  ];
}
