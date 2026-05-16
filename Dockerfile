FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ai-ticket-server ./cmd/mcp-server/

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /ai-ticket-server .
COPY config.yaml.example ./config.yaml

EXPOSE 8080

ENTRYPOINT ["./ai-ticket-server"]
CMD ["-config", "config.yaml"]
