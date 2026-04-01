# ── Build stage ─────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o rootrevolutionapi ./cmd/main.go

# ── Runtime stage ────────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/rootrevolutionapi .
COPY --from=builder /build/mockup ./mockup

ENV TZ=Africa/Johannesburg

EXPOSE 3000

CMD ["./rootrevolutionapi"]
