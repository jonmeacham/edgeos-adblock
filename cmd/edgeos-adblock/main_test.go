package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	e "github.com/jonmeacham/edgeos-adblock/internal/edgeos"
)

func TestProgName(t *testing.T) {
	for _, tt := range []struct {
		in, want string
	}{
		{"/usr/local/bin/edgeos-adblock", "edgeos-adblock"},
		{"edgeos-adblock.exe", "edgeos-adblock"},
	} {
		if got := progName(tt.in); got != tt.want {
			t.Errorf("progName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestVisibleOptions(t *testing.T) {
	o := getOpts()
	var out bytes.Buffer
	o.SetOutput(&out)
	o.Usage()
	help := out.String()
	for _, flagName := range []string{"-dir", "-f", "-h", "-v", "-version"} {
		if !strings.Contains(help, flagName) {
			t.Errorf("help does not contain %s", flagName)
		}
	}
	for _, removed := range []string{"-safe", "-dryrun", "-arch", "-tmp"} {
		if strings.Contains(help, removed) {
			t.Errorf("help unexpectedly contains %s", removed)
		}
	}
}

func TestGetCFG(t *testing.T) {
	o := getOpts()
	c := e.NewConfig()
	if _, ok := o.getCFG(c).(*e.CFGcli); !ok {
		t.Fatal("empty -f should use the live EdgeOS CLI")
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "config.boot")
	const cfg = "adblock { disabled false }"
	if err := os.WriteFile(file, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	*o.File = file
	loader, ok := o.getCFG(c).(*e.CFGstatic)
	if !ok || loader.Cfg != cfg {
		t.Fatalf("file loader = %#v", loader)
	}
}

func TestCleanArgs(t *testing.T) {
	got := cleanArgs([]string{"-test.v", "-convey-json", "-v", "-f", "config.boot"})
	want := []string{"-v", "-f", "config.boot"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPrintFlagUsage(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("f", "", "load `<file>`")
	var out bytes.Buffer
	printFlagUsage(&out, fs.Lookup("f"))
	if !strings.Contains(out.String(), "-f <file>") {
		t.Fatalf("usage = %q", out.String())
	}
}
