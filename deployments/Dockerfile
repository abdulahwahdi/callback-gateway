# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/webhook-middleware ./cmd/webhook-middleware

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/webhook-middleware ./webhook-middleware
COPY --from=builder /src/migrations ./migrations
EXPOSE 8090
ENTRYPOINT ["./webhook-middleware"]
