package edgeos

import (
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestAdblockParsesConfiguration(t *testing.T) {
	c := NewConfig()
	if err := c.Adblock(&CFGstatic{Cfg: fixtureConfig(t)}); err != nil {
		t.Fatal(err)
	}
	if got, want := c.Nodes(), []string{"adblock"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("nodes = %#v, want %#v", got, want)
	}
	if got, want := c.redirectIP(), "192.168.168.1"; got != want {
		t.Fatalf("root redirect = %q, want %q", got, want)
	}
	if got := c.GetAll(urlNode).Len(); got != 1 {
		t.Fatalf("URL sources = %d, want 1", got)
	}
	if got := c.GetAll(fileNode).Len(); got != 1 {
		t.Fatalf("file sources = %d, want 1", got)
	}
	if got, want := c.GetAll(fileNode).src[0].ip, "10.10.10.10"; got != want {
		t.Fatalf("file source redirect = %q, want %q", got, want)
	}
	if got, want := c.GetAll(urlNode).src[0].ip, "192.168.168.1"; got != want {
		t.Fatalf("URL source redirect = %q, want inherited %q", got, want)
	}
}

func TestAdblockHandlesDisabledAndMissingConfiguration(t *testing.T) {
	c := NewConfig()
	if err := c.Adblock(&CFGstatic{Cfg: fixtureDisabled}); err != nil {
		t.Fatal(err)
	}
	if !c.Disabled {
		t.Fatal("disabled configuration did not disable processing")
	}
	if got := c.Files().Strings(); len(got) != 0 {
		t.Fatalf("disabled configuration expects generated files: %#v", got)
	}

	for _, cfg := range []string{"", fixtureNoAdblock} {
		err := NewConfig().Adblock(&CFGstatic{Cfg: cfg})
		if !errors.Is(err, ErrNoAdblockCfg) {
			t.Fatalf("config %q: got %v, want ErrNoAdblockCfg", cfg, err)
		}
	}
}

func TestNewContent(t *testing.T) {
	c := NewConfig()
	if err := c.Adblock(&CFGstatic{Cfg: fixtureConfig(t)}); err != nil {
		t.Fatal(err)
	}
	for _, iface := range []IFace{AllowObj, BlockObj, FileObj, URLObj} {
		if _, err := c.NewContent(iface); err != nil {
			t.Errorf("NewContent(%v): %v", iface, err)
		}
	}
	if _, err := c.NewContent(Invalid); err == nil {
		t.Fatal("NewContent(Invalid) succeeded")
	}
}

func TestReloadDNS(t *testing.T) {
	out, err := NewConfig(Bash("/bin/bash"), DNSsvc("true")).ReloadDNS()
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("output = %q, want empty", out)
	}
}

func TestCFileRemoveCoversAllGeneratedClasses(t *testing.T) {
	dir := t.TempDir()
	stale := []string{
		dir + "/source.old.edgeos-adblock.conf",
		dir + "/block.explicit.edgeos-adblock.conf",
		dir + "/legacy.old.edgeos-adblock.conf",
	}
	for _, name := range stale {
		if err := os.WriteFile(name, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	c := &CFile{Env: NewConfig(
		Dir(dir),
		Ext("edgeos-adblock.conf"),
	).Env}
	if err := c.Remove(); err != nil {
		t.Fatal(err)
	}
	for _, name := range stale {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Errorf("stale file still exists: %s", name)
		}
	}
}

func TestCFileRemovePreservesExpectedAndRemovesRenamedSource(t *testing.T) {
	dir := t.TempDir()
	expected := dir + "/source.current.edgeos-adblock.conf"
	stale := dir + "/source.old.edgeos-adblock.conf"
	for _, name := range []string{expected, stale} {
		if err := os.WriteFile(name, nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	c := &CFile{
		Env:   NewConfig(Dir(dir), Ext("edgeos-adblock.conf")).Env,
		Names: []string{expected},
	}
	if err := c.Remove(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected file was removed: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale file still exists: %v", err)
	}
}

func TestAdblockRejectsInvalidSources(t *testing.T) {
	for _, cfg := range []string{
		"adblock {\nsource missing {\ndescription missing\n}\n}",
		"adblock {\nsource both {\nurl https://example.invalid/list\nfile /tmp/list\n}\n}",
		"adblock {\nsource same {\nurl https://example.invalid/one\n}\nsource same {\nurl https://example.invalid/two\n}\n}",
		"adblock {\nsource bad/name {\nurl https://example.invalid/one\n}\n}",
		"adblock {\nsource badurl {\nurl ftp://example.invalid/one\n}\n}",
		"adblock {\nhosts {\n}\n}",
		"adblock {\nexclude old.example\n}",
		"adblock {\nsource legacy {\nkind hosts\nurl https://example.invalid/one\n}\n}",
	} {
		if err := NewConfig().Adblock(&CFGstatic{Cfg: cfg}); err == nil {
			t.Fatalf("invalid configuration succeeded:\n%s", cfg)
		}
	}
}

func TestBoolConversions(t *testing.T) {
	for _, value := range []bool{true, false} {
		s := booltoStr(value)
		got, err := strToBool(s)
		if err != nil || got != value {
			t.Fatalf("round trip %v: got %v, %v", value, got, err)
		}
	}
}
