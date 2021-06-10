FROM golang:latest

RUN mkdir /app
ADD . /app

WORKDIR /app/divvy

RUN go mod download

RUN go build -o main

EXPOSE 8000

CMD ["/app/divvy/main"]