FROM golang:1.18 as builder

WORKDIR /go/src/spiffe_fog

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=1 go build -v -o server ./cmd/server/...

FROM ubuntu:latest

COPY --from=builder /go/src/spiffe_fog/server /
CMD ["/server"]
