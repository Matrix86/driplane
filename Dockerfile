FROM golang:1.27.2-alpine AS build-env

RUN apk add --update ca-certificates
RUN apk add --no-cache --update make

WORKDIR /go/src/app

COPY . .

RUN go mod download

RUN make build

FROM alpine:latest

COPY --from=build-env /go/src/app/bin/driplane /app/

WORKDIR /app

ENTRYPOINT ["/app/driplane"]
