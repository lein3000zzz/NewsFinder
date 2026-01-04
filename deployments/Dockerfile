FROM golang:1.25.1 AS builder
WORKDIR /app
ENV CGO_ENABLED=1
RUN apt-get update && apt-get install -y build-essential git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN cd tools/sqlc && sqlc generate
RUN go build -o newsfinder ./cmd/newsfinder

FROM ubuntu:latest
RUN apt-get update && apt-get install -y ca-certificates libstdc++6
WORKDIR /root/
COPY --from=builder /app/newsfinder .
COPY --from=builder /app/main.env ./main.env
COPY --from=builder /app/tools ./tools
CMD ["./newsfinder"]