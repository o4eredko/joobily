FROM golang:1.15-alpine AS builder
WORKDIR /app/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app cmd/main.go

FROM alpine:latest
WORKDIR /app/
COPY --from=builder /app/app .
COPY --from=builder /app/config.toml .
CMD ["./app"]
