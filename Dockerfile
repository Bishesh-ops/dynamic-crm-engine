# Stage 1: Build the Go Binary
FROM golang:1.25.10-alpine AS builder

# Install git and ca-certificates (needed for fetching dependencies)
RUN apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependencies first (improves build times)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically linked binary. 
# CGO_ENABLED=0 ensures it doesn't depend on local C libraries.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o crm-engine ./cmd/api/main.go

# Stage 2: Create the Micro-Container
FROM alpine:latest

# We need CA certs for HTTPS and tzdata for timezones
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/crm-engine .
# Copy the initialization SQL file so the app can find it at runtime
COPY --from=builder /app/cmd/api/migrations ./cmd/api/migrations

# Expose the API port
EXPOSE 8080

# Run the binary
CMD ["./crm-engine"]
