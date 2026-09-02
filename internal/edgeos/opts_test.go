package edgeos

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

type discardLogger struct{}

func (discardLogger) Debug(...any)             {}
func (discardLogger) Info(...any)              {}
func (discardLogger) Infof(string, ...any)     {}
func (discardLogger) Warning(...any)           {}
func (discardLogger) Warningf(string, ...any)  {}
func (discardLogger) Error(...any)             {}
func (discardLogger) Errorf(string, ...any)    {}
func (discardLogger) Noticef(string, ...any)   {}
func (discardLogger) Criticalf(string, ...any) {}

func newLog() Logger { return discardLogger{} }

type bufLogger struct{ buf *bytes.Buffer }

func (b *bufLogger) Debug(args ...any)                    { fmt.Fprintln(b.buf, args...) }
func (b *bufLogger) Info(args ...any)                     { fmt.Fprintln(b.buf, args...) }
func (b *bufLogger) Infof(format string, args ...any)     { fmt.Fprintf(b.buf, format+"\n", args...) }
func (b *bufLogger) Warning(args ...any)                  { fmt.Fprintln(b.buf, args...) }
func (b *bufLogger) Warningf(format string, args ...any)  { fmt.Fprintf(b.buf, format+"\n", args...) }
func (b *bufLogger) Error(args ...any)                    { fmt.Fprintln(b.buf, args...) }
func (b *bufLogger) Errorf(format string, args ...any)    { fmt.Fprintf(b.buf, format+"\n", args...) }
func (b *bufLogger) Noticef(format string, args ...any)   { fmt.Fprintf(b.buf, format+"\n", args...) }
func (b *bufLogger) Criticalf(format string, args ...any) { fmt.Fprintf(b.buf, format+"\n", args...) }

func TestDebugLogging(t *testing.T) {
	var out bytes.Buffer
	env := NewConfig(SetLogger(&bufLogger{buf: &out}), Dbug(true)).Env
	env.Debug("debug message")
	if got, want := out.String(), "debug message\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	out.Reset()
	env.Dbug = false
	env.Debug("hidden")
	if out.Len() != 0 {
		t.Fatalf("disabled debug wrote %q", out.String())
	}
}

func TestOptions(t *testing.T) {
	c := NewConfig(
		API("/bin/cli-shell-api"),
		Bash("/bin/bash"),
		Dir("/tmp"),
		DNSsvc("service dnsmasq restart"),
		Ext("edgeos-adblock.conf"),
		InCLI("inSession"),
		Method("GET"),
		Prefix("address="),
		Timeout(30*time.Second),
	)
	if c.API != "/bin/cli-shell-api" || c.Dir != "/tmp" || c.Timeout != 30*time.Second {
		t.Fatalf("options were not applied: %#v", c.Env)
	}
	if c.Pfx.domain != "address=" {
		t.Fatalf("prefixes = %#v", c.Pfx)
	}
}
