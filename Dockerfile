# Nimbus Go app - multi-stage build
FROM golang:1.26-alpine AS builder
WORKDIR /app

# Install build deps (for cgo/sqlite if needed)
RUN apk add --no-cache gcc musl-dev

# Copy go mod and source (vendor included if present for replace directives)
COPY go.mod go.sum ./
COPY . .
# Use vendor when present (deploy with replace => ../ fails remotely)
RUN if [ -d vendor ]; then \
    CGO_ENABLED=1 go build -mod=vendor -ldflags="-s -w" -o /app/server .; \
  else \
    go mod download && CGO_ENABLED=1 go build -ldflags="-s -w" -o /app/server .; \
  fi

# Minimal runtime image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/. .
RUN rm -rf vendor go.mod go.sum

# Nimbus uses PORT env (default 8080)
ENV PORT=8080
EXPOSE 8080

CMD ["./server"]
