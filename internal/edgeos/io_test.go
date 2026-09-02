package edgeos

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestLoad(t *testing.T) {
	c := NewConfig(
		API("/bin/cli-shell-api"),
		Bash("/bin/bash"),
		InCLI("inSession"),
	)

	_, err := c.load("zBroken")
	if err == nil {
		t.Fatal("load zBroken: expected error")
	}

	_, err = c.load("showConfig")
	if err == nil {
		t.Fatal("load showConfig: expected error")
	}

	r := CFGcli{Config: c}
	act, err := io.ReadAll(r.read())
	if err != nil {
		t.Fatal(err)
	}
	if string(act) != "" {
		t.Errorf("ReadAll: got %q, want empty", string(act))
	}

	cfg, err := c.load("echo")
	if err == nil {
		t.Fatal("load echo: expected error")
	}
	if !reflect.DeepEqual(cfg, []byte{}) {
		t.Errorf("load echo cfg: got %#v, want []byte{}", cfg)
	}
}

func TestPurgeFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "edgeos-adblock-purge-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	var (
		ext       = ".delete"
		purgeList []string
	)

	for i := range Iter(10) {
		f, err := os.CreateTemp(dir, fmt.Sprintf("%03d-*", i)+ext)
		if err != nil {
			t.Fatal(err)
		}
		purgeList = append(purgeList, f.Name())
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	if err := purgeFiles(purgeList); err != nil {
		t.Fatal(err)
	}
	// As root (e.g. default golang Docker), unlink("/dev/null") may succeed; only assert when non-root.
	if os.Geteuid() != 0 {
		if err := purgeFiles([]string{"/dev/null"}); err == nil {
			t.Fatal("purgeFiles(/dev/null): expected error")
		}
	}
	if err := purgeFiles([]string{"SpiegelAdlerIstHier"}); err != nil {
		t.Fatal(err)
	}
}

func TestAPICMD(t *testing.T) {
	tests := []struct {
		b    bool
		q, r string
	}{
		{
			b: false,
			q: "listNodes",
			r: "listNodes",
		},
		{
			b: true,
			q: "listNodes",
			r: "listActiveNodes",
		},
		{
			b: false,
			q: "listActiveNodes",
			r: "listNodes",
		},
		{
			b: false,
			q: "returnValue",
			r: "returnValue",
		},
		{
			b: true,
			q: "returnValue",
			r: "returnActiveValue",
		},
		{
			b: false,
			q: "returnActiveValue",
			r: "returnValue",
		},
		{
			b: false,
			q: "returnValues",
			r: "returnValues",
		},
		{
			b: true,
			q: "returnValues",
			r: "returnActiveValues",
		},
		{
			b: false,
			q: "returnActiveValues",
			r: "returnValues",
		},
		{
			b: false,
			q: "exists",
			r: "exists",
		},
		{
			b: true,
			q: "exists",
			r: "existsActive",
		},
		{
			b: false,
			q: "existsActive",
			r: "exists",
		},
		{
			b: false,
			q: "showCfg",
			r: "showCfg",
		},
		{
			b: true,
			q: "showCfg",
			r: "showConfig",
		},
		{
			b: false,
			q: "showConfig",
			r: "showCfg",
		},
	}

	for _, tt := range tests {
		if got, want := apiCMD(tt.q, tt.b), tt.r; got != want {
			t.Errorf("apiCMD(%q, %v): got %q, want %q", tt.q, tt.b, got, want)
		}
	}
}

func TestDeleteFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "edgeos-adblock-io-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	ext := "delete.me"

	tests := []struct {
		name string
		f    string
		exp  bool
	}{
		{
			name: "exists",
			f:    fmt.Sprintf("%v%v", "goodFile", ext),
			exp:  true,
		},
		{
			name: "non-existent",
			f:    fmt.Sprintf("%v%v", "badFile", ext),
			exp:  false,
		},
	}

	for _, tt := range tests {
		switch tt.name {
		case "exists":
			f, err := os.CreateTemp(dir, tt.f+"*")
			if err != nil {
				t.Fatal(err)
			}
			if got, want := deleteFile(f.Name()), tt.exp; got != want {
				t.Errorf("deleteFile: got %v, want %v", got, want)
			}
		default:
			if got, want := deleteFile(tt.f), tt.exp; got != want {
				t.Errorf("deleteFile: got %v, want %v", got, want)
			}
		}
	}
}
