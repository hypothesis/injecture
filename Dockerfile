FROM golang:1.8-alpine
MAINTAINER Hypothes.is Project and contributors

RUN apk add --no-cache --update git
RUN go get -u github.com/golang/dep/...

WORKDIR /go/src/app

COPY . .
RUN dep ensure
RUN go build .

CMD ./app
