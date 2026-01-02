FROM golang:1.25.1 AS builder
WORKDIR /app
ENV CGO_ENABLED=1
RUN apt-get update && apt-get install -y build-essential git && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o newsfinder ./cmd/newsfinder

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates libstdc++6 && rm -rf /var/lib/apt/lists/*
WORKDIR /root/
COPY --from=builder /app/newsfinder .
COPY --from=builder /app/main.env ./main.env
COPY --from=builder /app/tools ./tools
CMD ["./newsfinder"]