FROM golang:alpine AS build

RUN apk add --update --no-cache git

RUN go get github.com/digitalocean/godo
RUN go get golang.org/x/oauth2

COPY main.go /go/src/
WORKDIR /go/src/

RUN go build -o ddns

FROM alpine

RUN apk add --update --no-cache ca-certificates && \
    update-ca-certificates

COPY --from=build /go/src/ddns /bin

EXPOSE 80 443

ENTRYPOINT ["/bin/ddns"]
