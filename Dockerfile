# Build Frontend
FROM node:20-alpine AS builder-node
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .

RUN npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css

# Build Backend (Go + Templ)
FROM golang:alpine AS builder-go
WORKDIR /app
RUN apk add --no-cache gcc musl-dev

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Generate templ templates
RUN templ generate
# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/web

# Final Runtime Image
FROM alpine:latest
WORKDIR /app

# Install CA certificates for HTTPS and sqlite libs
RUN apk add --no-cache ca-certificates sqlite

# Copy binary from builder-go
COPY --from=builder-go /app/main .

# Copy static assets
COPY --from=builder-node /app/static ./static

# Copy migrations for runtime use
COPY --from=builder-go /app/migrations ./migrations

# Set environment variables
ENV PORT=8080
ENV DB_PATH=/data/op_rating.db

# Expose the port
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
