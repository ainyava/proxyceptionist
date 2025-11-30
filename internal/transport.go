package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/proxy"
)

// getTransport returns a transport with SOCKS5 or HTTP proxy
func getTransport() (http.RoundTripper, error) {
	proxyURL := os.Getenv("FORWARD_PROXY")
	if proxyURL == "" {
		return http.DefaultTransport, nil
	}

	// Detect if it's SOCKS5 or HTTP
	if proxyURL[:5] == "socks" {
		dialer, err := proxy.SOCKS5("tcp", proxyURL[9:], nil, proxy.Direct)
		if err != nil {
			return nil, err
		}
		return &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.Dial(network, addr)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}, nil
	}

	// Fallback to HTTP proxy
	proxyFunc := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyURL)
	}
	return &http.Transport{
		Proxy:           proxyFunc,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}, nil
}

func ProxyReq(c *gin.Context) {
	remoteURL, err := url.Parse(os.Getenv("REMOTE_URL"))
	if err != nil {
		c.String(http.StatusInternalServerError, "Invalid REMOTE URL")
		return
	}

	transport, err := getTransport()
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to create transport: %v", err))
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remoteURL)
	proxy.Transport = transport

	proxy.Director = func(req *http.Request) {
		// Copy headers from incoming request
		for k, v := range c.Request.Header {
			req.Header[k] = v
		}

		// Set custom Authorization header if env variable is set
		if auth := os.Getenv("AUTHORIZATION"); auth != "" {
			req.Header.Set("Authorization", auth)
		}

		req.Host = remoteURL.Host
		req.URL.Scheme = remoteURL.Scheme
		req.URL.Host = remoteURL.Host
		req.URL.Path = c.Param("proxyPath")
	}

	LogRequest(c, remoteURL.String())
	proxy.ServeHTTP(c.Writer, c.Request)
}
