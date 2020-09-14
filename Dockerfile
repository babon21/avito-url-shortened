FROM golang:1.15

ENV GO111MODULE=on

ADD . /app
WORKDIR /app

RUN go mod download
RUN go build

CMD ["./avito-url-shortened"]
