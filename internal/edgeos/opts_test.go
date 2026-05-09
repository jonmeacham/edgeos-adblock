package edgeos

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

// discardLogger implements Logger with no output (for tests that only need wiring).
type discardLogger struct{}

func (discardLogger) Debug(args ...any)                    {}
func (discardLogger) Info(args ...any)                     {}
func (discardLogger) Infof(format string, args ...any)     {}
func (discardLogger) Warning(args ...any)                  {}
func (discardLogger) Warningf(format string, args ...any)  {}
func (discardLogger) Error(args ...any)                    {}
func (discardLogger) Errorf(format string, args ...any)    {}
func (discardLogger) Noticef(format string, args ...any)   {}
func (discardLogger) Criticalf(format string, args ...any) {}

func newLog() Logger {
	return discardLogger{}
}

type bufLogger struct {
	buf *bytes.Buffer
}

func (b *bufLogger) Debug(args ...any) {
	fmt.Fprintln(b.buf, args...)
}

func (b *bufLogger) Info(args ...any) {
	fmt.Fprintln(b.buf, args...)
}

func (b *bufLogger) Infof(format string, args ...any) {
	fmt.Fprintf(b.buf, format+"\n", args...)
}

func (b *bufLogger) Warning(args ...any) {
	fmt.Fprintln(b.buf, args...)
}

func (b *bufLogger) Warningf(format string, args ...any) {
	fmt.Fprintf(b.buf, format+"\n", args...)
}

func (b *bufLogger) Error(args ...any) {
	fmt.Fprintln(b.buf, args...)
}

func (b *bufLogger) Errorf(format string, args ...any) {
	fmt.Fprintf(b.buf, format+"\n", args...)
}

func (b *bufLogger) Noticef(format string, args ...any) {
	fmt.Fprintf(b.buf, format+"\n", args...)
}

func (b *bufLogger) Criticalf(format string, args ...any) {
	fmt.Fprintf(b.buf, format+"\n", args...)
}

func TestEnvLog(t *testing.T) {
	tests := []struct {
		dbug bool
		name string
		str  string
	}{
		{name: "Info", str: "This is a log.Info test", dbug: false},
		{name: "Debug", str: "This is a log.Debug test", dbug: true},
		{name: "Error", str: "This is a log.Error test", dbug: true},
		{name: "Warning", str: "This is a log.Warning test", dbug: true},
		{name: "Not Debug", str: "This is a log.Debug test and there shouldn't be any output", dbug: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := &bytes.Buffer{}
			p := &Env{Log: &bufLogger{buf: act}, Verb: true}
			p.Dbug = tt.dbug

			switch {
			case tt.dbug:
				p.Debug(tt.str)
				if act.String() != tt.str+"\n" {
					t.Fatalf("got %q", act.String())
				}

			case tt.name == "Info":
				p.Log.Info(tt.str)
				if act.String() != tt.str+"\n" {
					t.Fatalf("got %q", act.String())
				}

			case tt.name == "Warning":
				p.Log.Warning(tt.str)
				if act.String() != tt.str+"\n" {
					t.Fatalf("got %q", act.String())
				}

			case tt.name == "Error":
				p.Log.Error(tt.str)
				if act.String() != tt.str+"\n" {
					t.Fatalf("got %q", act.String())
				}

			default:
				p.Debug(tt.str)
				if act.String() != "" {
					t.Fatalf("want empty, got %q", act.String())
				}
			}
		})
	}
}

func TestOption(t *testing.T) {
	vanilla := Env{ctr: ctr{RWMutex: &sync.RWMutex{}, stat: make(stat)}}
	exp := `{
	"API": "/bin/cli-shell-api",
	"Arch": "arm64",
	"Bash": "/bin/bash",
	"Cores": 2,
	"Disabled": false,
	"Dbug": true,
	"Dex": {},
	"Dir": "/tmp",
	"dnsmasq service": "service dnsmasq restart",
	"Exc": {},
	"dnsmasq fileExt.": "edgeos-adblock.conf",
	"File": "/config/config.boot",
	"File name fmt": "%v/%v.%v.%v",
	"HTTP method": "GET",
	"Prefix": {},
	"Test": true,
	"Timeout": 30000000000,
	"Wildcard": {
		"Node": "*s",
		"Name": "*"
	}
}`

	expRaw := Env{
		ctr:      ctr{RWMutex: &sync.RWMutex{}, stat: make(stat)},
		API:      "/bin/cli-shell-api",
		Arch:     "arm64",
		Bash:     "/bin/bash",
		Cores:    2,
		Disabled: false,
		Dbug:     true,
		Dex:      &list{entry: entry{}},
		Dir:      "/tmp",
		DNSsvc:   "service dnsmasq restart",
		Exc:      &list{entry: entry{}},
		Ext:      "edgeos-adblock.conf",
		File:     "/config/config.boot",
		FnFmt:    "%v/%v.%v.%v",
		InCLI:    "inSession",
		Method:   "GET",
		Pfx:      dnsPfx{domain: "address=", host: "server="},
		Test:     true,
		Timeout:  30000000000,
		Wildcard: Wildcard{Node: "*s", Name: "*"},
	}

	c := NewConfig()
	vanilla.Dex = c.Dex
	vanilla.Exc = c.Exc
	if !reflect.DeepEqual(c.Env, &vanilla) {
		t.Fatal("vanilla env mismatch")
	}

	c = NewConfig(
		Arch(runtime.GOARCH),
		API("/bin/cli-shell-api"),
		Bash("/bin/bash"),
		Cores(2),
		Dbug(true),
		Dir("/tmp"),
		DNSsvc("service dnsmasq restart"),
		Ext("edgeos-adblock.conf"),
		File("/config/config.boot"),
		FileNameFmt("%v/%v.%v.%v"),
		InCLI("inSession"),
		SetLogger(nil),
		Method("GET"),
		Prefix("address=", "server="),
		Test(true),
		Timeout(30*time.Second),
		Verb(false),
		WCard(Wildcard{Node: "*s", Name: "*"}),
	)

	expRaw.Dex.RWMutex = c.Dex.RWMutex
	expRaw.Exc.RWMutex = c.Exc.RWMutex

	if !reflect.DeepEqual(*c.Env, expRaw) {
		t.Fatal("expRaw mismatch")
	}
	if c.Env.String() != exp {
		t.Fatalf("JSON mismatch:\n%s", c.Env.String())
	}
}
