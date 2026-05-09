package edgeos

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// MaxDownloadBytes caps HTTP response bodies when fetching blocklists (DoS guard on routers).
const MaxDownloadBytes = 64 << 20 // 64 MiB

// NewDefaultHTTPClient returns a shared-policy client: TLS 1.2+, sane transport timeouts.
// Request deadlines should come from context (see download); Client.Timeout is left unset.
func NewDefaultHTTPClient() *http.Client {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return &http.Client{Transport: tr}
}

var defaultHTTPClient = NewDefaultHTTPClient()
