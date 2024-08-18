FROM golang:latest

WORKDIR /usr/src/app/reader

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /logoutput-reader

EXPOSE 8080
CMD ["/logoutput-reader"]