package edgeos

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

type dummyConfig struct {
	*Env
	s []string
	t *testing.T
}

func (d *dummyConfig) ProcessContent(cts ...Contenter) error {
	var (
		a, b  int32
		area  string
		tally = &stats{dropped: a, kept: b}
	)

	for _, ct := range cts {
		o := ct.GetList().src
		for _, src := range o {
			area = typeInt(src.nType)
			src.ctr.Lock()
			src.ctr.stat[area] = tally
			src.ctr.Unlock()
			b, _ := io.ReadAll(src.process().r)
			d.s = append(d.s, strings.TrimSuffix(string(b), "\n"))
		}
	}
	return nil
}

func TestConfigProcessContent(t *testing.T) {
	newCfg := func() *Config {
		return NewConfig(
			API("/bin/cli-shell-api"),
			Arch(runtime.GOARCH),
			Bash("/bin/bash"),
			Cores(runtime.NumCPU()),
			Dir("/tmp"),
			DNSsvc("service dnsmasq restart"),
			Ext("blocklist.conf"),
			FileNameFmt("%v/%v.%v.%v"),
			InCLI("inSession"),
			SetLogger(newLog()),
			Method("GET"),
			Prefix("address=", "server="),
			Timeout(30*time.Second),
			WCard(Wildcard{Node: "*s", Name: "*"}),
		)
	}

	tests := []struct {
		c      *Config
		cfg    string
		ct     IFace
		err    error
		expErr bool
		name   string
	}{
		{
			c:      newCfg(),
			cfg:    testallCfg,
			ct:     URLhObj,
			err:    errors.New("Get \"http://127.0.0.1:8081/hosts/host.txt\": dial tcp 127.0.0.1:8081: connect: connection refused"),
			expErr: true,
			name:   "Hosts blocklist source",
		},
		{
			c:      newCfg(),
			cfg:    testallCfg,
			ct:     URLdObj,
			err:    errors.New("Get \"http://127.0.0.1:8081/domains/domain.txt\": dial tcp 127.0.0.1:8081: connect: connection refused"),
			expErr: true,
			name:   "Domains blocklist source",
		},
		{
			c:      newCfg(),
			cfg:    testallCfg,
			ct:     FileObj,
			err:    errors.New("open /:~//hosts.tasty.blocklist.conf: no such file or directory"),
			expErr: true,
			name:   "File source",
		},
		{
			c:      newCfg(),
			cfg:    testallCfg,
			ct:     ExHtObj,
			err:    errors.New("open /:~//hosts.allowlisted-servers.blocklist.conf: no such file or directory"),
			expErr: true,
			name:   "Allowlisted hosts",
		},
		{
			c:      newCfg(),
			cfg:    testallCfg,
			ct:     ExDmObj,
			err:    errors.New("open /:~//domains.allowlisted-subdomains.blocklist.conf: no such file or directory"),
			expErr: true,
			name:   "Allowlisted domains",
		},
		{
			c:      newCfg(),
			cfg:    testallCfg,
			ct:     PreHObj,
			err:    errors.New("open /:~//hosts.blocklisted-servers.blocklist.conf: no such file or directory"),
			expErr: true,
			name:   "Blocklisted hosts",
		},
		{
			c:      newCfg(),
			cfg:    testallCfg,
			ct:     PreDObj,
			err:    errors.New("open /:~//domains.blocklisted-subdomains.blocklist.conf: no such file or directory"),
			expErr: true,
			name:   "Blocklisted domains",
		},
		{
			c:      newCfg(),
			cfg:    testallCfg,
			ct:     ExRtObj,
			err:    errors.New("open /:~//roots.global-allowlisted-domains.blocklist.conf: no such file or directory"),
			expErr: true,
			name:   "Global allowlist",
		},
		{
			c:      newCfg(),
			cfg:    testCfg,
			ct:     FileObj,
			err:    fmt.Errorf("open /:~/=../../internal/testdata/blist.hosts.src: no such file or directory"),
			expErr: true,
			name:   "Non-existent File source",
		},
	}
	for _, tt := range tests {
		t.Run("current test: "+tt.name, func(t *testing.T) {
			if tt.name == "" {
				tt.c.Dir = "/:~/"
			}
			if err := tt.c.Blocklist(&CFGstatic{Cfg: tt.cfg}); err != nil {
				t.Fatal(err)
			}

			obj, err := tt.c.NewContent(tt.ct)
			if err != nil {
				t.Fatal(err)
			}

			err = tt.c.ProcessContent(obj)
			if (err != nil) == tt.expErr {
				if err.Error() != tt.err.Error() {
					t.Errorf("got %q, want %q", err.Error(), tt.err.Error())
				}
			}
		})
	}

	t.Run("Testing ProcessContent() if no arguments ", func(t *testing.T) {
		// var g errgroup.Group
		// g.Go(func() error { return newCfg().ProcessContent() })
		// err := g.Wait()
		if newCfg().ProcessContent() == nil {
			t.Fatal("expected non-nil")
		}
	})
}

func TestNewContent(t *testing.T) {
	// ytimg lines are dropped once global-allowlisted domains are merged into Dex.
	expFileObj := "address=/0.really.bad.phishing.site.ru/0.0.0.0\naddress=/cw.bad.ultraadverts.site.eu/0.0.0.0\naddress=/really.bad.phishing.site.ru/0.0.0.0"

	type newContentTT struct {
		err       error
		exp       string
		fail      bool
		i         int
		ltype     string
		name      string
		obj       IFace
		page      string
		page2     string
		pageData  string
		pageData2 string
		pos       int
		svr       *HTTPserver
		svr2      *HTTPserver
	}

	testsPhase1 := []newContentTT{
		{
			i:     1,
			exp:   excRootContent,
			fail:  false,
			ltype: ExcRoots,
			name:  "z" + ExcRoots,
			obj:   ExRtObj,
			pos:   -1,
		},
		{
			i:     1,
			exp:   "server=/adinfuse.com/#",
			fail:  false,
			ltype: ExcDomns,
			name:  "z" + ExcDomns,
			obj:   ExDmObj,
			pos:   -1,
		},
		{
			i:     1,
			exp:   "server=/wv.inner-active.mobi/#",
			fail:  false,
			ltype: ExcHosts,
			name:  "z" + ExcHosts,
			obj:   ExHtObj,
			pos:   -1,
		},
		{
			i:     1,
			exp:   "address=/adtechus.net/192.1.1.1\naddress=/advertising.com/192.1.1.1\naddress=/centade.com/192.1.1.1\naddress=/doubleclick.net/192.1.1.1\naddress=/intellitxt.com/192.1.1.1\naddress=/patoghee.in/192.1.1.1",
			fail:  false,
			ltype: PreDomns,
			name:  "z" + PreDomns,
			obj:   PreDObj,
			pos:   -1,
		},
		{
			i:     1,
			exp:   "address=/beap.gemini.yahoo.com/0.0.0.0",
			fail:  false,
			ltype: PreHosts,
			name:  "z" + PreHosts,
			obj:   PreHObj,
			pos:   -1,
		},
		{
			i:     1,
			exp:   expFileObj,
			fail:  false,
			ltype: files,
			name:  "ztasty",
			obj:   FileObj,
			pos:   -1,
		},
		{
			i:         1,
			exp:       domainsContent,
			fail:      false,
			ltype:     urls,
			name:      "zmalc0de",
			obj:       URLdObj,
			pos:       -1,
			page:      "/hosts.txt",
			page2:     "/domains.txt",
			pageData:  httpHostData,
			pageData2: HTTPDomainData,
			svr:       new(HTTPserver),
			svr2:      new(HTTPserver),
		},
		{
			i:         1,
			exp:       hostsContent,
			fail:      false,
			ltype:     urls,
			name:      "zadaway",
			obj:       URLhObj,
			pos:       -1,
			page:      "/hosts.txt",
			page2:     "/domains.txt",
			pageData:  httpHostData,
			pageData2: HTTPDomainData,
			svr:       new(HTTPserver),
			svr2:      new(HTTPserver),
		},
	}

	testsPhase2 := []newContentTT{
		{
			i:     1,
			exp:   excRootContent,
			fail:  false,
			ltype: ExcRoots,
			name:  ExcRoots,
			obj:   ExRtObj,
			pos:   0,
		},
		{
			i:     1,
			exp:   "",
			fail:  false,
			ltype: PreRoots,
			name:  "z" + PreRoots,
			obj:   PreRObj,
			pos:   -1,
		},
		{
			i:     1,
			exp:   "server=/adinfuse.com/#",
			fail:  false,
			ltype: ExcDomns,
			name:  ExcDomns,
			obj:   ExDmObj,
			pos:   0,
		},
		{
			i:     1,
			exp:   "server=/wv.inner-active.mobi/#",
			fail:  false,
			ltype: ExcHosts,
			name:  ExcHosts,
			obj:   ExHtObj,
			pos:   0,
		},
		{
			i:     1,
			exp:   "address=/adtechus.net/192.1.1.1\naddress=/advertising.com/192.1.1.1\naddress=/centade.com/192.1.1.1\naddress=/doubleclick.net/192.1.1.1\naddress=/intellitxt.com/192.1.1.1\naddress=/patoghee.in/192.1.1.1",
			fail:  false,
			ltype: PreDomns,
			name:  PreDomns,
			obj:   PreDObj,
			pos:   0,
		},
		{
			i:     1,
			exp:   "address=/beap.gemini.yahoo.com/0.0.0.0",
			fail:  false,
			ltype: PreHosts,
			name:  PreHosts,
			obj:   PreHObj,
			pos:   0,
		},
		{
			i:     1,
			exp:   "",
			fail:  false,
			ltype: PreRoots,
			name:  PreRoots,
			obj:   PreRObj,
			pos:   0,
		},
		{
			i:     1,
			exp:   expFileObj,
			fail:  false,
			ltype: files,
			name:  "tasty",
			obj:   FileObj,
			pos:   0,
		},
		{
			i:         1,
			exp:       domainsContent,
			fail:      false,
			ltype:     urls,
			name:      "malc0de",
			obj:       URLdObj,
			pos:       0,
			page:      "/hosts.txt",
			page2:     "/domains.txt",
			pageData:  httpHostData,
			pageData2: HTTPDomainData,
			svr:       new(HTTPserver),
			svr2:      new(HTTPserver),
		},
		{
			i:         1,
			exp:       hostsContent,
			fail:      false,
			ltype:     urls,
			name:      "adaway",
			obj:       URLhObj,
			pos:       0,
			page:      "/hosts.txt",
			page2:     "/domains.txt",
			pageData:  httpHostData,
			pageData2: HTTPDomainData,
			svr:       new(HTTPserver),
			svr2:      new(HTTPserver),
		},
		{
			i:    0,
			err:  errors.New("invalid interface requested"),
			fail: true,
			obj:  Invalid,
			pos:  -1,
		},
	}

	runNewContentPhase := func(t *testing.T, phase string, tests []newContentTT) {
		t.Helper()
		t.Run(phase, func(t *testing.T) {
			c := NewConfig(
				API("/bin/cli-shell-api"),
				Arch(runtime.GOARCH),
				Bash("/bin/bash"),
				Cores(runtime.NumCPU()),
				Dir("/tmp"),
				Disabled(false),
				DNSsvc("service dnsmasq restart"),
				Ext("blocklist.conf"),
				FileNameFmt("%v/%v.%v.%v"),
				InCLI("inSession"),
				SetLogger(newLog()),
				Method("GET"),
				Prefix("address=", "server="),
				Timeout(30*time.Second),
				WCard(Wildcard{Node: "*s", Name: "*"}),
			)

			if err := c.Blocklist(&CFGstatic{Cfg: Cfg}); err != nil {
				t.Fatal(err)
			}

			c.Dex.merge(&list{RWMutex: &sync.RWMutex{}, entry: entry{"amazon-de.com": struct{}{}}})
			wantDex := `"amazon-de.com":{},
`
			if got := c.Dex.String(); got != wantDex {
				t.Errorf("Dex.String mismatch")
			}

			for _, tt := range tests {
				t.Run("processing "+tt.name, func(t *testing.T) {
					objs, err := c.NewContent(tt.obj)
					if tt.ltype == urls {
						uri1 := tt.svr.NewHTTPServer().String() + tt.page
						objs.SetURL("adaway", uri1)
						uri2 := tt.svr2.NewHTTPServer().String() + tt.page2
						objs.SetURL("malc0de", uri2)

						go tt.svr.Mux.HandleFunc(tt.page,
							func(w http.ResponseWriter, r *http.Request) {
								fmt.Fprint(w, tt.pageData)
							},
						)

						go tt.svr2.Mux.HandleFunc(tt.page2,
							func(w http.ResponseWriter, r *http.Request) {
								fmt.Fprint(w, tt.pageData2)
							},
						)
					}

					switch tt.fail {
					case false:
						if err != nil {
							t.Fatal(err)
						}

						d := &dummyConfig{Env: c.Env, t: t}
						if err := d.ProcessContent(objs); err != nil {
							t.Fatal(err)
						}

						if got, want := strings.Join(d.s, "\n"), tt.exp; got != want {
							t.Errorf("dummy join: got %q want %q", got, want)
						}

						objs.SetURL(tt.name, tt.name)

						if objs.Find(tt.name) != tt.pos {
							t.Errorf("got %v want %v", objs.Find(tt.name), tt.pos)
						}
						if objs.Len() != tt.i {
							t.Errorf("got %v want %v", objs.Len(), tt.i)
						}

					default:
						if err.Error() != tt.err.Error() {
							t.Errorf("got %v want %v", err.Error(), tt.err.Error())
						}
					}
				})
			}
		})
	}

	runNewContentPhase(t, "phase1_prefix_z", testsPhase1)
	runNewContentPhase(t, "phase2_unprefixed", testsPhase2)
}

func TestContenterString(t *testing.T) {
	c := NewConfig(
		Dir("/tmp"),
		Ext("blocklist.conf"),
		FileNameFmt("%v/%v.%v.%v"),
		Method("GET"),
		Prefix("address=", "server="),
	)

	if err := c.Blocklist(&CFGstatic{Cfg: testallCfg}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		iFace IFace
		exp   string
		name  string
	}{
		{name: "ExDmObj", iFace: ExDmObj, exp: "\nDesc:         \"pre-configured allowlisted subdomains\"\nDisabled:     \"false\"\nFile:         \"**Undefined**\"\nIP:           \"0.0.0.0\"\nLtype:        \"allowlisted-subdomains\"\nName:         \"allowlisted-subdomains\"\nnType:        \"excDomn\"\nPrefix:       \"**Undefined**\"\nType:         \"allowlisted-subdomains\"\nURL:          \"**Undefined**\"\nAllowlist:\n              \"**No entries found**\"\nBlocklist:\n              \"**No entries found**\"\n"},
		{name: "ExHtObj", iFace: ExHtObj, exp: "\nDesc:         \"pre-configured allowlisted servers\"\nDisabled:     \"false\"\nFile:         \"**Undefined**\"\nIP:           \"192.168.168.1\"\nLtype:        \"allowlisted-servers\"\nName:         \"allowlisted-servers\"\nnType:        \"excHost\"\nPrefix:       \"**Undefined**\"\nType:         \"allowlisted-servers\"\nURL:          \"**Undefined**\"\nAllowlist:\n              \"**No entries found**\"\nBlocklist:\n              \"**No entries found**\"\n"},
		{name: "ExRtObj", iFace: ExRtObj, exp: "\nDesc:         \"pre-configured global allowlisted domains\"\nDisabled:     \"false\"\nFile:         \"**Undefined**\"\nIP:           \"0.0.0.0\"\nLtype:        \"global-allowlisted-domains\"\nName:         \"global-allowlisted-domains\"\nnType:        \"excRoot\"\nPrefix:       \"**Undefined**\"\nType:         \"global-allowlisted-domains\"\nURL:          \"**Undefined**\"\nAllowlist:\n              \"122.2o7.net\"\n              \"1e100.net\"\n              \"adobedtm.com\"\n              \"akamai.net\"\n              \"amazon.com\"\n              \"amazonaws.com\"\n              \"apple.com\"\n              \"ask.com\"\n              \"avast.com\"\n              \"bitdefender.com\"\n              \"cdn.visiblemeasures.com\"\n              \"cloudfront.net\"\n              \"coremetrics.com\"\n              \"edgesuite.net\"\n              \"freedns.afraid.org\"\n              \"github.com\"\n              \"githubusercontent.com\"\n              \"google.com\"\n              \"googleadservices.com\"\n              \"googleapis.com\"\n              \"googleusercontent.com\"\n              \"gstatic.com\"\n              \"gvt1.com\"\n              \"gvt1.net\"\n              \"hb.disney.go.com\"\n              \"hp.com\"\n              \"hulu.com\"\n              \"images-amazon.com\"\n              \"msdn.com\"\n              \"paypal.com\"\n              \"rackcdn.com\"\n              \"schema.org\"\n              \"skype.com\"\n              \"smacargo.com\"\n              \"sourceforge.net\"\n              \"ssl-on9.com\"\n              \"ssl-on9.net\"\n              \"static.chartbeat.com\"\n              \"storage.googleapis.com\"\n              \"windows.net\"\n              \"yimg.com\"\n              \"ytimg.com\"\nBlocklist:\n              \"**No entries found**\"\n"},
		{name: "FileObj", iFace: FileObj, exp: "\nDesc:         \"File source\"\nDisabled:     \"false\"\nFile:         \"../../internal/testdata/blist.hosts.src\"\nIP:           \"0.0.0.0\"\nLtype:        \"file\"\nName:         \"tasty\"\nnType:        \"host\"\nPrefix:       \"**Undefined**\"\nType:         \"hosts\"\nURL:          \"**Undefined**\"\nAllowlist:\n              \"**No entries found**\"\nBlocklist:\n              \"**No entries found**\"\n"},
		{name: "PreDObj", iFace: PreDObj, exp: "\nDesc:         \"pre-configured blocklisted subdomains\"\nDisabled:     \"false\"\nFile:         \"**Undefined**\"\nIP:           \"0.0.0.0\"\nLtype:        \"blocklisted-subdomains\"\nName:         \"blocklisted-subdomains\"\nnType:        \"preDomn\"\nPrefix:       \"**Undefined**\"\nType:         \"blocklisted-subdomains\"\nURL:          \"**Undefined**\"\nAllowlist:\n              \"**No entries found**\"\nBlocklist:\n              \"adtechus.net\"\n              \"advertising.com\"\n              \"centade.com\"\n              \"doubleclick.net\"\n              \"intellitxt.com\"\n              \"patoghee.in\"\n"},
		{name: "PreHObj", iFace: PreHObj, exp: "\nDesc:         \"pre-configured blocklisted servers\"\nDisabled:     \"false\"\nFile:         \"**Undefined**\"\nIP:           \"192.168.168.1\"\nLtype:        \"blocklisted-servers\"\nName:         \"blocklisted-servers\"\nnType:        \"preHost\"\nPrefix:       \"**Undefined**\"\nType:         \"blocklisted-servers\"\nURL:          \"**Undefined**\"\nAllowlist:\n              \"**No entries found**\"\nBlocklist:\n              \"beap.gemini.yahoo.com\"\n"},
		{name: "PreRObj", iFace: PreRObj, exp: "\nDesc:         \"pre-configured global blocklisted domains\"\nDisabled:     \"false\"\nFile:         \"**Undefined**\"\nIP:           \"0.0.0.0\"\nLtype:        \"global-blocklisted-domains\"\nName:         \"global-blocklisted-domains\"\nnType:        \"preRoot\"\nPrefix:       \"**Undefined**\"\nType:         \"global-blocklisted-domains\"\nURL:          \"**Undefined**\"\nAllowlist:\n              \"**No entries found**\"\nBlocklist:\n              \"**No entries found**\"\n"},
		{name: "URLdObj", iFace: URLdObj, exp: "\nDesc:         \"List of zones serving malicious executables observed by malc0de.com/database/\"\nDisabled:     \"false\"\nFile:         \"**Undefined**\"\nIP:           \"0.0.0.0\"\nLtype:        \"url\"\nName:         \"malc0de\"\nnType:        \"domn\"\nPrefix:       \"zone \"\nType:         \"domains\"\nURL:          \"http://127.0.0.1:8081/domains/domain.txt\"\nAllowlist:\n              \"**No entries found**\"\nBlocklist:\n              \"**No entries found**\"\n"},
		{name: "URLhObj", iFace: URLhObj, exp: "\nDesc:         \"Blocking mobile ad providers and some analytics providers\"\nDisabled:     \"false\"\nFile:         \"**Undefined**\"\nIP:           \"192.168.168.1\"\nLtype:        \"url\"\nName:         \"adaway\"\nnType:        \"host\"\nPrefix:       \"127.0.0.1 \"\nType:         \"hosts\"\nURL:          \"http://127.0.0.1:8081/hosts/host.txt\"\nAllowlist:\n              \"**No entries found**\"\nBlocklist:\n              \"**No entries found**\"\n"},
	}

	for _, tt := range tests {
		t.Run("Testing "+tt.name+" Contenter.String()", func(t *testing.T) {
			ct, err := c.NewContent(tt.iFace)
			if err != nil {
				t.Fatal(err)
			}
			if ct.String() != tt.exp {
				t.Errorf("got %v, want %v", ct.String(), tt.exp)
			}
		})
	}
}

func TestIFaceString(t *testing.T) {
	tests := []struct {
		iface IFace
		name  string
		exp   string
	}{
		{name: "ExDmObj", iface: ExDmObj, exp: ExcDomns},
		{name: "ExHtObj", iface: ExHtObj, exp: ExcHosts},
		{name: "ExRtObj", iface: ExRtObj, exp: ExcRoots},
		{name: "FileObj", iface: FileObj, exp: files},
		{name: "Invalid", iface: Invalid, exp: notknown},
		{name: "PreDObj", iface: PreDObj, exp: PreDomns},
		{name: "PreHObj", iface: PreHObj, exp: PreHosts},
		{name: "PreRObj", iface: PreRObj, exp: PreRoots},
		{name: "URLdObj", iface: URLdObj, exp: urls},
		{name: "URLhObj", iface: URLhObj, exp: urls},
	}

	for _, tt := range tests {
		t.Run("with "+tt.name, func(t *testing.T) {
			s := tt.iface.String()
			if s != tt.exp {
				t.Errorf("got %v, want %v", tt.iface.String(), tt.exp)
			}
		})
	}
}

func TestMultiObjNewContent(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "testBlocklist*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	c := NewConfig(
		Dir(dir),
		Ext("blocklist.conf"),
		FileNameFmt("%v/%v.%v.%v"),
		SetLogger(newLog()),
		Method("GET"),
		Prefix("address=", "server="),
	)

	if err := c.Blocklist(&CFGstatic{Cfg: CfgMimimal}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		iFace IFace
		exp   string
		name  string
	}{
		{name: "ExRtObj", iFace: ExRtObj, exp: "server=/ytimg.com/#"},
		{name: "ExDmObj", iFace: ExDmObj, exp: ""},
		{name: "ExHtObj", iFace: ExHtObj, exp: ""},
		{name: "PreDObj", iFace: PreDObj, exp: "address=/awfuladvertising.com/0.0.0.0\naddress=/badadsrvr.org/0.0.0.0\naddress=/badintellitxt.com/0.0.0.0\naddress=/disgusting.unkiosked.com/0.0.0.0\naddress=/filthydoubleclick.net/0.0.0.0\naddress=/iffyfree-counter.co.uk/0.0.0.0\naddress=/nastycentade.com/0.0.0.0\naddress=/worseadtechus.net/0.0.0.0"},
		{name: "PreHObj", iFace: PreHObj, exp: "address=/beap.gemini.yahoo.com/192.168.168.1"},
		{name: "PreRObj", iFace: PreRObj, exp: "address=/adtechus.net/0.0.0.0\naddress=/advertising.com/0.0.0.0\naddress=/centade.com/0.0.0.0\naddress=/doubleclick.net/0.0.0.0\naddress=/intellitxt.com/0.0.0.0\naddress=/patoghee.in/0.0.0.0"},
		{name: "FileObj", iFace: FileObj, exp: expFileObj},
		{name: "URLdObj", iFace: URLdObj, exp: expURLdObj},
		{name: "URLhObj", iFace: URLhObj, exp: expURLhOBJ},
	}

	for _, tt := range tests {
		t.Run("Testing "+tt.name+" ProcessContent()", func(t *testing.T) {
			ct, err := c.NewContent(tt.iFace)
			if err != nil {
				t.Fatal(err)
			}

			switch tt.iFace {
			case ExRtObj, ExDmObj, ExHtObj, PreDObj, PreHObj, PreRObj:
				d := &dummyConfig{Env: c.Env, t: t}
				if err := d.ProcessContent(ct); err != nil {
					t.Fatal(err)
				}
				if got, want := strings.Join(d.s, "\n"), tt.exp; got != want {
					t.Errorf("dummy join: got %q want %q", got, want)
				}
			default:
				if ct.String() != tt.exp {
					t.Errorf("got %v, want %v", ct.String(), tt.exp)
				}
			}
		})
	}
}

func TestProcessContent(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "testBlocklist*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	t.Run("Testing ProcessContent()", func(t *testing.T) {
		c := NewConfig(
			Dir(dir),
			Ext("blocklist.conf"),
			FileNameFmt("%v/%v.%v.%v"),
			SetLogger(newLog()),
			Method("GET"),
			Prefix("address=", "server="),
		)

		tests := []struct {
			dropped   int32
			extracted int32
			kept      int32
			err       error
			exp       string
			expDexMap list
			expExcMap list
			f         string
			fdata     string
			name      string
			obj       IFace
		}{
			{
				name:      "ExRtObj",
				dropped:   0,
				extracted: 1,
				kept:      1,
				err:       nil,
				exp: `
Desc:         "pre-configured global allowlisted domains"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "global-allowlisted-domains"
Name:         "global-allowlisted-domains"
nType:        "excRoot"
Prefix:       "**Undefined**"
Type:         "global-allowlisted-domains"
URL:          "**Undefined**"
Allowlist:
              "ytimg.com"
Blocklist:
              "**No entries found**"
`,
				expDexMap: list{entry: entry{"ytimg.com": struct{}{}}},
				expExcMap: list{entry: entry{"ytimg.com": struct{}{}}},
				obj:       ExRtObj,
			},
			{
				name:      "ExDmObj",
				dropped:   0,
				extracted: 0,
				kept:      0,
				err:       nil,
				exp: `
Desc:         "pre-configured allowlisted subdomains"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "allowlisted-subdomains"
Name:         "allowlisted-subdomains"
nType:        "excDomn"
Prefix:       "**Undefined**"
Type:         "allowlisted-subdomains"
URL:          "**Undefined**"
Allowlist:
              "**No entries found**"
Blocklist:
              "**No entries found**"
`,
				expDexMap: list{RWMutex: &sync.RWMutex{}, entry: make(entry)},
				expExcMap: list{RWMutex: &sync.RWMutex{}, entry: make(entry)},
				obj:       ExDmObj,
			},
			{
				name:      "ExHtObj",
				dropped:   0,
				extracted: 0,
				kept:      0,
				err:       nil,
				exp: `
Desc:         "pre-configured allowlisted servers"
Disabled:     "false"
File:         "**Undefined**"
IP:           "192.168.168.1"
Ltype:        "allowlisted-servers"
Name:         "allowlisted-servers"
nType:        "excHost"
Prefix:       "**Undefined**"
Type:         "allowlisted-servers"
URL:          "**Undefined**"
Allowlist:
              "**No entries found**"
Blocklist:
              "**No entries found**"
`,
				expDexMap: list{RWMutex: &sync.RWMutex{}, entry: make(entry)},
				expExcMap: list{RWMutex: &sync.RWMutex{}, entry: make(entry)},
				obj:       ExHtObj,
			},
			{
				name:      "PreDObj",
				dropped:   0,
				extracted: 8,
				kept:      8,
				err:       nil,
				exp: `
Desc:         "pre-configured blocklisted subdomains"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "blocklisted-subdomains"
Name:         "blocklisted-subdomains"
nType:        "preDomn"
Prefix:       "**Undefined**"
Type:         "blocklisted-subdomains"
URL:          "**Undefined**"
Allowlist:
              "**No entries found**"
Blocklist:
              "adtechus.net"
              "advertising.com"
              "centade.com"
              "doubleclick.net"
              "intellitxt.com"
              "patoghee.in"
`,
				expDexMap: list{
					entry: entry{
						"adtechus.net":    struct{}{},
						"advertising.com": struct{}{},
						"centade.com":     struct{}{},
						"doubleclick.net": struct{}{},
						"intellitxt.com":  struct{}{},
						"patoghee.in":     struct{}{},
					},
				},
				expExcMap: list{entry: entry{"ytimg.com": struct{}{}}},
				f:         dir + "/domains.blocklisted-subdomains.blocklist.conf",
				fdata: `address=/awfuladvertising.com/0.0.0.0
address=/badadsrvr.org/0.0.0.0
address=/badintellitxt.com/0.0.0.0
address=/disgusting.unkiosked.com/0.0.0.0
address=/filthydoubleclick.net/0.0.0.0
address=/iffyfree-counter.co.uk/0.0.0.0
address=/nastycentade.com/0.0.0.0
address=/worseadtechus.net/0.0.0.0
`,
				obj: PreDObj,
			},
			{
				name:      "PreHObj",
				dropped:   0,
				extracted: 1,
				kept:      1,
				err:       nil,
				exp: `
Desc:         "pre-configured blocklisted servers"
Disabled:     "false"
File:         "**Undefined**"
IP:           "192.168.168.1"
Ltype:        "blocklisted-servers"
Name:         "blocklisted-servers"
nType:        "preHost"
Prefix:       "**Undefined**"
Type:         "blocklisted-servers"
URL:          "**Undefined**"
Allowlist:
              "**No entries found**"
Blocklist:
              "beap.gemini.yahoo.com"
`,
				expDexMap: list{entry: entry{"ytimg.com": struct{}{}}},
				expExcMap: list{entry: entry{"ytimg.com": struct{}{}}},
				f:         dir + "/hosts.blocklisted-servers.blocklist.conf",
				fdata:     "address=/beap.gemini.yahoo.com/192.168.168.1\n",
				obj:       PreHObj,
			},
			{
				name:      "PreRObj",
				dropped:   0,
				extracted: 6,
				kept:      6,
				err:       nil,
				exp: `
Desc:         "pre-configured global blocklisted domains"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "global-blocklisted-domains"
Name:         "global-blocklisted-domains"
nType:        "preRoot"
Prefix:       "**Undefined**"
Type:         "global-blocklisted-domains"
URL:          "**Undefined**"
Allowlist:
              "**No entries found**"
Blocklist:
              "adtechus.net"
              "advertising.com"
              "centade.com"
              "doubleclick.net"
              "intellitxt.com"
              "patoghee.in"
`,
				expDexMap: list{entry: entry{}},
				expExcMap: list{
					entry: entry{
						"adtechus.net":    struct{}{},
						"advertising.com": struct{}{},
						"centade.com":     struct{}{},
						"doubleclick.net": struct{}{},
						"intellitxt.com":  struct{}{},
						"patoghee.in":     struct{}{},
					},
				},
				obj: PreRObj,
			},
			{
				name:      "FileObj",
				dropped:   2,
				extracted: 23,
				kept:      21,
				err:       nil,
				exp:       filesMin,
				expDexMap: list{
					entry: entry{
						"cw.bad.ultraadverts.site.eu": struct{}{},
						"really.bad.phishing.site.ru": struct{}{},
					},
				},
				expExcMap: list{entry: entry{"ytimg.com": struct{}{}}},
				f:         dir + "/hosts.tasty.blocklist.conf",
				fdata: `address=/0.really.bad.phishing.site.ru/10.10.10.10
address=/cw.bad.ultraadverts.site.eu/10.10.10.10
address=/really.bad.phishing.site.ru/10.10.10.10
`,
				obj: FileObj,
			},
		}

		if err := c.Blocklist(&CFGstatic{Cfg: CfgMimimal}); err != nil {
			t.Fatal(err)
		}

		for _, tt := range tests {
			t.Run("Testing "+tt.name+" ProcessContent()", func(t *testing.T) {
				var (
					ct    Contenter
					objex []IFace
				)

				switch tt.obj {
				case FileObj, URLdObj, URLhObj:
					objex = []IFace{
						PreRObj,
						PreDObj,
						PreHObj,
						ExRtObj,
						ExDmObj,
						ExHtObj,
						tt.obj,
					}
				default:
					objex = []IFace{tt.obj}
				}

				var g errgroup.Group
				g.Go(
					func() (err error) {
						for _, o := range objex {
							ct, _ = c.NewContent(o)
							err = c.ProcessContent(ct)
						}
						return err
					})

				waitErr := g.Wait()
				if tt.err != nil {
					if waitErr == nil {
						t.Fatalf("expected error %v", tt.err)
					}
					if waitErr.Error() != tt.err.Error() {
						t.Fatalf("ProcessContent error: got %v want %v", waitErr, tt.err)
					}
					return
				}
				if waitErr != nil {
					t.Fatal(waitErr)
				}

				switch tt.f {
				case "":
					if ct.String() != tt.exp {
						t.Errorf("ct.String: got %q want %q", ct.String(), tt.exp)
					}
				default:
					reader, err := GetFile(tt.f)
					if err != nil {
						t.Fatal(err)
					}

					act, err := io.ReadAll(reader)
					if err != nil {
						t.Fatal(err)
					}

					if string(act) != tt.fdata {
						t.Errorf("file data: got %q want %q", string(act), tt.fdata)
					}
				}
			})
		}
	})
}

func TestProcessZeroContent(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "testBlocklist*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	c := NewConfig(
		Dir(dir),
		Ext("blocklist.conf"),
		FileNameFmt("%v/%v.%v.%v"),
		SetLogger(newLog()),
		Method("GET"),
		Prefix("address=", "server="),
	)

	err = c.Blocklist(&CFGstatic{Cfg: cfgRedundant})
	if err != nil {
		t.Fatal(err)
	}

	for _, o := range []IFace{ExRtObj, FileObj} {
		ct, err := c.NewContent(o)
		if err != nil {
			t.Fatal(err)
		}

		err = c.ProcessContent(ct)
		if err != nil {
			t.Fatal(err)
		}
	}

	dropped, extracted, kept := c.GetTotalStats()

	t.Run("Dropped entries should match", func(t *testing.T) {
		if dropped != 1 {
			t.Errorf("got %v, want %v", dropped, 1)
		}
	})

	t.Run("Extracted entries should match", func(t *testing.T) {
		if extracted != 2 {
			t.Errorf("got %v, want %v", extracted, 2)
		}
	})

	t.Run("Kept entries should match", func(t *testing.T) {
		if kept != 1 {
			t.Errorf("got %v, want %v", kept, 1)
		}
	})
}

func TestProcessBadFile(t *testing.T) {
	dir, err := os.MkdirTemp("/tmp", "testBlocklist*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	c := NewConfig(
		Dir("/:~/"),
		Ext("blocklist.conf"),
		FileNameFmt("%v/%v.%v.%v"),
		SetLogger(newLog()),
		Method("GET"),
		Prefix("address=", "server="),
	)

	err = c.Blocklist(&CFGstatic{Cfg: CfgMimimal})
	if err != nil {
		t.Fatal(err)
	}

	ct, err := c.NewContent(FileObj)
	if err != nil {
		t.Fatal(err)
	}

	err = c.ProcessContent(ct)
	if err.Error() != "open /:~//hosts.tasty.blocklist.conf: no such file or directory" {
		t.Errorf("got %v, want %v", err.Error(), "open /:~//hosts.tasty.blocklist.conf: no such file or directory")
	}
}

func TestWriteFile(t *testing.T) {
	tests := []struct {
		data  io.Reader
		dir   string
		fname string
		ok    bool
		want  string
	}{
		{
			data:  strings.NewReader("The rest is history!"),
			dir:   "/tmp",
			fname: "Test.util.writeFile",
			ok:    true,
			want:  "",
		},
		{
			data:  bytes.NewBuffer([]byte{84, 104, 101, 32, 114, 101, 115, 116, 32, 105, 115, 32, 104, 105, 115, 116, 111, 114, 121, 33}),
			dir:   "/tmp",
			fname: "Test.util.writeFile",
			ok:    true,
			want:  "",
		},
		{
			data:  bytes.NewBufferString("This shouldn't be written!"),
			dir:   "",
			fname: "/",
			ok:    false,
			want:  "open /: is a directory",
		},
	}

	for _, tt := range tests {
		switch tt.ok {
		case true:
			f, err := os.CreateTemp(tt.dir, tt.fname+"*")
			if err != nil {
				t.Fatal(err)
			}
			b := &bList{
				file: f.Name(),
				r:    tt.data,
				size: 20,
			}
			if err := b.writeFile(); err != nil {
				t.Fatal(err)
			}
			os.Remove(f.Name())

		default:
			b := &bList{
				file: tt.dir + tt.fname,
				r:    tt.data,
				size: 20,
			}
			err := b.writeFile()
			if err == nil {
				t.Fatal("expected error from writeFile")
			}
			msg := err.Error()
			if msg != tt.want && msg != "open /: file exists" {
				t.Errorf("writeFile: got %q want %q", msg, tt.want)
			}
		}
	}
}

var (
	// Cfg contains a valid full EdgeOS blocklist configuration
	Cfg = `blocklist {
    disabled false
    dns-redirect-ip 0.0.0.0
    domains {
        dns-redirect-ip 192.1.1.1
		exclude adinfuse.com
        include adtechus.net
        include advertising.com
        include centade.com
        include doubleclick.net
        include intellitxt.com
        include patoghee.in
        source malc0de {
            description "List of zones serving malicious executables observed by malc0de.com/database/"
            prefix "zone "
            url http://malc0de.com/bl/ZONES
        }
    }
    exclude 122.2o7.net
    exclude 1e100.net
    exclude adobedtm.com
    exclude akamai.net
    exclude amazon.com
    exclude amazonaws.com
    exclude apple.com
    exclude ask.com
    exclude avast.com
    exclude bitdefender.com
    exclude cdn.visiblemeasures.com
    exclude cloudfront.net
    exclude coremetrics.com
    exclude edgesuite.net
    exclude freedns.afraid.org
    exclude github.com
    exclude githubusercontent.com
    exclude google.com
    exclude googleadservices.com
    exclude googleapis.com
    exclude googleusercontent.com
    exclude gstatic.com
    exclude gvt1.com
    exclude gvt1.net
    exclude hb.disney.go.com
    exclude hp.com
    exclude hulu.com
    exclude images-amazon.com
	exclude jumptap.com
    exclude msdn.com
    exclude paypal.com
    exclude rackcdn.com
    exclude schema.org
    exclude skype.com
    exclude smacargo.com
    exclude sourceforge.net
    exclude ssl-on9.com
    exclude ssl-on9.net
    exclude static.chartbeat.com
    exclude storage.googleapis.com
	exclude usemaxserver.de
    exclude windows.net
    exclude yimg.com
    exclude ytimg.com
    hosts {
		exclude wv.inner-active.mobi
        include beap.gemini.yahoo.com
        source adaway {
            description "Blocking mobile ad providers and some analytics providers"
			dns-redirect-ip 192.168.168.1
            prefix "127.0.0.1 "
            url http://adaway.org/hosts.txt
        }
				source tasty {
						description "File source"
						dns-redirect-ip 0.0.0.0
						file ../../internal/testdata/blist.hosts.src
				}
    }
}`

	cfgRedundant = `blocklist {
	disabled false
	dns-redirect-ip 0.0.0.0
	domains {
		source tasty {
			description "File source"
			dns-redirect-ip 10.10.10.10
			file ../../internal/testdata/blist.nohosts.src
	}
	}
	exclude ytimg.com
}`
	// CfgMimimal contains a valid minimal EdgeOS blocklist configuration
	CfgMimimal = `blocklist {
	disabled false
	dns-redirect-ip 0.0.0.0
	domains {
			include badadsrvr.org
			include worseadtechus.net
			include awfuladvertising.com
			include nastycentade.com
			include filthydoubleclick.net
			include iffyfree-counter.co.uk
			include badintellitxt.com
			include disgusting.unkiosked.com
			source malc0de {
					description "List of zones serving malicious executables observed by malc0de.com/database/"
					prefix "zone "
					url http://malc0de.com/bl/ZONES
			}
	}
	exclude ytimg.com
	include adtechus.net
	include advertising.com
	include centade.com
	include doubleclick.net
	include intellitxt.com
	include patoghee.in
	hosts {
			dns-redirect-ip 192.168.168.1
			include beap.gemini.yahoo.com
			source tasty {
					description "File source"
					dns-redirect-ip 10.10.10.10
					file ../../internal/testdata/blist.hosts.src
			}
			source adaway {
          description "Blocking mobile ad providers and some analytics providers"
			    dns-redirect-ip 192.168.168.1
          prefix "127.0.0.1 "
          url http://adaway.org/hosts.txt
      }
	}
}`

	// testallCfg contains a valid full EdgeOS blocklist configuration with localized URLs
	testallCfg = `blocklist {
	disabled false
	dns-redirect-ip 0.0.0.0
	domains {
			dns-redirect-ip 0.0.0.0
			include adtechus.net
			include advertising.com
			include centade.com
			include doubleclick.net
			include intellitxt.com
			include patoghee.in
			source malc0de {
					description "List of zones serving malicious executables observed by malc0de.com/database/"
					prefix "zone "
					url http://127.0.0.1:8081/domains/domain.txt
			}
	}
	exclude 122.2o7.net
	exclude 1e100.net
	exclude adobedtm.com
	exclude akamai.net
	exclude amazon.com
	exclude amazonaws.com
	exclude apple.com
	exclude ask.com
	exclude avast.com
	exclude bitdefender.com
	exclude cdn.visiblemeasures.com
	exclude cloudfront.net
	exclude coremetrics.com
	exclude edgesuite.net
	exclude freedns.afraid.org
	exclude github.com
	exclude githubusercontent.com
	exclude google.com
	exclude googleadservices.com
	exclude googleapis.com
	exclude googleusercontent.com
	exclude gstatic.com
	exclude gvt1.com
	exclude gvt1.net
	exclude hb.disney.go.com
	exclude hp.com
	exclude hulu.com
	exclude images-amazon.com
	exclude msdn.com
	exclude paypal.com
	exclude rackcdn.com
	exclude schema.org
	exclude skype.com
	exclude smacargo.com
	exclude sourceforge.net
	exclude ssl-on9.com
	exclude ssl-on9.net
	exclude static.chartbeat.com
	exclude storage.googleapis.com
	exclude windows.net
	exclude yimg.com
	exclude ytimg.com
	hosts {
			dns-redirect-ip 192.168.168.1
			include beap.gemini.yahoo.com
			source adaway {
					description "Blocking mobile ad providers and some analytics providers"
					prefix "127.0.0.1 "
					url http://127.0.0.1:8081/hosts/host.txt
			}
			source tasty {
					description "File source"
					dns-redirect-ip 0.0.0.0
					file ../../internal/testdata/blist.hosts.src
			}
	}
}`

	hostsContent = `address=/a.applovin.com/192.168.168.1
address=/a.glcdn.co/192.168.168.1
address=/a.vserv.mobi/192.168.168.1
address=/ad.leadboltapps.net/192.168.168.1
address=/ad.madvertise.de/192.168.168.1
address=/ad.where.com/192.168.168.1
address=/adcontent.saymedia.com/192.168.168.1
address=/admicro1.vcmedia.vn/192.168.168.1
address=/admicro2.vcmedia.vn/192.168.168.1
address=/admin.vserv.mobi/192.168.168.1
address=/ads.adiquity.com/192.168.168.1
address=/ads.admarvel.com/192.168.168.1
address=/ads.admoda.com/192.168.168.1
address=/ads.celtra.com/192.168.168.1
address=/ads.flurry.com/192.168.168.1
address=/ads.matomymobile.com/192.168.168.1
address=/ads.mobgold.com/192.168.168.1
address=/ads.mobilityware.com/192.168.168.1
address=/ads.mopub.com/192.168.168.1
address=/ads.n-ws.org/192.168.168.1
address=/ads.ookla.com/192.168.168.1
address=/ads.saymedia.com/192.168.168.1
address=/ads.smartdevicemedia.com/192.168.168.1
address=/ads.srcxad.net/192.168.168.1
address=/ads.vserv.mobi/192.168.168.1
address=/ads2.mediaarmor.com/192.168.168.1
address=/adserver.ubiyoo.com/192.168.168.1
address=/adultmoda.com/192.168.168.1
address=/android-sdk31.transpera.com/192.168.168.1
address=/android.bcfads.com/192.168.168.1
address=/api.airpush.com/192.168.168.1
address=/api.analytics.omgpop.com/192.168.168.1
address=/api.yp.com/192.168.168.1
address=/apps.buzzcity.net/192.168.168.1
address=/apps.mobilityware.com/192.168.168.1
address=/as.adfonic.net/192.168.168.1
address=/asotrack1.fluentmobile.com/192.168.168.1
address=/assets.cntdy.mobi/192.168.168.1
address=/atti.velti.com/192.168.168.1
address=/b.scorecardresearch.com/192.168.168.1
address=/banners.bigmobileads.com/192.168.168.1
address=/bigmobileads.com/192.168.168.1
address=/c.vrvm.com/192.168.168.1
address=/c.vserv.mobi/192.168.168.1
address=/cache-ssl.celtra.com/192.168.168.1
address=/cache.celtra.com/192.168.168.1
address=/cdn.celtra.com/192.168.168.1
address=/cdn.nearbyad.com/192.168.168.1
address=/cdn.trafficforce.com/192.168.168.1
address=/cdn.us.goldspotmedia.com/192.168.168.1
address=/cdn.vdopia.com/192.168.168.1
address=/cdn1.crispadvertising.com/192.168.168.1
address=/cdn1.inner-active.mobi/192.168.168.1
address=/cdn2.crispadvertising.com/192.168.168.1
address=/click.buzzcity.net/192.168.168.1
address=/creative1cdn.mobfox.com/192.168.168.1
address=/d.applovin.com/192.168.168.1
address=/edge.reporo.net/192.168.168.1
address=/ftpcontent.worldnow.com/192.168.168.1
address=/gemini.yahoo.com/192.168.168.1
address=/go.mobpartner.mobi/192.168.168.1
address=/go.vrvm.com/192.168.168.1
address=/gsmtop.net/192.168.168.1
address=/gts-ads.twistbox.com/192.168.168.1
address=/hhbeksrcw5d9e.pflexads.com/192.168.168.1
address=/hybl9bazbc35.pflexads.com/192.168.168.1
address=/i.tapit.com/192.168.168.1
address=/images.millennialmedia.com/192.168.168.1
address=/images.mpression.net/192.168.168.1
address=/img.ads.huntmad.com/192.168.168.1
address=/img.ads.mobilefuse.net/192.168.168.1
address=/img.ads.mocean.mobi/192.168.168.1
address=/img.ads.mojiva.com/192.168.168.1
address=/img.ads.taptapnetworks.com/192.168.168.1
address=/m.adsymptotic.com/192.168.168.1
address=/m2m1.inner-active.mobi/192.168.168.1
address=/media.mobpartner.mobi/192.168.168.1
address=/medrx.sensis.com.au/192.168.168.1
address=/mobile.banzai.it/192.168.168.1
address=/mobiledl.adboe.com/192.168.168.1
address=/mobpartner.mobi/192.168.168.1
address=/mwc.velti.com/192.168.168.1
address=/netdna.reporo.net/192.168.168.1
address=/oasc04012.247realmedia.com/192.168.168.1
address=/orencia.pflexads.com/192.168.168.1
address=/pdn.applovin.com/192.168.168.1
address=/r.edge.inmobicdn.net/192.168.168.1
address=/r.mobpartner.mobi/192.168.168.1
address=/req.appads.com/192.168.168.1
address=/rs-staticart.ybcdn.net/192.168.168.1
address=/ru.velti.com/192.168.168.1
address=/s0.2mdn.net/192.168.168.1
address=/s3.phluant.com/192.168.168.1
address=/sf.vserv.mobi/192.168.168.1
address=/show.buzzcity.net/192.168.168.1
address=/static.cdn.gtsmobi.com/192.168.168.1
address=/static.estebull.com/192.168.168.1
address=/stats.pflexads.com/192.168.168.1
address=/track.celtra.com/192.168.168.1
address=/tracking.klickthru.com/192.168.168.1
address=/www.eltrafiko.com/192.168.168.1
address=/www.mmnetwork.mobi/192.168.168.1
address=/www.pflexads.com/192.168.168.1
address=/wwww.adleads.com/192.168.168.1`

	domainsContent = "address=/192-168-0-255.com/192.1.1.1\naddress=/asi-37.fr/192.1.1.1\naddress=/bagbackpack.com/192.1.1.1\naddress=/bitmeyenkartusistanbul.com/192.1.1.1\naddress=/byxon.com/192.1.1.1\naddress=/img001.com/192.1.1.1\naddress=/loadto.net/192.1.1.1\naddress=/roastfiles2017.com/192.1.1.1"

	// domainsPreContent = "address=/adsrvr.org/192.1.1.1\naddress=/adtechus.net/192.1.1.1\naddress=/advertising.com/192.1.1.1\naddress=/centade.com/192.1.1.1\naddress=/doubleclick.net/192.1.1.1\naddress=/free-counter.co.uk/192.1.1.1\naddress=/intellitxt.com/192.1.1.1\naddress=/kiosked.com/192.1.1.1\n"

	// expPreGetAll = "[\nDesc:\t \"pre-configured blocklisted subdomains\"\nDisabled: false\nFile:\t \"\"\nIP:\t \"0.0.0.0\"\nLtype:\t \"blocklisted-subdomains\"\nName:\t \"blocklisted-subdomains\"\nnType:\t \"preDomn\"\nPrefix:\t \"\"\nType:\t \"blocklisted-subdomains\"\nURL:\t \"\"\n \nDesc:\t \"pre-configured blocklisted servers\"\nDisabled: false\nFile:\t \"\"\nIP:\t \"192.168.168.1\"\nLtype:\t \"blocklisted-servers\"\nName:\t \"blocklisted-servers\"\nnType:\t \"preHost\"\nPrefix:\t \"\"\nType:\t \"blocklisted-servers\"\nURL:\t \"\"\n]"

	// expAll = "[\nDesc:\t \"pre-configured blocklisted subdomains\"\nDisabled: false\nFile:\t \"\"\nIP:\t \"0.0.0.0\"\nLtype:\t \"blocklisted-subdomains\"\nName:\t \"blocklisted-subdomains\"\nnType:\t \"preDomn\"\nPrefix:\t \"\"\nType:\t \"blocklisted-subdomains\"\nURL:\t \"\"\n \nDesc:\t \"List of zones serving malicious executables observed by malc0de.com/database/\"\nDisabled: false\nFile:\t \"\"\nIP:\t \"0.0.0.0\"\nLtype:\t \"url\"\nName:\t \"malc0de\"\nnType:\t \"domn\"\nPrefix:\t \"zone \"\nType:\t \"domains\"\nURL:\t \"http://127.0.0.1:8081/domains/domain.txt\"\n \nDesc:\t \"pre-configured blocklisted servers\"\nDisabled: false\nFile:\t \"\"\nIP:\t \"192.168.168.1\"\nLtype:\t \"blocklisted-servers\"\nName:\t \"blocklisted-servers\"\nnType:\t \"preHost\"\nPrefix:\t \"\"\nType:\t \"blocklisted-servers\"\nURL:\t \"\"\n \nDesc:\t \"Blocking mobile ad providers and some analytics providers\"\nDisabled: false\nFile:\t \"\"\nIP:\t \"192.168.168.1\"\nLtype:\t \"url\"\nName:\t \"adaway\"\nnType:\t \"host\"\nPrefix:\t \"127.0.0.1 \"\nType:\t \"hosts\"\nURL:\t \"http://127.0.0.1:8081/hosts/host.txt\"\n \nDesc:\t \"File source\"\nDisabled: false\nFile:\t \"../../internal/testdata/blist.hosts.src\"\nIP:\t \"0.0.0.0\"\nLtype:\t \"file\"\nName:\t \"tasty\"\nnType:\t \"host\"\nPrefix:\t \"\"\nType:\t \"hosts\"\nURL:\t \"\"\n]"

	expFileObj = `
Desc:         "File source"
Disabled:     "false"
File:         "../../internal/testdata/blist.hosts.src"
IP:           "10.10.10.10"
Ltype:        "file"
Name:         "tasty"
nType:        "host"
Prefix:       "**Undefined**"
Type:         "hosts"
URL:          "**Undefined**"
Allowlist:
              "**No entries found**"
Blocklist:
              "**No entries found**"
`

	expURLdObj = `
Desc:         "List of zones serving malicious executables observed by malc0de.com/database/"
Disabled:     "false"
File:         "**Undefined**"
IP:           "0.0.0.0"
Ltype:        "url"
Name:         "malc0de"
nType:        "domn"
Prefix:       "zone "
Type:         "domains"
URL:          "http://malc0de.com/bl/ZONES"
Allowlist:
              "**No entries found**"
Blocklist:
              "**No entries found**"
`

	expURLhOBJ = `
Desc:         "Blocking mobile ad providers and some analytics providers"
Disabled:     "false"
File:         "**Undefined**"
IP:           "192.168.168.1"
Ltype:        "url"
Name:         "adaway"
nType:        "host"
Prefix:       "127.0.0.1 "
Type:         "hosts"
URL:          "http://adaway.org/hosts.txt"
Allowlist:
              "**No entries found**"
Blocklist:
              "**No entries found**"
`

	filesMin = "[\nDesc:\t \"File source\"\nDisabled: false\nFile:\t \"../../internal/testdata/blist.hosts.src\"\nIP:\t \"10.10.10.10\"\nLtype:\t \"file\"\nName:\t \"tasty\"\nnType:\t \"host\"\nPrefix:\t \"\"\nType:\t \"hosts\"\nURL:\t \"\"\n \nDesc:\t \"File source\"\nDisabled: false\nFile:\t \"../../internal/testdata/blist.hosts.src\"\nIP:\t \"10.10.10.10\"\nLtype:\t \"file\"\nName:\t \"/tasty\"\nnType:\t \"host\"\nPrefix:\t \"\"\nType:\t \"hosts\"\nURL:\t \"\"\n]"

	excRootContent = "server=/122.2o7.net/#\nserver=/1e100.net/#\nserver=/adobedtm.com/#\nserver=/akamai.net/#\nserver=/amazon.com/#\nserver=/amazonaws.com/#\nserver=/apple.com/#\nserver=/ask.com/#\nserver=/avast.com/#\nserver=/bitdefender.com/#\nserver=/cdn.visiblemeasures.com/#\nserver=/cloudfront.net/#\nserver=/coremetrics.com/#\nserver=/edgesuite.net/#\nserver=/freedns.afraid.org/#\nserver=/github.com/#\nserver=/githubusercontent.com/#\nserver=/google.com/#\nserver=/googleadservices.com/#\nserver=/googleapis.com/#\nserver=/googleusercontent.com/#\nserver=/gstatic.com/#\nserver=/gvt1.com/#\nserver=/gvt1.net/#\nserver=/hb.disney.go.com/#\nserver=/hp.com/#\nserver=/hulu.com/#\nserver=/images-amazon.com/#\nserver=/jumptap.com/#\nserver=/msdn.com/#\nserver=/paypal.com/#\nserver=/rackcdn.com/#\nserver=/schema.org/#\nserver=/skype.com/#\nserver=/smacargo.com/#\nserver=/sourceforge.net/#\nserver=/ssl-on9.com/#\nserver=/ssl-on9.net/#\nserver=/static.chartbeat.com/#\nserver=/storage.googleapis.com/#\nserver=/usemaxserver.de/#\nserver=/windows.net/#\nserver=/yimg.com/#\nserver=/ytimg.com/#"

	testCfg = `blocklist {
	disabled false
	dns-redirect-ip 0.0.0.0
	domains {
			include adtechus.net
			include advertising.com
			include centade.com
			include doubleclick.net
			include intellitxt.com
			include patoghee.in
	}
	exclude ytimg.com
	hosts {
		dns-redirect-ip 192.168.168.1
		include beap.gemini.yahoo.com
		source tasty {
			description "File source"
			dns-redirect-ip 10.10.10.10
			file /:~/=../../internal/testdata/blist.hosts.src
		}
	}
}`
)
