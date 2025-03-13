FROM golang:1.24 as builder

WORKDIR /go/src/spiffe_fog

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go build -v -ldflags="-s -w" -o server ./cmd/server/...

FROM gcr.io/distroless/static

COPY --from=builder /go/src/spiffe_fog/server /
CMD ["/server"]
