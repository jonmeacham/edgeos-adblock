package dnsmasq

import (
	"bytes"
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const fixtureConfGlob = "../testdata/etc/dnsmasq.d/*.conf"

// maxDomainsToCheck is how many distinct blocked names we query per run (random subset of the fixture set).
const maxDomainsToCheck = 5

// mergeFixtureAddressLines writes all address=/name/0.0.0.0 lines from fixture dnsmasq includes into dst.
func mergeFixtureAddressLines(dst string) error {
	matches, err := filepath.Glob(fixtureConfGlob)
	if err != nil {
		return err
	}
	var b strings.Builder
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.HasPrefix(line, "address=") && strings.HasSuffix(line, "/0.0.0.0") {
				b.WriteString(line)
				b.WriteByte('\n')
			}
		}
	}
	if b.Len() == 0 {
		return fmt.Errorf("no address=/…/0.0.0.0 lines found under %s", fixtureConfGlob)
	}
	return os.WriteFile(dst, []byte(b.String()), 0o644)
}

func fixtureAddressDomains(t *testing.T) []string {
	t.Helper()
	matches, err := filepath.Glob(fixtureConfGlob)
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]struct{})
	var doms []string
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "address=") || !strings.HasSuffix(line, "/0.0.0.0") {
				continue
			}
			// address=/fqdn/0.0.0.0
			parts := strings.Split(line, "/")
			if len(parts) < 3 {
				continue
			}
			d := parts[1]
			if d == "" {
				continue
			}
			if _, ok := seen[d]; ok {
				continue
			}
			seen[d] = struct{}{}
			doms = append(doms, d)
		}
	}
	if len(doms) == 0 {
		t.Fatal("fixture conf: no address= domains")
	}
	return doms
}

func pickRandomSubset(domains []string, n int, rng *rand.Rand) []string {
	if len(domains) <= n {
		out := make([]string, len(domains))
		copy(out, domains)
		return out
	}
	shuffled := make([]string, len(domains))
	copy(shuffled, domains)
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:n]
}

func writeMainConf(path string, port int, includeAbs string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "no-resolv\n")
	fmt.Fprintf(&b, "no-hosts\n")
	fmt.Fprintf(&b, "listen-address=127.0.0.1\n")
	fmt.Fprintf(&b, "port=%d\n", port)
	fmt.Fprintf(&b, "bind-interfaces\n")
	fmt.Fprintf(&b, "cache-size=0\n")
	fmt.Fprintf(&b, "conf-file=%s\n", includeAbs)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func udpResolverFor(port int) *net.Resolver {
	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "udp", fmt.Sprintf("127.0.0.1:%d", port))
		},
	}
}

func waitUntilDNSAnswers(t *testing.T, r *net.Resolver, domain string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var lastErr error
	for {
		if ctx.Err() != nil {
			t.Fatalf("dnsmasq did not answer for %q in time: %v", domain, lastErr)
		}
		_, err := r.LookupIPAddr(ctx, domain)
		if err == nil {
			return
		}
		lastErr = err
		time.Sleep(50 * time.Millisecond)
	}
}

func hasBlockedIP(addrs []net.IPAddr) bool {
	for _, a := range addrs {
		if a.IP.Equal(net.IPv4zero) || a.IP.IsUnspecified() {
			return true
		}
	}
	return false
}

// TestDnsmasqServesBlockedLookups starts a real dnsmasq with fixture address= entries and checks that a
// random subset resolves to 0.0.0.0 via the local instance. Skips when dnsmasq is not installed (e.g. dev image).
func TestDnsmasqServesBlockedLookups(t *testing.T) {
	if _, err := exec.LookPath("dnsmasq"); err != nil {
		t.Skip("dnsmasq not in PATH (install in E2E image or on host to run this test)")
	}

	domains := fixtureAddressDomains(t)
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(os.Getpid())))
	nPick := maxDomainsToCheck
	if len(domains) < nPick {
		nPick = len(domains)
	}
	picked := pickRandomSubset(domains, nPick, rng)
	t.Logf("checking %d domains: %v", len(picked), picked)

	l, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.LocalAddr().(*net.UDPAddr).Port
	_ = l.Close()

	tmp := t.TempDir()
	includePath := filepath.Join(tmp, "blocked.conf")
	if err := mergeFixtureAddressLines(includePath); err != nil {
		t.Fatal(err)
	}
	includeAbs, err := filepath.Abs(includePath)
	if err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(tmp, "dnsmasq.conf")
	if err := writeMainConf(mainPath, port, includeAbs); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	cmd := exec.Command("dnsmasq", "-k", "-C", mainPath)
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start dnsmasq: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	})

	r := udpResolverFor(port)
	waitUntilDNSAnswers(t, r, picked[0])

	ctx := context.Background()
	for _, d := range picked {
		addrs, err := r.LookupIPAddr(ctx, d)
		if err != nil {
			t.Fatalf("LookupIPAddr(%q): %v\nstderr:\n%s", d, err, stderr.String())
		}
		if !hasBlockedIP(addrs) {
			t.Errorf("domain %q: got %v; want an answer including 0.0.0.0 (or unspecified)", d, addrs)
		}
	}
}

// TestFixtureAddressDomainsParity ensures fixture parsing for TestDnsmasqServesBlockedLookups stays in sync
// with the dnsmasq config parser for address= lines.
func TestFixtureAddressDomainsParity(t *testing.T) {
	doms := fixtureAddressDomains(t)
	main := filepath.Join(t.TempDir(), "all.conf")
	if err := mergeFixtureAddressLines(main); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(main)
	if err != nil {
		t.Fatal(err)
	}
	c := make(Conf)
	if err := c.Parse(&Mapping{Contents: data}); err != nil {
		t.Fatal(err)
	}
	for _, d := range doms {
		if _, ok := c[d]; !ok {
			t.Errorf("domain %q from line scan missing in Conf.Parse", d)
		}
	}
}
