# --- Build stage ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/notification-service ./cmd

# --- Run stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/bin/notification-service .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh && chown -R app:app /app

USER app

ENTRYPOINT ["./entrypoint.sh"]