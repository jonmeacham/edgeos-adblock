package edgeos

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/jonmeacham/edgeos-adblock/internal/tdata"
)

func TestAddInc(t *testing.T) {
	var (
		c   = NewConfig()
		err = c.Blacklist(&CFGstatic{Cfg: tdata.Cfg})
	)

	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		exp  *source
		node string
	}{
		{
			name: "rootNode",
			node: rootNode,
			exp: &source{
				Env: &Env{
					Wildcard: Wildcard{
						Node: "",
						Name: "",
					},
					ctr:   ctr{RWMutex: &sync.RWMutex{}, stat: make(stat)},
					API:   "",
					Arch:  "",
					Bash:  "",
					Cores: 0,
					Dbug:  false,
					Dex: &list{
						RWMutex: &sync.RWMutex{},
						entry:   entry{},
					},
					Dir:    "",
					DNSsvc: "",
					Exc: &list{
						RWMutex: &sync.RWMutex{},
						entry:   entry{},
					},
					Ext:   "",
					File:  "",
					FnFmt: "",
					InCLI: "",
					// ioWriter: nil,
					Method:  "",
					Pfx:     dnsPfx{domain: "", host: ""},
					Test:    false,
					Timeout: time.Duration(0),
					Verb:    false,
				},
				desc:     "pre-configured global blacklisted domains",
				disabled: false,
				err:      nil,
				exc:      nil,
				file:     "",
				inc:      []string{},
				iface:    PreRObj,
				ip:       "0.0.0.0",
				ltype:    "global-blacklisted-domains",
				name:     "global-blacklisted-domains",
				nType:    ntype(8),
				Objects: Objects{
					Env: nil,
					src: nil,
				},
				prefix: "",
				r:      nil,
				url:    "",
			},
		},
		{
			name: "domains",
			node: domains,
			exp: &source{
				Env: &Env{
					Wildcard: Wildcard{
						Node: "",
						Name: "",
					},
					ctr:   ctr{RWMutex: &sync.RWMutex{}, stat: make(stat)},
					API:   "",
					Arch:  "",
					Bash:  "",
					Cores: 0,
					Dbug:  false,
					Dex: &list{
						RWMutex: &sync.RWMutex{},
						entry:   entry{},
					},
					Dir:    "",
					DNSsvc: "",
					Exc: &list{
						RWMutex: &sync.RWMutex{},
						entry:   entry{},
					},
					Ext:   "",
					File:  "",
					FnFmt: "",
					InCLI: "",
					// ioWriter: nil,
					Method:  "",
					Pfx:     dnsPfx{domain: "", host: ""},
					Test:    false,
					Timeout: time.Duration(0),
					Verb:    false,
				},
				desc:     "pre-configured blacklisted subdomains",
				disabled: false,
				err:      nil,
				exc:      nil,
				file:     "",
				inc:      []string{},
				iface:    PreDObj,
				ip:       "192.168.100.1",
				ltype:    "blacklisted-subdomains",
				name:     "blacklisted-subdomains",
				nType:    ntype(6),
				Objects: Objects{
					Env: nil,
					src: nil,
				},
				prefix: "",
				r:      nil,
				url:    "",
			},
		},
		{
			name: "hosts",
			node: hosts,
			exp: &source{
				Env: &Env{
					Wildcard: Wildcard{
						Node: "",
						Name: "",
					},
					ctr:   ctr{RWMutex: &sync.RWMutex{}, stat: make(stat)},
					API:   "",
					Arch:  "",
					Bash:  "",
					Cores: 0,
					Dbug:  false,
					Dex: &list{
						RWMutex: &sync.RWMutex{},
						entry:   entry{},
					},
					Dir:    "",
					DNSsvc: "",
					Exc: &list{
						RWMutex: &sync.RWMutex{},
						entry:   entry{},
					},
					Ext:   "",
					File:  "",
					FnFmt: "",
					InCLI: "",
					// ioWriter: nil,
					Method:  "",
					Pfx:     dnsPfx{domain: "", host: ""},
					Test:    false,
					Timeout: time.Duration(0),
					Verb:    false,
				},
				desc:     "pre-configured blacklisted servers",
				disabled: false,
				err:      nil,
				exc:      nil,
				file:     "",
				iface:    PreHObj,
				inc:      []string{"beap.gemini.yahoo.com"},
				ip:       "0.0.0.0",
				ltype:    "blacklisted-servers",
				name:     "blacklisted-servers",
				nType:    ntype(7),
				Objects: Objects{
					Env: nil,
					src: nil,
				},
				prefix: "",
				r:      nil,
				url:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inc := c.addInc(tt.node)

			if !reflect.DeepEqual(inc, tt.exp) {
				t.Errorf("addInc(%q): mismatch", tt.node)
			}
		})
	}
}

func TestGetIP(t *testing.T) {
	b := tree{}
	t.Run("badnode", func(t *testing.T) {
		if got, want := b.getIP("badnode"), "0.0.0.0"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	b = tree{
		rootNode: &source{
			ip: "192.168.1.50",
		},
		domains: &source{
			ip: "192.168.1.20",
		},
		hosts: &source{
			ip: "192.168.1.30",
		},
	}
	t.Run(rootNode, func(t *testing.T) {
		if got, want := b.getIP(rootNode), "192.168.1.50"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run(domains, func(t *testing.T) {
		if got, want := b.getIP(domains), "192.168.1.20"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run(hosts, func(t *testing.T) {
		if got, want := b.getIP(hosts), "192.168.1.30"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestFiles(t *testing.T) {
	r := &CFGstatic{Cfg: tdata.Cfg}
	c := NewConfig(
		Dir("/tmp"),
		Ext("edgeos-adblock.conf"),
	)

	if err := c.Blacklist(r); err != nil {
		t.Fatal(err)
	}

	exp := `/tmp/domains.blacklisted-subdomains.edgeos-adblock.conf
/tmp/domains.tasty.edgeos-adblock.conf
/tmp/hosts.blacklisted-servers.edgeos-adblock.conf
/tmp/hosts.hageziPro.edgeos-adblock.conf
/tmp/roots.global-blacklisted-domains.edgeos-adblock.conf`

	act := c.GetAll().Files().String()
	if act != exp {
		t.Errorf("Files.String(): mismatch\n got:\n%s\n want:\n%s", act, exp)
	}
}

func TestInSession(t *testing.T) {
	c := NewConfig()
	if c.InSession() {
		t.Error("InSession: expected false initially")
	}

	if err := os.Setenv("_OFR_CONFIGURE", "ok"); err != nil {
		t.Fatal(err)
	}
	if !c.InSession() {
		t.Error("InSession: expected true after setenv")
	}

	if err := os.Unsetenv("_OFR_CONFIGURE"); err != nil {
		t.Fatal(err)
	}
	if c.InSession() {
		t.Error("InSession: expected false after unsetenv")
	}
}

func TestIsSource(t *testing.T) {
	var node []string
	if !isntSource(node) {
		t.Error("isntSource(nil): expected true")
	}
}

func TestNodeExists(t *testing.T) {
	var (
		c   = NewConfig()
		err = c.Blacklist(&CFGstatic{Cfg: tdata.Cfg})
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.nodeExists("broken") {
		t.Error("nodeExists(broken): expected false")
	}
}

func TestReadCfg(t *testing.T) {
	var (
		err error
		b   []byte
		f   = "../testdata/config.erx.boot"
		r   io.Reader
	)

	if r, err = GetFile(f); err != nil {
		t.Fatalf("cannot open configuration file %s: %v", f, err)
	}

	b, err = io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("configuration loaded from file", func(t *testing.T) {
		act := NewConfig().Blacklist(&CFGstatic{Cfg: string(b)})
		if act != nil {
			t.Fatalf("Blacklist: got %v, want nil", act)
		}
	})

	t.Run("empty configuration", func(t *testing.T) {
		act := NewConfig().Blacklist(&CFGstatic{Cfg: ""})
		if !errors.Is(act, ErrNoBlacklistCfg) {
			t.Fatalf("got %v, want ErrNoBlacklistCfg", act)
		}
		if act.Error() != "no EdgeOS dns forwarding blacklist configuration has been detected" {
			t.Errorf("wrong error message: %q", act.Error())
		}
	})
	t.Run("disabled configuration", func(t *testing.T) {
		act := NewConfig().Blacklist(&CFGstatic{Cfg: tdata.DisabledCfg})
		if act != nil {
			t.Fatalf("got %v, want nil", act)
		}
	})

	t.Run("single source configuration", func(t *testing.T) {
		act := NewConfig().Blacklist(&CFGstatic{Cfg: tdata.SingleSource})
		if act != nil {
			t.Fatalf("got %v, want nil", act)
		}
	})

	t.Run("active configuration", func(t *testing.T) {
		c := NewConfig()
		if err := c.Blacklist(&CFGstatic{Cfg: tdata.Cfg}); err != nil {
			t.Fatal(err)
		}
		want := []string{"blacklist", "domains", "hosts"}
		if !reflect.DeepEqual(c.Nodes(), want) {
			t.Errorf("Nodes(): got %#v, want %#v", c.Nodes(), want)
		}
	})
}

func TestReadUnconfiguredCfg(t *testing.T) {
	act := NewConfig().Blacklist(&CFGstatic{Cfg: tdata.NoBlacklist})
	if !errors.Is(act, ErrNoBlacklistCfg) {
		t.Fatalf("got %v, want ErrNoBlacklistCfg", act)
	}
	if act.Error() != "no EdgeOS dns forwarding blacklist configuration has been detected" {
		t.Errorf("wrong error message: %q", act.Error())
	}
}

func TestReloadDNS(t *testing.T) {
	act, err := NewConfig(Bash("/bin/bash"), DNSsvc("true")).ReloadDNS()
	if err != nil {
		t.Fatal(err)
	}
	if string(act) != "" {
		t.Errorf("ReloadDNS output: got %q, want empty", string(act))
	}
}

func TestRemove(t *testing.T) {
	dir, err := os.MkdirTemp("", "edgeos-adblock-config-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	c := NewConfig(
		Dir(dir),
		Ext("edgeos-adblock.conf"),
		FileNameFmt("%v/%v.%v.%v"),
		WCard(Wildcard{Node: "*s", Name: "*"}),
	)

	if err := c.Blacklist(&CFGstatic{Cfg: tdata.CfgMimimal}); err != nil {
		t.Fatal(err)
	}

	t.Run("creating special case file", func(t *testing.T) {
		f, err := os.Create(fmt.Sprintf("%v/hosts.raw.github.com.edgeos-adblock.conf", dir))
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	})

	for _, node := range c.sortKeys() {
		for i := range Iter(10) {
			fname := fmt.Sprintf("%v/%v.%v.%v", dir, node, i, c.Ext)
			f, err := os.Create(fname)
			if err != nil {
				t.Fatal(err)
			}
			f.Close()
		}
	}

	for _, fname := range c.GetAll().Files().Strings() {
		f, err := os.Create(fname)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	if err := c.GetAll().Files().Remove(); err != nil {
		t.Fatal(err)
	}

	cf := &CFile{Env: c.Env}
	pattern := fmt.Sprintf(c.FnFmt, c.Dir, "*s", "*", c.Env.Ext)
	act, err := cf.readDir(pattern)

	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(act, c.GetAll().Files().Strings()) {
		t.Errorf("readDir: got %#v, want %#v", act, c.GetAll().Files().Strings())
	}

	prev := c.SetOpt(WCard(Wildcard{Node: "[]a]", Name: "]"}))

	if err := cf.Remove(); err == nil {
		t.Fatal("Remove: expected error")
	}
	c.SetOpt(prev)
}

func TestBooltoString(t *testing.T) {
	if got, want := booltoStr(true), True; got != want {
		t.Errorf("booltoStr(true): got %q, want %q", got, want)
	}
	if got, want := booltoStr(false), False; got != want {
		t.Errorf("booltoStr(false): got %q, want %q", got, want)
	}
}

func TestToBool(t *testing.T) {
	b, err := strToBool(True)
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Error("strToBool(True): expected true")
	}
	b, err = strToBool(False)
	if err != nil {
		t.Fatal(err)
	}
	if b {
		t.Error("strToBool(False): expected false")
	}
}

func TestGetAll(t *testing.T) {
	c := NewConfig(
		Dir("/tmp"),
		Ext(".edgeos-adblock.conf"),
	)

	if err := c.Blacklist(&CFGstatic{Cfg: tdata.Cfg}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		exp   string
		ltype string
		name  string
	}{
		{name: "GetAll()", ltype: "", exp: expGetAll},
		{name: "GetAll(url)", ltype: urls, exp: expURLS},
		{name: "GetAll(files)", ltype: files, exp: expFiles},
		{name: "GetAll(PreDomns, PreHosts)", ltype: PreDomns, exp: expPre},
		{name: "Get(all).String()", ltype: all, exp: c.Get(all).String()},
		{name: "c.Get(hosts)", ltype: hosts, exp: expHostObj},
		{name: "c.Get(domains)", ltype: domains, exp: expDomainObj},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.ltype {
			case "":
				got = c.GetAll().String()
			case all:
				got = c.GetAll().String()
			case domains:
				got = c.Get(domains).String()
			case hosts:
				got = c.Get(hosts).String()
			case PreDomns:
				got = c.GetAll(PreDomns, PreHosts).String()
			default:
				got = c.GetAll(tt.ltype).String()
			}
			if got != tt.exp {
				t.Errorf("got mismatch for ltype %q", tt.ltype)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	b := make(tree)
	if got, want := b.validate("borked").String(), ""; got != want {
		t.Errorf("validate: got %q, want %q", got, want)
	}
}

var (
	expDomainObj = `
Desc:         "pre-configured blacklisted subdomains"
Disabled:     "false"
File:         "**Undefined**"
IP:           "192.168.100.1"
Ltype:        "blacklisted-subdomains"
Name:         "blacklisted-subdomains"
nType:        "preDomn"
Prefix:       "**Undefined**"
Type:         "blacklisted-subdomains"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"

Desc:         "File source"
Disabled:     "false"
File:         "../../internal/testdata/blist.hosts.src"
IP:           "10.10.10.10"
Ltype:        "file"
Name:         "tasty"
nType:        "domn"
Prefix:       "**Undefined**"
Type:         "domains"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"
`

	expFiles = `
Desc:         "File source"
Disabled:     "false"
File:         "../../internal/testdata/blist.hosts.src"
IP:           "10.10.10.10"
Ltype:        "file"
Name:         "tasty"
nType:        "domn"
Prefix:       "**Undefined**"
Type:         "domains"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"
`

	expGetAll = `
Desc:         "pre-configured global blacklisted domains"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "global-blacklisted-domains"
Name:         "global-blacklisted-domains"
nType:        "preRoot"
Prefix:       "**Undefined**"
Type:         "global-blacklisted-domains"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"

Desc:         "pre-configured blacklisted subdomains"
Disabled:     "false"
File:         "**Undefined**"
IP:           "192.168.100.1"
Ltype:        "blacklisted-subdomains"
Name:         "blacklisted-subdomains"
nType:        "preDomn"
Prefix:       "**Undefined**"
Type:         "blacklisted-subdomains"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"

Desc:         "File source"
Disabled:     "false"
File:         "../../internal/testdata/blist.hosts.src"
IP:           "10.10.10.10"
Ltype:        "file"
Name:         "tasty"
nType:        "domn"
Prefix:       "**Undefined**"
Type:         "domains"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"

Desc:         "pre-configured blacklisted servers"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "blacklisted-servers"
Name:         "blacklisted-servers"
nType:        "preHost"
Prefix:       "**Undefined**"
Type:         "blacklisted-servers"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "beap.gemini.yahoo.com"

Desc:         "HaGeZi DNS Blocklists — Pro (dnsmasq)"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "url"
Name:         "hageziPro"
nType:        "host"
Prefix:       "**Undefined**"
Type:         "hosts"
URL:          "https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"
`

	expHostObj = `
Desc:         "pre-configured blacklisted servers"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "blacklisted-servers"
Name:         "blacklisted-servers"
nType:        "preHost"
Prefix:       "**Undefined**"
Type:         "blacklisted-servers"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "beap.gemini.yahoo.com"

Desc:         "HaGeZi DNS Blocklists — Pro (dnsmasq)"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "url"
Name:         "hageziPro"
nType:        "host"
Prefix:       "**Undefined**"
Type:         "hosts"
URL:          "https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"
`

	expPre = `
Desc:         "pre-configured blacklisted subdomains"
Disabled:     "false"
File:         "**Undefined**"
IP:           "192.168.100.1"
Ltype:        "blacklisted-subdomains"
Name:         "blacklisted-subdomains"
nType:        "preDomn"
Prefix:       "**Undefined**"
Type:         "blacklisted-subdomains"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"

Desc:         "pre-configured blacklisted servers"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "blacklisted-servers"
Name:         "blacklisted-servers"
nType:        "preHost"
Prefix:       "**Undefined**"
Type:         "blacklisted-servers"
URL:          "**Undefined**"
Whitelist:
              "**No entries found**"
Blacklist:
              "beap.gemini.yahoo.com"
`

	expURLS = `
Desc:         "HaGeZi DNS Blocklists — Pro (dnsmasq)"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "url"
Name:         "hageziPro"
nType:        "host"
Prefix:       "**Undefined**"
Type:         "hosts"
URL:          "https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt"
Whitelist:
              "**No entries found**"
Blacklist:
              "**No entries found**"
`
)
