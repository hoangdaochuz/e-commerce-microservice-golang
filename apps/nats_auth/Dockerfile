# Stage 1: build Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY configs ./configs
COPY apps/nats_auth ./apps/nats_auth
RUN go build -o main ./apps/nats_auth/main.go


FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY configs ./configs

# Make binary executable
RUN chmod +x ./main

CMD ["./main"]