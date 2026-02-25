# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build a statically linked binary with debug symbols stripped
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o cloudflare-ddns .

# Final stage
FROM scratch

# CA certificates are required for HTTPS calls to Cloudflare API and ipify.org
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/cloudflare-ddns /cloudflare-ddns

ENTRYPOINT ["/cloudflare-ddns", "run"]
