# syntax=docker/dockerfile:1

# ── Stage 1: Build ───────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /airsenal ./cmd/airsenal

RUN git clone --depth 1 https://github.com/Orange-Cyberdefense/arsenal.git /arsenal-repo

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /airsenal /airsenal
COPY cheats/ /cheats/
COPY --from=builder /arsenal-repo/arsenal/data/cheats/ /arsenal-cheats/

ENV AIRSENAL_PORT=8080
ENV AIRSENAL_CHEATS_DIR=/cheats
ENV ARSENAL_CHEATS_DIR=/arsenal-cheats

EXPOSE 8080

ENTRYPOINT ["/airsenal"]
