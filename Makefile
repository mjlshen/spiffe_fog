.PHONY: all gen build send clean

all: clean build send

gen:
	protoc --proto_path=proto proto/agent/agent.proto --go_out=./proto --go_opt=paths=source_relative --go-grpc_out=./proto --go-grpc_opt=paths=source_relative

build:
	go mod tidy; cd server && GOOS=linux GOARCH=arm GOARM=6 go build .; cd ../client && CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build .

send:
	scp -i ~/.ssh/id_ed25519_pi ./server/server pi@raspberrypi.local:/home/pi/server; scp -i ~/.ssh/id_ed25519_pi ./client/client pi@raspberrypi.local:/home/pi/client

clean:
	rm -f client/client server/server
