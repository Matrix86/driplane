FROM golang:alpine AS build-env

RUN apk add --update ca-certificates
RUN apk add --no-cache --update make

WORKDIR /go/src/app

COPY . .

RUN go get -d -v ./...

RUN make build

FROM alpine:latest

COPY --from=build-env /go/src/app/bin/driplane /app/

WORKDIR /app

ENTRYPOINT ["/app/driplane"]
