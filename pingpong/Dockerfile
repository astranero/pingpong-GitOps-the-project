FROM golang:latest

WORKDIR /usr/src/app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /pong

EXPOSE 8081
CMD ["/pong"]