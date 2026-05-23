# Multi-stage build for Tibia-Char Go application
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Copy go.mod and go.sum first to leverage Docker cache
COPY app/go.mod app/go.sum ./app/

WORKDIR /src/app
RUN go mod download

# Copy the rest of the application source code
COPY app/ /src/app/

# Build the main server binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /src/tibia-char-server ./cmd/tibia-char/main.go

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /src/tibia-char-server .

# Copy UI templates and assets
COPY --from=builder /src/app/ui ./ui

# Expose server port
EXPOSE 8080

CMD ["./tibia-char-server"]
