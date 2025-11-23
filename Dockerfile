# Stage 1: Build the application
FROM golang:1.25.4-alpine AS builder

# Install git
RUN apk add --no-cache git

WORKDIR /app

# Copy dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Final Runtime Image
FROM alpine:latest

# Install dependencies:
# - ca-certificates: for HTTPS
# - postgresql-client: contains 'psql' for our init script
# - bash: to run the script
RUN apk --no-cache add ca-certificates postgresql-client bash

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Copy the init script from your local folder
COPY scripts/init_db.sh ./init_db.sh

# Make script executable
RUN chmod +x ./init_db.sh

EXPOSE 1326

# Run the script to wait/create DB, THEN start the app
CMD ["/bin/sh", "-c", "./init_db.sh && ./main"]