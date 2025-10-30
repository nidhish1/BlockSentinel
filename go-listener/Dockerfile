FROM golang:1.24-alpine

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]