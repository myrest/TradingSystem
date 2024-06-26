# Use an official Golang runtime as a parent image
FROM golang

COPY ./ /app

WORKDIR /app

RUN go mod init TradingSystem
RUN go mod tidy

# Build the Go app
RUN go build -o main ./src/main.go

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]