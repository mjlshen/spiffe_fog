all: clean build

gen:
	protoc --proto_path=proto proto/agent/agent.proto --go_out=./proto --go_opt=paths=source_relative --go-grpc_out=./proto --go-grpc_opt=paths=source_relative

build: clean
	CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server/...; \
	CGO_ENABLED=0 go build -ldflags="-s -w" -o client ./cmd/client/...

rpi-build: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o server ./cmd/server/...; \
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o client ./cmd/client/...

rpi-send: rpi-build
	scp -i ~/.ssh/id_ed25519_pi ./server pi@raspberrypi.local:/home/pi/server; \
	scp -i ~/.ssh/id_ed25519_pi ./client pi@raspberrypi.local:/home/pi/client

clean:
	go mod tidy; \
	rm -f client server

.PHONY: all gen build rpi-build rpi-send clean
