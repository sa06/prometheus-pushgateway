FROM golang:1.15

WORKDIR /go/src/app

COPY ./src ./src
COPY go.mod .
COPY go.sum .

RUN go build -o $GOPATH/bin/app ./src

CMD ["app"]