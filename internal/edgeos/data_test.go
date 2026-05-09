package edgeos

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jonmeacham/edgeos-adblock/internal/tdata"
)

// logIt writes to io.Writer
func logIt(w io.Writer, s string) {
	_, err := io.Copy(w, strings.NewReader(s))
	if err != nil {
		panic(err)
	}
}

func shuffleArray(slice []string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(slice) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func TestDiffArray(t *testing.T) {
	biggest := sort.StringSlice{"one", "two", "three", "four", "five", "six"}
	smallest := sort.StringSlice{"one", "two", "three"}
	exp := sort.StringSlice{"five", "four", "six"}

	if !reflect.DeepEqual(diffArray(biggest, smallest), exp) {
		t.Errorf("diffArray(biggest, smallest): got %#v, want %#v", diffArray(biggest, smallest), exp)
	}
	if !reflect.DeepEqual(diffArray(smallest, biggest), exp) {
		t.Errorf("diffArray(smallest, biggest): got %#v, want %#v", diffArray(smallest, biggest), exp)
	}

	shuffleArray(biggest)
	if !reflect.DeepEqual(diffArray(smallest, biggest), exp) {
		t.Errorf("after shuffle biggest: got %#v, want %#v", diffArray(smallest, biggest), exp)
	}

	shuffleArray(smallest)
	if !reflect.DeepEqual(diffArray(smallest, biggest), exp) {
		t.Errorf("after shuffle smallest: got %#v, want %#v", diffArray(smallest, biggest), exp)
	}
}

func TestFormatData(t *testing.T) {
	c := NewConfig(
		Dir("/tmp"),
		Ext("edgeos-adblock.conf"),
		Prefix("address=", "server="),
	)

	if err := c.Blacklist(&CFGstatic{Cfg: tdata.Cfg}); err != nil {
		t.Fatal(err)
	}

	for _, node := range c.sortKeys() {
		var (
			actList = &list{RWMutex: &sync.RWMutex{}, entry: make(entry)}

			o = &source{
				ip: c.tree[node].ip,
				Env: &Env{
					Pfx: dnsPfx{
						domain: "address=",
						host:   "server=",
					},
				},
				nType: domn,
			}
			fmttr    = getDnsmasqPrefix(o)
			expBytes []byte
			lines    []string
		)

		r := func() io.Reader {
			sort.Strings(c.tree[node].inc)
			return strings.NewReader(strings.Join(c.tree[node].inc, "\n"))
		}

		b := bufio.NewScanner(r())
		for b.Scan() {
			k := b.Text()
			lines = append(lines, fmt.Sprintf(fmttr, k)+"\n")
			actList.set([]byte(k))
		}

		sort.Strings(lines)
		expBytes = []byte(strings.Join(lines, ""))
		actBytes, err := io.ReadAll(formatData(fmttr, actList))

		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actBytes, expBytes) {
			t.Errorf("node %q: formatData bytes mismatch", node)
		}
	}
}

func TestGetType(t *testing.T) {
	tests := []struct {
		ntypestr string
		typeint  ntype
		typestr  string
	}{
		{typeint: 100, typestr: notknown, ntypestr: "ntype(100)"},
		{typeint: domn, typestr: domains, ntypestr: "domn"},
		{typeint: excDomn, typestr: ExcDomns, ntypestr: "excDomn"},
		{typeint: excHost, typestr: ExcHosts, ntypestr: "excHost"},
		{typeint: excRoot, typestr: ExcRoots, ntypestr: "excRoot"},
		{typeint: host, typestr: hosts, ntypestr: "host"},
		{typeint: preDomn, typestr: PreDomns, ntypestr: "preDomn"},
		{typeint: preHost, typestr: PreHosts, ntypestr: "preHost"},
		{typeint: root, typestr: rootNode, ntypestr: "root"},
		{typeint: unknown, typestr: notknown, ntypestr: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.ntypestr, func(t *testing.T) {
			if tt.typeint != 100 {
				if got, want := typeStr(tt.typestr), tt.typeint; got != want {
					t.Errorf("typeStr: got %v, want %v", got, want)
				}
				if got, want := typeInt(tt.typeint), tt.typestr; got != want {
					t.Errorf("typeInt: got %q, want %q", got, want)
				}
				if got, want := getType(tt.typeint), tt.typestr; got != want {
					t.Errorf("getType(ntype): got %q, want %q", got, want)
				}
				if got, want := getType(tt.typestr), tt.typeint; got != want {
					t.Errorf("getType(string): got %v, want %v", got, want)
				}
			}
			if got, want := fmt.Sprint(tt.typeint), tt.ntypestr; got != want {
				t.Errorf("Sprint(typeint): got %q, want %q", got, want)
			}
		})
	}
}

func TestNewWriter(t *testing.T) {
	tests := []struct {
		name   string
		exp    io.Writer
		expStr string
	}{
		{
			name: "vanilla",
			exp: func() io.Writer {
				var b bytes.Buffer
				return bufio.NewWriter(&b)
			}(),
			expStr: "Es ist Krieg!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			act := NewWriter()
			if !reflect.DeepEqual(act, tt.exp) {
				t.Errorf("NewWriter(): writer mismatch")
			}
			logIt(act, tt.expStr)
			var b bytes.Buffer
			want := bufio.NewWriter(&b)
			if _, err := io.Copy(want, strings.NewReader(tt.expStr)); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(act, want) {
				t.Errorf("after logIt: writer mismatch")
			}
		})
	}
}
