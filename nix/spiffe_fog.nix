{ mkShell
, go
, grpcurl
, protobuf
, protoc-gen-go
, protoc-gen-go-grpc
}:

mkShell rec {
  packages = [
    go
    grpcurl
    protobuf
    protoc-gen-go
    protoc-gen-go-grpc
  ];
}
