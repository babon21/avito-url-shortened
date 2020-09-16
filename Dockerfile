FROM golang:1.15


ADD . /app
WORKDIR /app

RUN go build -v

CMD ["./avito-url-shortened"]
