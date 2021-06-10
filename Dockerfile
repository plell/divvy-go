FROM golang:latest

RUN mkdir /app
ADD . /app

WORKDIR /app/divvy

RUN go mod download

RUN go build -o main

CMD ["/app/divvy/main"]