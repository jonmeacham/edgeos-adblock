package dnsmasq

import (
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"testing"
)

func TestConfigFile(t *testing.T) {
	t.Run("dnsmasq entries loaded from files", func(t *testing.T) {
		var (
			b     []byte
			dir   = "../testdata/etc/dnsmasq.d/"
			err   error
			files []string
			r     io.Reader
		)

		files, err = filepath.Glob(dir + "*.conf")
		if err != nil {
			t.Fatal(err)
		}

		for _, f := range files {
			t.Run(f, func(t *testing.T) {
				r, err = ConfigFile(f)
				if err != nil {
					t.Fatalf("cannot open configuration file %s: %v", f, err)
				}

				b, err = io.ReadAll(r)
				if err != nil {
					t.Fatal(err)
				}
				c := make(Conf)
				ip := "0.0.0.0"
				if err := c.Parse(&Mapping{Contents: b}); err != nil {
					t.Fatal(err)
				}

				for k := range c {
					if !c.Redirect(k, ip) {
						t.Errorf("Redirect(%q, %q): expected true", k, ip)
					}
				}
			})
		}
	})

	t.Run("misdirected dnsmasq address entry", func(t *testing.T) {
		c := make(Conf)
		ip := "0.0.0.0"
		k := "address=/www.google.com/0.0.0.0"

		if err := c.Parse(&Mapping{Contents: []byte(k)}); err != nil {
			t.Fatal(err)
		}
		if c.Redirect(k, ip) {
			t.Error("Redirect: expected false for misdirected entry")
		}
	})
}

func BenchmarkFetchHost(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fetchHost("www.microsoft.com", "0.0.0.0")
	}
}

func TestFetchHost(t *testing.T) {
	tests := []struct {
		conf Conf
		exp  bool
		ip   string
		key  string
		name string
	}{
		{
			ip:   "0.0.0.0",
			key:  "badguy_s.com",
			conf: Conf{"badguys.com": Host{IP: "0.0.0.0", Server: false}},
			exp:  false,
			name: "badguys.com",
		},
		{
			ip:   "127.0.0.1",
			key:  "localhoster",
			conf: Conf{"localhost": Host{IP: "127.0.0.1", Server: false}},
			exp:  false,
			name: "localhoster",
		},
		{
			ip:   "127.0.0.1",
			key:  "localhost",
			conf: Conf{"localhost": Host{IP: "#", Server: true}},
			exp:  true,
			name: "localServer",
		},
		{
			ip:   "127.0.0.1",
			key:  "localhost",
			conf: Conf{"localhost": Host{IP: "127.0.0.1", Server: false}},
			exp:  true,
			name: "localhost",
		},
		{
			ip:   "127.0.0.1",
			exp:  false,
			name: "no name",
		},
		{
			ip:   "::1",
			key:  "localhost",
			conf: Conf{"localhost": Host{IP: "127.0.0.1", Server: false}},
			exp:  true,
			name: "localhost IPv6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := fetchHost(tt.key, tt.ip), tt.exp; got != want {
				t.Errorf("fetchHost: got %v, want %v", got, want)
			}
			if got, want := tt.conf.Redirect(tt.key, tt.ip), tt.exp; got != want {
				t.Errorf("Redirect: got %v, want %v", got, want)
			}
		})
	}
}

func TestMatchIP(t *testing.T) {
	tests := []struct {
		exp  bool
		ip   string
		ips  []string
		name string
	}{
		{name: "Fail with IPv4", exp: false, ip: "0.0.0.0", ips: []string{"192.150.200.1", "72.65.23.17", "204.78.13.40"}},
		{name: "Fail with IPv6", exp: false, ip: "0.0.0.0", ips: []string{"0.0.0.0", "0.0.0.0", "fe80::7a8a:20ff:fe44:390d"}},
		{name: "Loopback and unspecified", exp: false, ip: "0.0.0.0", ips: []string{"0.0.0.0", "127.0.0.1", "0.0.0.0"}},
		{name: "Normal specified", exp: true, ip: "192.167.2.2", ips: []string{"192.167.2.2", "192.167.2.2", "192.167.2.2"}},
		{name: "Normal unspecified", exp: true, ip: "0.0.0.0", ips: []string{"0.0.0.0", "0.0.0.0", "0.0.0.0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := matchIP(tt.ip, tt.ips), tt.exp; got != want {
				t.Errorf("matchIP: got %v, want %v", got, want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		act string
		Host
		err    error
		exp    string
		name   string
		reader Mapping
	}{
		{
			Host: Host{
				IP:     "127.0.0.1",
				Server: false,
			},
			act:    `{"badguys.com":{"IP":"0.0.0.0"}}`,
			err:    nil,
			exp:    "127.0.0.1",
			name:   "badguys.com",
			reader: Mapping{Contents: []byte(`address=/badguys.com/0.0.0.0`)},
		},
		{
			Host: Host{
				IP:     "127.0.0.1",
				Server: true,
			},
			act:    `{"xrated.com":{"IP":"0.0.0.0","Server":true}}`,
			err:    nil,
			exp:    "127.0.0.1",
			name:   "xrated.com",
			reader: Mapping{Contents: []byte(`server=/xrated.com/0.0.0.0`)},
		},
		{
			act:  `{}`,
			err:  errors.New("no dnsmasq configuration mapping entries found"),
			exp:  "127.0.0.1",
			name: "No dnsmasq entry",
			reader: Mapping{Contents: []byte(`# All files in this directory will be read by dnsmasq as 
# configuration files, except if their names end in 
# ".dpkg-dist",".dpkg-old" or ".dpkg-new"
#
# This can be changed by editing /etc/default/dnsmasq`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := make(Conf)
			err := c.Parse(&tt.reader)
			if tt.err != nil {
				if err == nil {
					t.Fatal("Parse: expected error")
				}
				if err.Error() != tt.err.Error() {
					t.Fatalf("Parse error: got %q, want %q", err.Error(), tt.err.Error())
				}
			} else if err != nil {
				t.Fatal(err)
			}
			j, err := json.Marshal(c)
			if err != nil {
				t.Fatal(err)
			}
			if string(j) != tt.act {
				t.Errorf("Marshal: got %s, want %s", string(j), tt.act)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		conf Conf
		exp  string
	}{
		{
			conf: Conf{"badguys.com": Host{IP: "0.0.0.0", Server: false}},
			exp:  `{"badguys.com":{"IP":"0.0.0.0"}}`,
		},
		{
			exp: `null`,
		},
	}

	for _, tt := range tests {
		if got, want := tt.conf.String(), tt.exp; got != want {
			t.Errorf("String(): got %s, want %s", got, want)
		}
	}
}
