# Stage 1: Build the Go application
FROM docker.io/golang:1.22 AS builder

ARG KUBESEAL_VERSION=0.27.0

# Download and install Kubeseal
RUN curl --noproxy '*' -k https://github.com/bitnami-labs/sealed-secrets/releases/download/v${KUBESEAL_VERSION}/kubeseal-${KUBESEAL_VERSION}-linux-amd64.tar.gz --output /tmp/kubeseal-linux-amd64.tar.gz && \
    tar xzvf /tmp/kubeseal-linux-amd64.tar.gz -C /tmp && \
    chmod +x /tmp/kubeseal && \
    rm /tmp/kubeseal-linux-amd64.tar.gz

WORKDIR /app

# Copy the Go module files and download dependencies
COPY app/go.mod .
RUN go mod download

# Copy the application source code
COPY app/main.go main.go

# Build the Go application as a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Stage 2: Create a minimal Docker image
FROM gcr.io/distroless/static:nonroot

# Copy the static binary from the builder stage
COPY --from=builder /app/app /app/app
COPY --from=builder /tmp/kubeseal /usr/bin/kubeseal
COPY app/static /app/static
COPY app/templates /app/templates
COPY app/hack/config /.kube/config

# Expose the port the application listens on
EXPOSE 8080

# Set environment variables
WORKDIR /app

# Command to run the application
CMD ["/app/app"]