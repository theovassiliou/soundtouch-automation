# Stage 1: Build the Go application
FROM golang:alpine AS build
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum  ./
RUN go mod download

# Copy the source code
COPY . .

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

# Build the Go application
RUN go build -ldflags="-s -w" -o myapp .

# Stage 2: Create a minimal runtime image
FROM --platform=linux/amd64 alpine
WORKDIR /app

# Copy the built executable from the build stage
COPY --from=build /app/myapp .
COPY --from=build /app/config-docker.toml .

# Set the entrypoint with command line arguments
ENTRYPOINT [ "./myapp" ]

# Define default command line arguments
# CMD ["-l", "debug"]