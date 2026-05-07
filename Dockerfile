# Multi-stage build for Grimorio MCP Server

# Stage 1: Builder
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -a -installsuffix cgo \
    -o grimorio ./cmd/grimorio

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -a -installsuffix cgo \
    -o migrate-v1-to-v2 ./cmd/migrate-v1-to-v2

# Stage 2: Final image with wkhtmltopdf
FROM alpine:3.20

RUN apk add --no-cache \
    ca-certificates \
    wkhtmltopdf \
    fontconfig \
    freetype \
    ttf-dejavu \
    libstdc++ \
    libx11 \
    libxrender \
    libxext \
    libssl3 \
    curl

WORKDIR /app

COPY --from=builder /build/grimorio /app/grimorio
COPY --from=builder /build/migrate-v1-to-v2 /app/migrate-v1-to-v2

# Create non-root user
RUN adduser -D -g '' grimorio
RUN mkdir -p /data /home/grimorio/.config/grimorio && \
    chown -R grimorio:grimorio /data /home/grimorio

USER grimorio

ENV GRIMORIO_DATA_DIR=/data
ENV PATH="/app:${PATH}"

# MCP servers communicate over stdio by default
# No EXPOSE needed for stdio, but expose for potential HTTP fallback
EXPOSE 8080

ENTRYPOINT ["/app/grimorio"]
