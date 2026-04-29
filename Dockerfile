# syntax=docker/dockerfile:1

# ── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

WORKDIR /src

# Cache dependency downloads separately from source compilation.
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /airsenal ./cmd/airsenal

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM scratch

# Import CA certificates for any future HTTPS calls.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the statically-linked binary.
COPY --from=builder /airsenal /airsenal

# Bundle the default cheatsheets; users can override via a volume mount.
COPY cheats/ /cheats/

EXPOSE 8080

ENTRYPOINT ["/airsenal"]
