# Build Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

# Cache dependencies first to speed up subsequent builds
COPY third_party/p4rse_tan ./third_party/p4rse_tan
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code
COPY . .

# Build with optimization flags and compress with UPX
# -s: disable symbol table | -w: disable DWARF generation
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-s -w" -o campaign_cli campaign.go

# Final Stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

WORKDIR /app

# Copy the optimized binary from the builder
COPY --from=builder /app/campaign_cli .

ENTRYPOINT ["./campaign_cli"]