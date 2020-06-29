FROM golang:1.14-alpine


RUN apk add --no-cache git musl-dev gcc g++ make ca-certificates

RUN go get -u github.com/discretemind/glink.git
WORKDIR /go/src/github.com/discretemind/glink.git

COPY . .

RUN GOARCH=amd64 GOOS=linux GO111MODULE=on make

FROM alpine:latest

#RUN apk --no-cache add openssl ca-certificates

WORKDIR /root/

COPY --from=0 /go/src/github.com/discretemind/glink.git .

RUN chmod +x ./glink

