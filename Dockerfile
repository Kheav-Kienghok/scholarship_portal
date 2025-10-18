# ============================
# Stage 1: Build the Go binary
# ============================
FROM golang:1.25 AS builder

WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and migrations
COPY . .

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

# Copy only the binary
COPY --from=builder /app/main .

USER 1000

EXPOSE 8080
CMD ["./main"]