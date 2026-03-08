package main

import (
	"github.com/mitchgrogg/rita-devtools/slothproxy/pkg/slothproxy"
	"github.com/spf13/cobra"
)

func buildRootCommand(p *slothproxy.SlothProxy) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slothproxy",
		Short: "An HTTP proxy that adds configurable latency to requests",
	}

	cmd.AddCommand(buildServeCommand(p))
	cmd.AddCommand(buildVersionCommand(p))

	return cmd
}

func buildServeCommand(p *slothproxy.SlothProxy) *cobra.Command {
	opts := slothproxy.ServeOptions{}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the proxy server",
		Long: `Start a proxy server that adds configurable latency to requests.

There are two modes of operation:

  Reverse proxy (--target): Forwards all incoming requests to a specific
  upstream server. Clients make normal requests to slothproxy's address.
    slothproxy serve --target http://api.example.com --latency 200ms

  Forward proxy (default): Acts as an HTTP proxy. Clients must configure
  their HTTP_PROXY to point at slothproxy.
    curl --proxy http://localhost:8080 http://example.com

Use --match to apply latency only to requests whose URL matches a regex pattern.
Without --match, all requests are throttled. For HTTP requests, the pattern is
matched against the full request URL path. For HTTPS in forward proxy mode, the
proxy only sees the host during CONNECT tunnel setup, so the pattern is matched
against the host:port (e.g. "example.com:443") rather than the full URL.

In forward proxy mode, use --forward to chain through an upstream proxy (e.g.
mitmproxy). --target and --forward are mutually exclusive.`,
		Example: `  # Reverse proxy to an upstream server with latency
  slothproxy serve --target http://api.example.com --latency 200ms

  # Forward proxy with latency
  slothproxy serve --port 8080 --latency 200ms

  # Forward proxy chained through mitmproxy
  slothproxy serve -l 500ms --forward localhost:8081

  # Only throttle requests matching a pattern
  slothproxy serve -l 500ms --match "api\.example\.com"
  slothproxy serve --target http://localhost:3000 -l 1s --match "/api/v2/"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Serve(cmd.Context(), opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Port, "port", "p", 8080, "port to listen on")
	cmd.Flags().DurationVarP(&opts.Latency, "latency", "l", 0, "latency to add to each request (e.g. 200ms, 1s)")
	cmd.Flags().StringVarP(&opts.Target, "target", "t", "", "upstream server URL for reverse proxy mode (e.g. http://localhost:3000)")
	cmd.Flags().StringVarP(&opts.Match, "match", "m", "", "regex pattern to select which URLs get latency applied")
	cmd.Flags().StringVarP(&opts.Forward, "forward", "f", "", "upstream proxy to forward traffic to in forward proxy mode (e.g. localhost:8080)")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "enable verbose logging")

	return cmd
}
