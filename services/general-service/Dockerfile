FROM golang:1.25.1-alpine3.22

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o /authorization-service ./cmd/main.go

# Expose the port
EXPOSE 8085

# Run the application
CMD [ "/authorization-service" ]