# slothproxy

slothproxy is a proxy server that adds configurable latency to requests, useful for testing how clients behave on slow connections. It supports two modes:

## Reverse proxy mode

Forwards all incoming requests to a specific upstream server. Clients make normal requests to slothproxy's address.

```bash
# All requests to localhost:8080 are forwarded to api.example.com with 500ms latency
slothproxy serve --target http://api.example.com --latency 500ms
```

## Forward proxy mode

Acts as a standard HTTP proxy. Clients configure their `HTTP_PROXY` to point at slothproxy.

```bash
# Start the forward proxy with 1s latency
slothproxy serve --latency 1s

# In another terminal
curl --proxy http://localhost:8080 http://example.com
```

## Chaining with mitmproxy

Use `--forward` in forward proxy mode to chain through an upstream proxy like mitmproxy, giving you both latency simulation and traffic inspection:

```bash
slothproxy serve --latency 500ms --port 9999 --forward localhost:8080
```

Clients point at `localhost:9999` -> slothproxy adds latency -> mitmproxy on `localhost:8080` inspects traffic -> internet.

## Selective throttling

Use `--match` to only throttle requests matching a regex pattern. Unmatched requests pass through at full speed.

```bash
# Only throttle API requests
slothproxy serve --target http://localhost:3000 --latency 2s --match "/api/"

# Only throttle a specific domain
slothproxy serve --latency 500ms --match "slow\.example\.com"
```

In forward proxy mode with HTTPS, `--match` is matched against the `host:port` (e.g. `example.com:443`) since the full URL isn't visible during CONNECT tunnel setup.

## How latency works

The `--latency` value represents the minimum total request time, not additional delay. If you specify `--latency 5s` and the upstream responds in 2s, slothproxy sleeps for the remaining 3s. If the upstream takes longer than the specified latency, no extra delay is added.

## All options

```
slothproxy serve [flags]

Flags:
  -f, --forward string     upstream proxy to forward traffic to in forward proxy mode
  -l, --latency duration   latency to add to each request (e.g. 200ms, 1s)
  -m, --match string       regex pattern to select which URLs get latency applied
  -p, --port int           port to listen on (default 8080)
  -t, --target string      upstream server URL for reverse proxy mode
      --verbose            enable verbose logging
```
