package edgeos

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

type HTTPserver struct {
	Mux    *http.ServeMux
	Server *httptest.Server
}

func (h *HTTPserver) NewHTTPServer() *url.URL {
	h.Mux = http.NewServeMux()
	h.Server = httptest.NewServer(h.Mux)
	u, _ := url.Parse(h.Server.URL)
	return u
}

func TestGetHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "example.invalid\n")
	}))
	defer srv.Close()

	o := download(&source{Env: &Env{Log: newLog(), Method: http.MethodGet}, url: srv.URL})
	if o.err != nil {
		t.Fatal(o.err)
	}
	b, err := io.ReadAll(o.r)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), "example.invalid\n"; got != want {
		t.Fatalf("body %q, want %q", got, want)
	}
}

func TestGetHTTPRejectsBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	o := download(&source{Env: &Env{Log: newLog(), Method: http.MethodGet}, url: srv.URL})
	if o.err == nil || !strings.Contains(o.err.Error(), "unexpected HTTP status") {
		t.Fatalf("got %v, want HTTP status error", o.err)
	}
}

type countingHandler struct {
	sync.Mutex
	count int
}

func (h *countingHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.Lock()
	defer h.Unlock()
	h.count++
	fmt.Fprintf(w, "request %d", h.count)
}

func TestHTTPHandler(t *testing.T) {
	srv := httptest.NewServer(&countingHandler{})
	defer srv.Close()

	for _, want := range []string{"request 1", "request 2"} {
		resp, err := http.Get(srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		b, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(b) != want {
			t.Fatalf("got %q, want %q", b, want)
		}
	}
}

func TestDownloadUsesInjectedHTTPClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "injected-ok")
	}))
	defer srv.Close()

	o := download(&source{
		Env: &Env{Log: newLog(), Method: http.MethodGet, HTTP: srv.Client()},
		url: srv.URL,
	})
	if o.err != nil {
		t.Fatal(o.err)
	}
	b, err := io.ReadAll(o.r)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "injected-ok" {
		t.Fatalf("body %q", b)
	}
}
