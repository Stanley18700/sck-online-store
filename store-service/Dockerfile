# syntax=docker/dockerfile:1.7
# Optimized store-service Dockerfile.
# Key wins vs. previous version:
#   1. BuildKit cache mounts for go module cache and build cache
#      -> `go mod download` (~60s) skipped on warm builds
#   2. Pinned base + pinned swag version (no `@latest` re-resolution)
#   3. Removed `go get -u` (which could mutate go.sum and bust cache)
#   4. Tools layer separated from source layer (source edits don't
#      re-fetch swag or its deps)

# ---------- Stage 1: install swag (only depends on Go version) ----------
FROM golang:1.26.5-alpine3.24 AS tools
RUN apk add --no-cache git
# Pin to a specific swag version for reproducible builds
ARG SWAG_VERSION=v1.16.3
RUN GOBIN=/usr/local/bin go install github.com/swaggo/swag/cmd/swag@${SWAG_VERSION}

# ---------- Stage 2: build the binary ----------
FROM golang:1.26.5-alpine3.24 AS builder
WORKDIR /module

# Bring the pinned swag binary in from the tools stage.
COPY --from=tools /usr/local/bin/swag /usr/local/bin/swag

# Copy module manifests first so a source-only edit doesn't
# invalidate the module download cache. Cache mounts are attached
# to the next RUN so the module + build caches survive across
# builds (skipping the ~60s `go mod download` on warm builds).
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Now bring in the rest of the source. Anything here invalidates
# only the swag-init + go-build layers, not the module cache.
COPY . .

RUN swag init -g cmd/main.go -o cmd/docs

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o app ./cmd/main.go

# ---------- Stage 3: minimal runtime ----------
FROM alpine:3.24
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Bangkok /etc/localtime && \
    echo "Asia/Bangkok" > /etc/timezone
WORKDIR /root/
COPY --from=builder /module/app .
ENV GIN_MODE=release
EXPOSE 8000
CMD ["./app"]