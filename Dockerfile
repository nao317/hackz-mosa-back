# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS dependencies

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

FROM dependencies AS builder

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/api \
    ./cmd/api

FROM alpine:3.23 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 1000 render-secrets \
    && addgroup -S app \
    && adduser -S app -G app \
    && addgroup app render-secrets
USER app

COPY --from=builder --chown=app:app /out/api /usr/local/bin/api

EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O - http://127.0.0.1:8080/health >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
