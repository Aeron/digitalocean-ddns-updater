FROM golang:1.18-bullseye AS build-env

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update -qq && apt-get install -y --no-install-recommends \
    ca-certificates \
    git
RUN rm -r /var/lib/apt/lists /var/cache/apt/archives

WORKDIR /go/src/do-ddns-up

COPY main.go .
COPY go.mod .
COPY go.sum .

RUN go mod download && \
    go mod verify

ENV CGO_ENABLED 0

RUN go build \
    -ldflags '-extldflags "-static"' \
    -tags timetzdata \
    -o /go/bin/do-ddns-up

# An actual image

FROM scratch

LABEL org.opencontainers.image.source https://github.com/Aeron/digitalocean-ddns-updater

COPY --from=build-env /etc/ssl/certs /etc/ssl/certs
COPY --from=build-env /go/bin/do-ddns-up /do-ddns-up

ENTRYPOINT ["/do-ddns-up"]
