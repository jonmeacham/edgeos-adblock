package edgeos

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// Env is struct of parameters
type Env struct {
	ctr
	Log      Logger          `json:"-"`
	HTTPCtx  context.Context `json:"-"`
	HTTP     *http.Client    `json:"-"`
	API      string          `json:"API,omitempty"`
	Bash     string          `json:"Bash,omitempty"`
	Disabled bool            `json:"Disabled"`
	Dbug     bool            `json:"Dbug,omitempty"`
	Dex      *list           `json:"Dex,omitempty"`
	Dir      string          `json:"Dir,omitempty"`
	DNSsvc   string          `json:"dnsmasq service,omitempty"`
	Exc      *list           `json:"Exc,omitempty"`
	Ext      string          `json:"dnsmasq fileExt.,omitempty"`
	InCLI    string          `json:"-"`
	Method   string          `json:"HTTP method,omitempty"`
	Pfx      dnsPfx          `json:"Prefix"`
	Timeout  time.Duration   `json:"Timeout,omitempty"`
}

// dnsPfx defines the prefix entries in the dnsmasq configuration file
type dnsPfx struct {
	domain string
}

// Debug logs debug messages when the Dbug flag is true
func (e *Env) Debug(s ...any) {
	if !e.Dbug || e.Log == nil {
		return
	}
	e.Log.Debug(s...)
}

// Option is a recursive function
type Option func(c *Config) Option

// API sets the EdgeOS CLI API command
func API(s string) Option {
	return func(c *Config) Option {
		previous := c.API
		c.API = s
		return API(previous)
	}
}

// Bash sets the shell processor
func Bash(s string) Option {
	return func(c *Config) Option {
		previous := c.Bash
		c.Bash = s
		return Bash(previous)
	}
}

// Disabled toggles Disabled
func Disabled(b bool) Option {
	return func(c *Config) Option {
		previous := c.Disabled
		c.Disabled = b
		return Disabled(previous)
	}
}

// Dbug toggles Debug level on or off
func Dbug(b bool) Option {
	return func(c *Config) Option {
		previous := c.Dbug
		c.Dbug = b
		return Dbug(previous)
	}
}

// Dir sets directory location
func Dir(s string) Option {
	return func(c *Config) Option {
		previous := c.Dir
		c.Dir = s
		return Dir(previous)
	}
}

// DNSsvc sets dnsmasq restart command
func DNSsvc(s string) Option {
	return func(c *Config) Option {
		previous := c.DNSsvc
		c.DNSsvc = s
		return DNSsvc(previous)
	}
}

// Ext sets the adblock file n extension
func Ext(s string) Option {
	return func(c *Config) Option {
		previous := c.Ext
		c.Ext = s
		return Ext(previous)
	}
}

// InCLI sets the CLI inSession command
func InCLI(s string) Option {
	return func(c *Config) Option {
		previous := c.InCLI
		c.InCLI = s
		return InCLI(previous)
	}
}

// SetLogger wires the logger implementation (typically slog-backed from main).
func SetLogger(l Logger) Option {
	return func(c *Config) Option {
		previous := c.Log
		c.Log = l
		return SetLogger(previous)
	}
}

// Context sets the parent context for outbound HTTP downloads (timeouts stack with request timeout).
func Context(ctx context.Context) Option {
	return func(c *Config) Option {
		previous := c.HTTPCtx
		c.HTTPCtx = ctx
		return Context(previous)
	}
}

// HTTPClient sets the HTTP client for downloads (nil uses [defaultHTTPClient]).
func HTTPClient(client *http.Client) Option {
	return func(c *Config) Option {
		previous := c.HTTP
		c.HTTP = client
		return HTTPClient(previous)
	}
}

// Method sets the HTTP method
func Method(s string) Option {
	return func(c *Config) Option {
		previous := c.Method
		c.Method = s
		return Method(previous)
	}
}

// NewConfig returns a new *Config initialized with the parameter options passed to it
func NewConfig(opts ...Option) *Config {
	c := Config{
		Env: &Env{
			ctr: ctr{RWMutex: &sync.RWMutex{}, stat: make(stat)},
			Dex: &list{RWMutex: &sync.RWMutex{}, entry: make(entry)},
			Exc: &list{RWMutex: &sync.RWMutex{}, entry: make(entry)},
		},
	}
	for _, opt := range opts {
		opt(&c)
	}
	return &c
}

// Prefix sets the dnsmasq configuration address line prefix
func Prefix(d string) Option {
	return func(c *Config) Option {
		previous := c.Pfx.domain
		c.Pfx = dnsPfx{domain: d}
		return Prefix(previous)
	}
}

// Timeout sets how long before an unresponsive goroutine is aborted
func Timeout(t time.Duration) Option {
	return func(c *Config) Option {
		previous := c.Timeout
		c.Timeout = t
		return Timeout(previous)
	}
}
