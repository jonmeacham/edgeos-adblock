package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	e "github.com/jonmeacham/edgeos-adblock/internal/edgeos"
)

var update = flag.Bool("update", false, "update .golden files")

func readGolden(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name+".golden") // relative path
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func writeGolden(t *testing.T, actual []byte, name string) error {
	golden := filepath.Join("testdata", name+".golden")
	if *update {
		return os.WriteFile(golden, actual, 0o644)
	}
	return nil
}

func (o *opts) String() string {
	var b strings.Builder
	o.VisitAll(func(f *flag.Flag) {
		if !o.visible[f.Name] {
			return
		}
		printFlagUsage(&b, f)
	})
	return b.String()
}

func TestLogFatalf(t *testing.T) {
	var (
		act string
		exp = "Something fatal happened!"
	)

	exitCmd = func(int) {}
	logFatalf = func(f string, args ...any) {
		act = fmt.Sprintf(f, args...)
	}

	logFatalf("%v", exp)
	if act != exp {
		t.Errorf("got %q want %q", act, exp)
	}
}

func TestMain(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("requires root (production parity)")
	}
	origArgs := os.Args
	prog := path.Base(os.Args[0])
	prfx := fmt.Sprintf("%s: ", prog)

	exitCmd = func(int) {}

	var act, actReloadDNS string
	logFatalf = func(f string, args ...any) {
		act = fmt.Sprintf(f, args...)
	}
	logPrintf = func(f string, vals ...any) {
		actReloadDNS = fmt.Sprintf(f, vals...)
	}

	screenLog(prfx)
	main()
	if act == "" {
		t.Error("expected logFatalf to run")
	}
	if actReloadDNS == "" {
		t.Error("expected reload log")
	}

	t.Run("config file load", func(t *testing.T) {
		act = ""
		os.Args = []string{prog, "-convey-json", "-f", "github.com/jonmeacham/edgeos-adblock/internal/testdata/config.erx.boot"}
		main()
		if act != "" {
			t.Errorf("act = %q, want empty", act)
		}
		os.Args = origArgs
	})

	t.Run("failed initEnv", func(t *testing.T) {
		buf := new(bytes.Buffer)
		initEnvirons = func() (env *e.Config, err error) {
			env, _ = initEnv()
			err = errors.New("initEnvirons failed")
			return env, err
		}
		os.Args = []string{prog, "-convey-json"}
		o := getOpts()
		o.Init("edgeos-adblock", flag.ContinueOnError)
		o.SetOutput(buf)
		main()
		if buf.String() != "" {
			t.Errorf("got %q", buf.String())
		}
		os.Args = origArgs
	})
}

func TestScreenLog(t *testing.T) {
	inTerminalHook = func() bool {
		return true
	}
	defer func() { inTerminalHook = nil }()

	screenLog("")
}

func TestExitCmd(t *testing.T) {
	var act int
	exitCmd = func(i int) {
		act = i
	}

	exitCmd(0)
	if act != 0 {
		t.Errorf("got %d want 0", act)
	}
}

func TestInitEnv(t *testing.T) {
	initEnv := func() (*e.Config, error) {
		return &e.Config{
			Env: &e.Env{Arch: "MegaOS"},
		}, nil
	}
	act, _ := initEnv()
	if act.Arch != "MegaOS" {
		t.Errorf("Arch %q", act.Arch)
	}

	origArgs := os.Args
	o := getOpts()
	o.setArgs()

	origBkpCfgFile := bkpCfgFile
	bkpCfgFile = "github.com/jonmeacham/edgeos-adblock/internal/testdata/config.test.boot"
	c := o.initEdgeOS()

	*o.ARCH = *o.MIPS64
	*o.Safe = true

	var err error
	c, err = loadConfig(c, o)
	if err != nil {
		t.Fatal(err)
	}
	if c == nil {
		t.Fatal("nil config")
	}

	bkpCfgFile = origBkpCfgFile
	os.Args = origArgs
}

func TestProcessObjects(t *testing.T) {
	c, _ := initEnv()
	badFileError := `open EinenSieAugenBlick/domains.tasty.edgeos-adblock.conf: no such file or directory`

	t.Run("loaded config string", func(t *testing.T) {
		if c.String() != mainGetConfig {
			t.Error("config string mismatch")
		}
		err := processObjects(c,
			[]e.IFace{
				e.ExRtObj,
				e.ExDmObj,
				e.ExHtObj,
			})
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Dex", func(t *testing.T) {
		if c.Dex.String() != expMap {
			t.Error("Dex mismatch")
		}
	})
	t.Run("Exc", func(t *testing.T) {
		if c.Exc.String() != expMap {
			t.Error("Exc mismatch")
		}
	})
	t.Run("bad iface", func(t *testing.T) {
		if processObjects(c, []e.IFace{100}) == nil {
			t.Error("want error")
		}
	})
	t.Run("bad dir", func(t *testing.T) {
		c.Dir = "EinenSieAugenBlick"
		err := processObjects(c, []e.IFace{e.FileObj})
		if err == nil || err.Error() != badFileError {
			t.Fatalf("got %v want %v", err, badFileError)
		}
	})
}

func TestSetArgs(t *testing.T) {
	var (
		origArgs = os.Args
		prog     = path.Base(os.Args[0])
	)

	exitCmd = func(int) {}
	defer func() { os.Args = origArgs }()

	tests := []struct {
		name string
		args []string
		exp  any
	}{
		{
			name: "h",
			args: []string{prog, "-convey-json", "-h"},
			exp:  true,
		},
		{
			name: "debug",
			args: []string{prog, "-debug"},
			exp:  true,
		},
		{
			name: "dryrun",
			args: []string{prog, "-dryrun"},
			exp:  true,
		},
		{
			name: "version",
			args: []string{prog, "-version"},
			exp:  true,
		},
		{
			name: "v",
			args: []string{prog, "-v"},
			exp:  true,
		},
		{
			name: "invalid flag",
			args: []string{prog, "-z"},
			exp:  readGolden(t, "testInvalidArgs"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args != nil {
				os.Args = tt.args
			}
			env := getOpts()
			env.Init(prog, flag.ContinueOnError)
			if tt.name == "invalid flag" {
				act := new(bytes.Buffer)
				env.SetOutput(act)
				env.setArgs()
				if err := writeGolden(t, act.Bytes(), "testInvalidArgs"); err != nil {
					t.Fatal(err)
				}
				want := tt.exp.([]byte)
				if !bytes.Equal(act.Bytes(), want) {
					t.Fatalf("invalid flag output mismatch")
				}
				return
			}
			env.setArgs()
			got := fmt.Sprint(env.Lookup(tt.name).Value.String())
			want := fmt.Sprint(tt.exp)
			if got != want {
				t.Errorf("Lookup(%s) got %q want %q", tt.name, got, want)
			}
		})
	}
}

func TestBasename(t *testing.T) {
	tests := []struct {
		s   string
		exp string
	}{
		{s: "e.txt", exp: "e"},
		{s: "/internal/edgeos", exp: "edgeos"},
	}

	for _, tt := range tests {
		if got := progName(tt.s); got != tt.exp {
			t.Errorf("%q: got %q want %q", tt.s, got, tt.exp)
		}
	}
}

func TestBuild(t *testing.T) {
	want := map[string]string{
		"build":   build,
		"githash": githash,
		"version": version,
	}

	for k := range want {
		if want[k] != "UNKNOWN" {
			t.Errorf("%s = %q", k, want[k])
		}
	}
}

func TestGetCFG(t *testing.T) {
	exitCmd = func(int) {}
	o := getOpts()
	c := o.initEdgeOS()

	_ = c.Blocklist(o.getCFG(c))
	if c.String() != mainGetConfig {
		t.Error("main config mismatch")
	}

	origBkpCfgFile := bkpCfgFile
	bkpCfgFile = "github.com/jonmeacham/edgeos-adblock/internal/testdata/config.test.boot"
	_ = c.Blocklist(o.getCFG(c))
	if c.String() != mainGetConfig {
		t.Error("after bkp path")
	}
	bkpCfgFile = origBkpCfgFile

	origFile := *o.File
	*o.File = "github.com/jonmeacham/edgeos-adblock/internal/testdata/config.test.boot"
	_ = c.Blocklist(o.getCFG(c))
	if c.String() != mainGetConfig {
		t.Error("with file")
	}
	*o.File = origFile

	*o.MIPS64 = "arm64"
	c = o.initEdgeOS()
	_ = c.Blocklist(o.getCFG(c))
	if c.String() != intelCfg {
		t.Error("intelCfg mismatch")
	}
}

func TestFiles(t *testing.T) {
	exp := ""
	env, _ := initEnv()
	act := files(env)
	if fmt.Sprintf("%v", act) != fmt.Sprintf("%v", exp) {
		t.Errorf("got %v want %v", act, exp)
	}
}

func TestReloadDNS(t *testing.T) {
	var (
		act string
		exp = "Successfully restarted dnsmasq"
	)

	c, _ := initEnv()
	exitCmd = func(int) {}
	logPrintf = func(s string, v ...any) {
		act = fmt.Sprintf(s, v...)
	}

	reloadDNS(c)
	if act != exp {
		t.Errorf("got %q want %q", act, exp)
	}
}

func TestRemoveStaleFiles(t *testing.T) {
	c, _ := initEnv()
	if err := removeStaleFiles(c); err != nil {
		t.Fatal(err)
	}
	_ = c.SetOpt(e.Dir("EinenSieAugenBlick"), e.Ext("[]a]"), e.FileNameFmt("[]a]"), e.WCard(e.Wildcard{Node: "[]a]", Name: "]"}))
	if removeStaleFiles(c) == nil {
		t.Error("want error")
	}
}

func TestSetArch(t *testing.T) {
	exitCmd = func(int) {}
	o := getOpts()

	tests := []struct {
		arch string
		exp  string
	}{
		{arch: "mips64", exp: "/etc/dnsmasq.d"},
		{arch: "linux", exp: "/tmp"},
		{arch: "darwin", exp: "/tmp"},
	}

	for _, test := range tests {
		if got := o.setDir(test.arch); got != test.exp {
			t.Errorf("%s: got %q want %q", test.arch, got, test.exp)
		}
	}
}

func TestSetLogFile(t *testing.T) {
	oldprog := prog
	prog = "update-dnsmasq"
	tests := []struct {
		os  string
		exp string
	}{
		{os: "darwin", exp: fmt.Sprintf("/tmp/%s.log", prog)},
		{os: "linux", exp: fmt.Sprintf("/var/log/%s.log", prog)},
	}

	for _, tt := range tests {
		if got := setLogFile(tt.os); got != tt.exp {
			t.Errorf("%s: got %q want %q", tt.os, got, tt.exp)
		}
	}
	prog = oldprog
}

func TestInitEdgeOS(t *testing.T) {
	exitCmd = func(int) {}
	o := getOpts()
	p := o.initEdgeOS()
	exp := fmt.Sprintf(`{
	"API": "/bin/cli-shell-api",
	"Arch": "%s",
	"Bash": "/bin/bash",
	"Cores": 2,
	"Disabled": false,
	"Dex": {},
	"Dir": "/tmp",
	"dnsmasq service": "/etc/init.d/dnsmasq restart",
	"Exc": {},
	"dnsmasq fileExt.": "edgeos-adblock.conf",
	"File name fmt": "%%v/%%v.%%v.%%v",
	"HTTP method": "GET",
	"Prefix": {},
	"Timeout": 30000000000,
	"Wildcard": {
		"Node": "*s",
		"Name": "*"
	}
}`, runtime.GOARCH)
	if fmt.Sprint(p.Env) != exp {
		t.Errorf("Env mismatch:\n%s", p.Env)
	}
}

var (
	mainGetConfig = `{
  "nodes": [
    {
      "blocklist": {
        "disabled": "false",
        "ip": "192.168.168.1",
        "excludes": [
          "1e100.net",
          "2o7.net",
          "adobedtm.com",
          "akamai.net",
          "akamaihd.net",
          "amazon.com",
          "amazonaws.com",
          "apple.com",
          "ask.com",
          "avast.com",
          "avira-update.com",
          "bannerbank.com",
          "bing.com",
          "bit.ly",
          "bitdefender.com",
          "cdn.ravenjs.com",
          "cdn.visiblemeasures.com",
          "cloudfront.net",
          "coremetrics.com",
          "dropbox.com",
          "ebay.com",
          "edgesuite.net",
          "evernote.com",
          "express.co.uk",
          "feedly.com",
          "freedns.afraid.org",
          "github.com",
          "githubusercontent.com",
          "global.ssl.fastly.net",
          "google.com",
          "googleads.g.doubleclick.net",
          "googleadservices.com",
          "googleapis.com",
          "googletagmanager.com",
          "googleusercontent.com",
          "gstatic.com",
          "gvt1.com",
          "gvt1.net",
          "hb.disney.go.com",
          "herokuapp.com",
          "hp.com",
          "hulu.com",
          "images-amazon.com",
          "live.com",
          "magnetmail1.net",
          "microsoft.com",
          "microsoftonline.com",
          "msdn.com",
          "msecnd.net",
          "msftncsi.com",
          "mywot.com",
          "nsatc.net",
          "paypal.com",
          "pop.h-cdn.co",
          "rackcdn.com",
          "rarlab.com",
          "schema.org",
          "shopify.com",
          "skype.com",
          "smacargo.com",
          "sourceforge.net",
          "spotify.com",
          "spotify.edgekey.net",
          "spotilocal.com",
          "ssl-on9.com",
          "ssl-on9.net",
          "sstatic.net",
          "static.chartbeat.com",
          "storage.googleapis.com",
          "twimg.com",
          "viewpoint.com",
          "windows.net",
          "xboxlive.com",
          "yimg.com",
          "ytimg.com"
        ],
        "includes": [],
        "sources": [
          {}
        ]
      },
      "domains": {
        "disabled": "false",
        "excludes": [],
        "includes": [],
        "sources": [
          {
            "tasty": {
              "disabled": "false",
              "description": "File source",
              "ip": "10.10.10.10",
              "file": "../../internal/testdata/blist.hosts.src"
            }
          }
        ]
      },
      "hosts": {
        "disabled": "false",
        "excludes": [],
        "includes": [
          "beap.gemini.yahoo.com"
        ],
        "sources": [
          {
            "hageziPro": {
              "disabled": "false",
              "description": "HaGeZi DNS Blocklists — Pro (dnsmasq)",
              "url": "https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt"
            }
          }
        ]
      }
    }
  ]
}`

	expMap = `"1e100.net":{},
"2o7.net":{},
"adobedtm.com":{},
"akamai.net":{},
"akamaihd.net":{},
"amazon.com":{},
"amazonaws.com":{},
"apple.com":{},
"ask.com":{},
"avast.com":{},
"avira-update.com":{},
"bannerbank.com":{},
"bing.com":{},
"bit.ly":{},
"bitdefender.com":{},
"cdn.ravenjs.com":{},
"cdn.visiblemeasures.com":{},
"cloudfront.net":{},
"coremetrics.com":{},
"dropbox.com":{},
"ebay.com":{},
"edgesuite.net":{},
"evernote.com":{},
"express.co.uk":{},
"feedly.com":{},
"freedns.afraid.org":{},
"github.com":{},
"githubusercontent.com":{},
"global.ssl.fastly.net":{},
"google.com":{},
"googleads.g.doubleclick.net":{},
"googleadservices.com":{},
"googleapis.com":{},
"googletagmanager.com":{},
"googleusercontent.com":{},
"gstatic.com":{},
"gvt1.com":{},
"gvt1.net":{},
"hb.disney.go.com":{},
"herokuapp.com":{},
"hp.com":{},
"hulu.com":{},
"images-amazon.com":{},
"live.com":{},
"magnetmail1.net":{},
"microsoft.com":{},
"microsoftonline.com":{},
"msdn.com":{},
"msecnd.net":{},
"msftncsi.com":{},
"mywot.com":{},
"nsatc.net":{},
"paypal.com":{},
"pop.h-cdn.co":{},
"rackcdn.com":{},
"rarlab.com":{},
"schema.org":{},
"shopify.com":{},
"skype.com":{},
"smacargo.com":{},
"sourceforge.net":{},
"spotify.com":{},
"spotify.edgekey.net":{},
"spotilocal.com":{},
"ssl-on9.com":{},
"ssl-on9.net":{},
"sstatic.net":{},
"static.chartbeat.com":{},
"storage.googleapis.com":{},
"twimg.com":{},
"viewpoint.com":{},
"windows.net":{},
"xboxlive.com":{},
"yimg.com":{},
"ytimg.com":{},
`
	intelCfg = `{
  "nodes": [
    {}
  ]
}`
)
