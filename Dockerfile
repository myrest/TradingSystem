# Use an official Golang runtime as a parent image
FROM golang

COPY ./ /app

WORKDIR /app

RUN go mod download

# Build the Go app
RUN go build -o ./src/main ./src/main.go

# Expose port 8080 to the outside world
EXPOSE 8080

WORKDIR /app/src

# Command to run the executable
CMD ["./main"]