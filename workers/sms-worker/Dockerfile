FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /sms-worker ./cmd/worker

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /sms-worker .
COPY --from=builder /app/migrations ./migrations
CMD ["./sms-worker"]
