package edgeos

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/jonmeacham/edgeos-adblock/internal/regx"
)

// source is a configured or explicit source of normalized domains.
type source struct {
	*Env
	allow  []string
	block  []string
	desc   string
	err    error
	file   string
	input  string
	ip     string
	kind   ntype
	name   string
	prefix string
	r      io.Reader
	src    []*source
	url    string
}

func newSource() *source {
	return &source{
		allow: []string{},
		block: []string{},
		src:   []*source{},
	}
}

func (s *source) filename() string {
	return filepath.Join(s.Dir, fmt.Sprintf("%s.%s.%s", s.kind.String(), s.name, s.Ext))
}

// dnsmasqBlockedDomain extracts a domain from supported dnsmasq forms.
func dnsmasqBlockedDomain(line []byte) ([]byte, bool) {
	line = bytes.TrimSpace(line)
	switch {
	case bytes.HasPrefix(line, []byte("local=/")):
		rest := bytes.TrimPrefix(line, []byte("local=/"))
		rest = bytes.TrimSuffix(rest, []byte("/"))
		if len(rest) == 0 {
			return nil, false
		}
		return rest, true
	case bytes.HasPrefix(line, []byte("address=/")):
		parts := bytes.Split(line, []byte("/"))
		if len(parts) >= 3 && len(parts[1]) > 0 {
			return parts[1], true
		}
	}
	return nil, false
}

// process normalizes source content and renders a dnsmasq fragment.
func (s *source) process() *bList {
	scanner := bufio.NewScanner(s.r)
	find := regx.NewRegex()
	values := list{RWMutex: &sync.RWMutex{}, entry: make(entry)}
	dropped, extracted, kept := 0, 0, 0

	for scanner.Scan() {
		line := bytes.ToLower(bytes.TrimSpace(scanner.Bytes()))
		if len(line) == 0 || bytes.HasPrefix(line, []byte("#")) ||
			bytes.HasPrefix(line, []byte("//")) || bytes.HasPrefix(line, []byte("<")) {
			continue
		}

		var candidate []byte
		if domain, ok := dnsmasqBlockedDomain(line); ok {
			candidate = domain
		} else {
			if s.prefix != "" && !bytes.HasPrefix(line, []byte(s.prefix)) {
				continue
			}
			stripped, ok := find.StripPrefixAndSuffix(line, s.prefix)
			if !ok {
				continue
			}
			candidate = stripped
		}

		for _, fqdn := range find.RX[regx.FQDN].FindAll(candidate, -1) {
			extracted++
			if s.kind == allowKind {
				if values.addIfAbsent(fqdn) {
					kept++
				} else {
					dropped++
				}
				continue
			}
			if s.Dex.subKeyExists(fqdn) || !s.Exc.addIfAbsent(fqdn) {
				dropped++
				continue
			}
			values.set(fqdn)
			kept++
		}
	}

	if s.kind == allowKind {
		s.Dex.merge(&values)
	}
	s.sum(dropped, extracted, kept)

	result := &bList{file: s.filename(), size: kept}
	if s.kind == allowKind {
		result.size = 0
		result.file = ""
		result.r = strings.NewReader("")
		return result
	}
	result.r = formatData(s.Pfx.domain+"/%v/"+s.ip, &values)
	return result
}

func (s *source) sum(dropped, extracted, kept int) {
	s.Lock()
	counter := s.stat[s.kind.String()]
	s.Unlock()
	atomic.AddInt32(&counter.dropped, int32(dropped))
	atomic.AddInt32(&counter.extracted, int32(extracted))
	atomic.AddInt32(&counter.kept, int32(kept))

	if kept > 0 {
		s.Log.Infof("%s: found: %d", s.name, extracted)
		s.Log.Infof("%s: emitted: %d", s.name, kept)
		s.Log.Infof("%s: dropped: %d", s.name, dropped)
	} else if extracted > 0 && dropped == extracted {
		s.Log.Warningf("%s: 0 records emitted; check the source and allow rules", s.name)
	}
}
