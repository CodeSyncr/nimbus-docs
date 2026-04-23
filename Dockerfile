# ============================================
# Nimbus Forge — Multi-stage Dockerfile
# ============================================

# Stage 1: Build
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app

# Cache dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /app/server .

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /app

# Copy binary and assets.
COPY --from=builder /app/server /app/server
COPY --from=builder /app/public ./public/
COPY --from=builder /app/resources ./resources/

# Create non-root user.
RUN addgroup -S nimbus && adduser -S nimbus -G nimbus
RUN chown -R nimbus:nimbus /app
USER nimbus

EXPOSE 3000
ENV PORT=3000
ENV APP_ENV=production
ENV NIMBUS_SERVE=1

HEALTHCHECK --interval=15s --timeout=5s --retries=3 \
  CMD curl -fs http://localhost:3000/health || exit 1

ENTRYPOINT ["/app/server"]
