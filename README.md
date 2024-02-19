# DigitalOcean DDNS Updater

A tiny DDNS updater (service) for [DigitalOcean][digitalocean], primarily intended to
be used as a container image.

The priority is to be compatible with [Tomato][freshtomato] firmware. Yet, it probably
will work with [OpenWRT][openwrt] and [DD-WRT][dd-wrt] as well.

As a security measure, the updater uses a simple token as a first authentication factor
and a strict incoming request rate limiter to mitigate a brute-force attack risk.

[openwrt]: https://openwrt.org
[dd-wrt]: https://dd-wrt.com

## Motivation

As a human being with [Tomato][freshtomato] firmware on all my routers who need access
to a home network once a year, I need an alternative to stop paying my ISP for a static
IP address. And you’re right—it sounds like DDNS.

But I‘ve been managing my domains on [DigitalOcean][digitalocean] for a long time,
so why use any 3rd-party service? A self-hosted solution is clearly preferable here.
Yet, I don’t recall finding any suitable options at that time.

Also, in 2017, I just began skimming Go, and it was an excellent opportunity to make
something small yet usable, a tiny pet project. And so here it is.

It wasn’t pretty back then, but it worked well enough. After [version 1.0][release-1.0],
I haven’t touched it for years, but now it looks okay as well, somewhat representative.

If it fits your DDNS needs, give it a try.

[freshtomato]: https://freshtomato.org
[digitalocean]: https://digitalocean.com
[release-1.0]: https://github.com/Aeron/digitalocean-ddns-updater/releases/tag/1.0.0

## Usage

The container image is available as
[`docker.io/aeron/digitalocean-ddns-updater`][docker] and
[`ghcr.io/Aeron/digitalocean-ddns-updater`][github]. You can use them both
interchangeably.

```sh
docker pull docker.io/aeron/digitalocean-ddns-updater
# …or…
docker pull ghcr.io/aeron/digitalocean-ddns-updater
```

[docker]: https://hub.docker.com/r/aeron/digitalocean-ddns-updater
[github]: https://github.com/Aeron/digitalocean-ddns-updater/pkgs/container/digitalocean-ddns-updater

### Container Running

Start a container along with the `DIGITALOCEAN_API_TOKEN` environment variable supplied:

```sh
docker run -d --restart unless-stopped --name ddns \
    -p 80:8080/tcp \
    -e DIGITALOCEAN_API_TOKEN=$DIGITALOCEAN_API_TOKEN \
    aeron/digitalocean-ddns-updater:latest
```

Or, without publishing ports, in case of a reverse proxy is handling things.

Optionally, [other enviroment variables](#application-options) can be provided.

### Application Running

Although it’s not intended to run the app outside a container, it’s still 100% doable.
But, probably, it’ll require compiling the app first.

### Application Options

The application supports the following options:

| Argument                  | Environment Variable     | Default      | Type    |
| ------------------------- | ------------------------ | ------------ | ------- |
| `-address`                | -                        | `:8080`      | string  |
| `-endpoint`               | -                        | `/ddns`      | string  |
| `-digitalocean-api-token` | `DIGITALOCEAN_API_TOKEN` | **Required** | string  |
| `-security-token`         | `SECURITY_TOKEN`         | -            | string  |
| `-limit-rps`              | `LIMIT_RPS`              | `.01`        | float64 |
| `-limit-burst`            | `LIMIT_BURST`            | `1`          | integer |

You can use a container’s `CMD` directive to supply arguments as well.

The default RPS limit value is `0.01`, and the default burst limit is `1`. It means the
updater will accept one request per 100 seconds. Also, the delay will increase for
another 100 seconds on each failed attempt.

Because the token parameter is optional, the updater can generate SHA512 checksum hash
from the provided DigitalOcean API token. Considering that a DO API token is private,
it should be safe. In this case, a new security token will be in container logs on
start.

### Router Configuration

In order to make it work, the app’s endpoint will expect the following URL query
parameters:

- `type` (optional) — a DNS record type to update (default: `A`; choices: `AAAA` and
  `A`);
- `domain` — a domain name to update with a supplied IP address;
- `ip` — an IP address to supply a new value for a provided domain;
- `token` — a security token, which must be the same as the updater.

In case of Tomato, a custom URL must look similar to the following:

```text
https://example.net/ddns?domain=home.example.net&type=A&ip=@IP&token=123abc
```

> [!NOTE]
> Parameters `type` and `ip` work in pair, so you cannot specify an IPv6 address for
> the `A` record and vise-versa.

The `token` must be an URL-safe string and long enough. The `@IP` part is a standard
placeholder for an IP address that Tomato will replace with the real one.

> [!WARNING]
> Plain HTTP use is not even considered here. It must run behind HTTPS (reverse proxy,
> balancer, or whatever you have).
