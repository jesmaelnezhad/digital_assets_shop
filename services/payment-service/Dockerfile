# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download || true

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /server .

# Final stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /

# Copy binary from builder
COPY --from=builder /server /server

# Expose port
EXPOSE 8086

# Run
CMD ["/server"]
