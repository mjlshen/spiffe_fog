.PHONY: all gen

all: gen

gen:
	mkdir -p workload
	protoc --proto_path=proto proto/spiffe_fog.proto --go_out=./workload --go_opt=paths=source_relative --go-grpc_out=./workload --go-grpc_opt=paths=source_relative
