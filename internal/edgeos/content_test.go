package edgeos

import (
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestSourceProcessing(t *testing.T) {
	env := NewConfig(
		Dir(t.TempDir()),
		Ext("edgeos-adblock.conf"),
		Prefix("address="),
		SetLogger(newLog()),
	).Env
	env.stat[sourceNode] = &stats{}
	s := &source{
		Env:    env,
		name:   "synthetic",
		kind:   sourceKind,
		ip:     "0.0.0.0",
		prefix: "0.0.0.0 ",
		r: strings.NewReader(strings.Join([]string{
			"0.0.0.0 ADS.EXAMPLE.INVALID",
			"0.0.0.0 metrics.example.invalid",
			"0.0.0.0 metrics.example.invalid",
			"local=/Direct.Example.Invalid/",
			"# ignored",
		}, "\n")),
	}
	b := s.process()
	got, err := io.ReadAll(b.r)
	if err != nil {
		t.Fatal(err)
	}
	want := "address=/ads.example.invalid/0.0.0.0\naddress=/direct.example.invalid/0.0.0.0\naddress=/metrics.example.invalid/0.0.0.0\n"
	if string(got) != want {
		t.Fatalf("processed content:\n%s\nwant:\n%s", got, want)
	}
	if b.size != 3 {
		t.Fatalf("kept = %d, want 3", b.size)
	}
}

func TestAllowPrecedesExplicitAndSourceBlocks(t *testing.T) {
	dir := t.TempDir()
	env := NewConfig(
		Dir(dir),
		Ext("edgeos-adblock.conf"),
		Prefix("address="),
		SetLogger(newLog()),
	).Env

	allow := &source{Env: env, kind: allowKind, name: "explicit", r: strings.NewReader("allow.example.invalid")}
	block := &source{Env: env, kind: blockKind, name: "explicit", ip: "0.0.0.0", r: strings.NewReader("allow.example.invalid\nblocked.example.invalid")}
	remote := &source{Env: env, kind: sourceKind, name: "remote", ip: "0.0.0.0", r: strings.NewReader("sub.allow.example.invalid\nblocked.example.invalid\nsource.example.invalid")}

	for _, item := range []*source{allow, block, remote} {
		env.Lock()
		env.stat[item.kind.String()] = &stats{}
		env.Unlock()
		if err := item.process().writeFile(); err != nil {
			t.Fatal(err)
		}
	}

	blockData, err := os.ReadFile(dir + "/block.explicit.edgeos-adblock.conf")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(blockData), "address=/blocked.example.invalid/0.0.0.0\n"; got != want {
		t.Fatalf("explicit block = %q, want %q", got, want)
	}
	sourceData, err := os.ReadFile(dir + "/source.remote.edgeos-adblock.conf")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(sourceData), "address=/source.example.invalid/0.0.0.0\n"; got != want {
		t.Fatalf("source block = %q, want %q", got, want)
	}
}

func TestDNSMasqInputProcessing(t *testing.T) {
	for _, line := range []string{
		"local=/ads.example.invalid/",
		"address=/ads.example.invalid/0.0.0.0",
	} {
		got, ok := dnsmasqBlockedDomain([]byte(line))
		if !ok || string(got) != "ads.example.invalid" {
			t.Fatalf("dnsmasqBlockedDomain(%q) = %q, %v", line, got, ok)
		}
	}
}

func TestProcessContentRequiresInput(t *testing.T) {
	if err := NewConfig().ProcessContent(); err == nil {
		t.Fatal("ProcessContent() succeeded without content")
	}
}

func TestFileContent(t *testing.T) {
	env := NewConfig().Env
	s := &source{Env: env, name: "file", file: "/does/not/exist"}
	objects := &content{kind: FileObj, Objects: &Objects{
		Env: env,
		src: []*source{s},
	}}
	list := objects.GetList()
	if list.Len() != 1 || list.src[0].err == nil {
		t.Fatal("missing source file did not return an error")
	}
	objects.src[0].err = nil
	if err := NewConfig().ProcessContent(objects); err == nil {
		t.Fatal("processing a missing source file succeeded")
	}
}

func TestListMergeIsConcurrentSafe(t *testing.T) {
	l := &list{RWMutex: &sync.RWMutex{}, entry: make(entry)}
	var wg sync.WaitGroup
	for _, value := range []string{"one.example", "two.example", "one.example"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l.set([]byte(value))
		}()
	}
	wg.Wait()
	if len(l.entry) != 2 {
		t.Fatalf("entries = %d, want 2", len(l.entry))
	}
}
