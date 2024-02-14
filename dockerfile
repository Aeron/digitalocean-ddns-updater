### Builder stage ###
FROM docker.io/library/golang:1.22-bookworm AS build-env

ENV DEBIAN_FRONTEND noninteractive

RUN apt update -qq && apt install -y --no-install-recommends \
    ca-certificates \
    git
RUN rm -r /var/lib/apt/lists /var/cache/apt/archives

WORKDIR /go/src

COPY . .

RUN go mod download && \
    go mod verify

ENV CGO_ENABLED 0

RUN go build \
    -ldflags '-extldflags "-static"' \
    -tags timetzdata \
    -o /go/bin \
    ./...

### Runtime stage ###
FROM scratch

LABEL org.opencontainers.image.source https://github.com/Aeron/digitalocean-ddns-updater

COPY --from=build-env /etc/ssl/certs /etc/ssl/certs
COPY --from=build-env /go/bin/app /app

ENTRYPOINT ["/app"]
