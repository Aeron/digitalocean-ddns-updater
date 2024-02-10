### Builder stage ###
FROM docker.io/library/golang:1.22-bookworm AS build-env

ENV DEBIAN_FRONTEND noninteractive

RUN apt update -qq && apt install -y --no-install-recommends \
    ca-certificates \
    git
RUN rm -r /var/lib/apt/lists /var/cache/apt/archives

WORKDIR /go/src/do-ddns-up

COPY internal/ internal/
COPY go.mod .
COPY go.sum .

RUN go mod download && \
    go mod verify

ENV CGO_ENABLED 0

RUN go build \
    -ldflags '-extldflags "-static"' \
    -tags timetzdata \
    -o /go/bin/do-ddns-up \
    ./internal

### Runtime stage ###
FROM scratch

LABEL org.opencontainers.image.source https://github.com/Aeron/digitalocean-ddns-updater

COPY --from=build-env /etc/ssl/certs /etc/ssl/certs
COPY --from=build-env /go/bin/do-ddns-up /do-ddns-up

ENTRYPOINT ["/do-ddns-up"]
