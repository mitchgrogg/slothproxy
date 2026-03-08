package slothproxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"time"

	"github.com/elazarl/goproxy"
)

const Version = "0.1.0"

type SlothProxy struct{}

type ServeOptions struct {
	Port    int
	Latency time.Duration
	Target  string
	Forward string
	Match   string
	Verbose bool
}

func New() *SlothProxy {
	return &SlothProxy{}
}

func (p *SlothProxy) Serve(ctx context.Context, opts ServeOptions) error {
	if opts.Target != "" && opts.Forward != "" {
		return fmt.Errorf("--target and --forward are mutually exclusive")
	}

	var handler http.Handler
	var mode string

	if opts.Target != "" {
		h, err := p.buildReverseProxy(opts)
		if err != nil {
			return err
		}
		handler = h
		mode = "reverse proxy"
	} else {
		h, err := p.buildForwardProxy(opts)
		if err != nil {
			return err
		}
		handler = h
		mode = "forward proxy"
	}

	addr := fmt.Sprintf(":%d", opts.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	msg := fmt.Sprintf("Starting slothproxy on %s (%s)", addr, mode)
	if opts.Latency > 0 {
		msg += fmt.Sprintf(", latency %v", opts.Latency)
	}
	if opts.Match != "" {
		msg += fmt.Sprintf(", match %q", opts.Match)
	}
	if opts.Target != "" {
		msg += fmt.Sprintf(", target %s", opts.Target)
	}
	if opts.Forward != "" {
		msg += fmt.Sprintf(", forwarding to %s", opts.Forward)
	}
	log.Println(msg)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("proxy server error: %w", err)
	}

	return nil
}

func sleepRemaining(latency, elapsed time.Duration) {
	if remaining := latency - elapsed; remaining > 0 {
		time.Sleep(remaining)
	}
}

func (p *SlothProxy) buildReverseProxy(opts ServeOptions) (http.Handler, error) {
	targetURL, err := url.Parse(opts.Target)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}

	var matchRe *regexp.Regexp
	if opts.Match != "" {
		matchRe, err = regexp.Compile(opts.Match)
		if err != nil {
			return nil, fmt.Errorf("invalid match pattern: %w", err)
		}
	}

	rp := httputil.NewSingleHostReverseProxy(targetURL)
	if opts.Verbose {
		rp.ErrorLog = log.Default()
	}

	if opts.Latency <= 0 {
		return rp, nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if matchRe == nil || matchRe.MatchString(r.URL.String()) {
			start := time.Now()
			rp.ServeHTTP(w, r)
			sleepRemaining(opts.Latency, time.Since(start))
			return
		}
		rp.ServeHTTP(w, r)
	}), nil
}

func (p *SlothProxy) buildForwardProxy(opts ServeOptions) (http.Handler, error) {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = opts.Verbose

	if opts.Forward != "" {
		forwardURL, err := url.Parse("http://" + opts.Forward)
		if err != nil {
			return nil, fmt.Errorf("invalid forward address: %w", err)
		}
		proxy.Tr = &http.Transport{
			Proxy: http.ProxyURL(forwardURL),
		}
		proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler("http://"+opts.Forward, nil)
	}

	var matchRe *regexp.Regexp
	if opts.Match != "" {
		var err error
		matchRe, err = regexp.Compile(opts.Match)
		if err != nil {
			return nil, fmt.Errorf("invalid match pattern: %w", err)
		}
	}

	if opts.Latency > 0 {
		proxy.OnRequest().DoFunc(
			func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				if matchRe == nil || matchRe.MatchString(r.URL.String()) {
					ctx.UserData = time.Now()
				}
				return r, nil
			})

		proxy.OnResponse().DoFunc(
			func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
				if start, ok := ctx.UserData.(time.Time); ok {
					sleepRemaining(opts.Latency, time.Since(start))
				}
				return resp
			})

		proxy.OnRequest().HandleConnectFunc(
			func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
				if matchRe == nil || matchRe.MatchString(host) {
					time.Sleep(opts.Latency)
				}
				return goproxy.OkConnect, host
			})
	}

	return proxy, nil
}
