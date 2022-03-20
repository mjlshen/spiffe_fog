{ mkShell
, go
, protobuf
, protoc-gen-go
, protoc-gen-go-grpc
}:

mkShell rec {
  packages = [
    go
    protobuf
    protoc-gen-go
    protoc-gen-go-grpc
  ];
}
