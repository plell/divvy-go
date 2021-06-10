FROM golang:latest

# RUN export GO111MODULE=on

RUN mkdir /app
ADD . /app

WORKDIR /app/divvy

# COPY ./go.mod ./
# COPY ./go.sum ./

RUN go mod download

RUN go build -o main

ENV PORT 8000

CMD ["/app/divvy/main"]