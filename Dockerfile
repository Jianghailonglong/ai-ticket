FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloud-mcp-server ./cmd/mcp-server/

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /cloud-mcp-server .
COPY config.yaml.example ./config.yaml

EXPOSE 8080

ENTRYPOINT ["./cloud-mcp-server"]
CMD ["-config", "config.yaml"]
