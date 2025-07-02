# Dockerfile for go-airflow-trigger backend

# Build stage
ARG GO_VERSION=1.21
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the Go binary
RUN go build -o go-airflow-trigger ./main.go

# Final stage
FROM alpine:latest

# Create non-root user
RUN adduser -D appuser

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/go-airflow-trigger .

# Copy any static folders if needed (optional)
# COPY --from=builder /app/templates ./templates

# Set executable permissions
RUN chmod +x go-airflow-trigger

RUN mkdir -p /app/data/incoming

USER appuser

EXPOSE 8081

CMD ["./go-airflow-trigger"]
