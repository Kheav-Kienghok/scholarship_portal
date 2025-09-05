# ============================
# Stage 1: Build the Go binary
# ============================
FROM golang:1.24 AS builder

# Create non-root user
RUN useradd -m builderuser
USER builderuser

WORKDIR /app

# Create directories for logs and cache
RUN mkdir -p .cache logs

# Copy go module files and download dependencies
COPY --chown=builderuser:builderuser go.mod go.sum ./
RUN go mod download

# Copy source code
COPY --chown=builderuser:builderuser . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main main.go


# ============================
# Debug runtime (with shell tools)
# ============================
FROM alpine:3.19 AS debug

WORKDIR /app
COPY --from=builder /app/main .

# Install some useful debug tools
RUN apk add --no-cache curl bash

EXPOSE 8080
CMD ["./main"]

# ============================
# Production runtime (distroless)
# ============================
FROM gcr.io/distroless/static:nonroot AS prod

WORKDIR /app
COPY --from=builder /app/main .

# Create writable dirs
COPY --from=builder /app/main .
COPY --from=builder /app/logs ./logs
COPY --from=builder /app/.cache ./.cache

# Default Gin mode release
ENV GIN_MODE=release

USER 1000
EXPOSE 8080

CMD ["./main"]
