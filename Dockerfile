FROM golang:alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /main ./src/main.go

FROM scratch

COPY --from=builder /main /main

EXPOSE 8080

CMD ["/main"]