FROM golang:1.22.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gql_server cmd/gql_server/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/gql_server .

CMD ["./gql_server"]
