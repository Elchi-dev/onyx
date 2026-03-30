# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -buildvcs=false \
    -o onyx ./cmd/onyx

# ── Final stage ───────────────────────────────────────────────────────────────
FROM scratch

# CA certificates for Let's Encrypt ACME requests.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /build/onyx /onyx

# Proxy HTTP, proxy HTTPS, dashboard
EXPOSE 80 443 8080

VOLUME ["/data"]

ENTRYPOINT ["/onyx"]
CMD ["start", "--config", "/data/onyx.toml"]
