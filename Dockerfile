# ---- Build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod ./
RUN go mod tidy

COPY . .
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# ---- Runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /app/server /app/server

ENV ADDR=:8080
ENV DB_PATH=/data/data.db

RUN mkdir -p /data && chown -R app:app /data
USER app

EXPOSE 8080

ENTRYPOINT ["/app/server"]
