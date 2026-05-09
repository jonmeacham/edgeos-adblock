package edgeos

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"

	"github.com/jonmeacham/edgeos-adblock/internal/tdata"
)

func TestKeys(t *testing.T) {
	var keys sort.StringSlice
	c := NewConfig()
	if err := c.Blocklist(&CFGstatic{Cfg: tdata.Cfg}); err != nil {
		t.Fatal(err)
	}

	wantKeys := sort.StringSlice{"blocklist", "domains", "hosts"}
	if !reflect.DeepEqual(c.sortKeys(), wantKeys) {
		t.Errorf("sortKeys(): got %#v, want %#v", c.sortKeys(), wantKeys)
	}

	wantNames := sort.StringSlice{"blocklisted-servers", "blocklisted-subdomains", "global-blocklisted-domains", "hageziPro", "tasty"}
	if !reflect.DeepEqual(c.GetAll().Names(), wantNames) {
		t.Errorf("Names(): got %#v, want %#v", c.GetAll().Names(), wantNames)
	}

	for _, k := range []string{"a", "b", "c", "z", "q", "s", "e", "i", "x", "m"} {
		keys = append(keys, k)
	}

	if keys.Len() != 10 {
		t.Errorf("keys.Len(): got %d, want 10", keys.Len())
	}
}

func TestKeyExists(t *testing.T) {
	exp := list{
		RWMutex: &sync.RWMutex{},
		entry: entry{
			"five.six.intellitxt.com":                        struct{}{},
			"four.five.six.intellitxt.com":                   struct{}{},
			"intellitxt.com":                                 struct{}{},
			"one.two.three.four.five.six.intellitxt.com":     struct{}{},
			"six.intellitxt.com":                             struct{}{},
			"three.four.five.six.intellitxt.com":             struct{}{},
			"top.one.two.three.four.five.six.intellitxt.com": struct{}{},
			"two.three.four.five.six.intellitxt.com":         struct{}{},
		},
	}
	for _, k := range keyArray {
		if !exp.keyExists(k) {
			t.Errorf("keyExists(%s): expected true", k)
		}
	}
	if exp.keyExists([]byte("zKeyDoesn'tExist")) {
		t.Error("keyExists(missing): expected false")
	}
}

func TestSubKeyExists(t *testing.T) {
	exp := list{
		RWMutex: &sync.RWMutex{},
		entry: entry{
			"five.six.intellitxt.com":                        struct{}{},
			"four.five.six.intellitxt.com":                   struct{}{},
			"intellitxt.com":                                 struct{}{},
			"one.two.three.four.five.six.intellitxt.com":     struct{}{},
			"six.intellitxt.com":                             struct{}{},
			"three.four.five.six.intellitxt.com":             struct{}{},
			"top.one.two.three.four.five.six.intellitxt.com": struct{}{},
			"two.three.four.five.six.intellitxt.com":         struct{}{},
		},
	}
	for _, k := range keyArray {
		if !exp.subKeyExists(k) {
			t.Errorf("subKeyExists: expected true for %s", k)
		}
	}
	if exp.subKeyExists([]byte("zKeyDoesn'tExist")) {
		t.Error("subKeyExists: expected false for missing key")
	}
	if exp.subKeyExists([]byte("com")) {
		t.Error("subKeyExists(com): expected false")
	}
}

func TestMerge(t *testing.T) {
	testList1 := list{RWMutex: &sync.RWMutex{}, entry: make(entry)}
	testList2 := list{RWMutex: &sync.RWMutex{}, entry: make(entry)}
	exp := list{RWMutex: &sync.RWMutex{}, entry: make(entry)}

	for i := range Iter(20) {
		exp.entry[fmt.Sprint(i)] = struct{}{}
		switch {
		case i%2 == 0:
			testList1.entry[fmt.Sprint(i)] = struct{}{}
		case i%2 != 0:
			testList2.entry[fmt.Sprint(i)] = struct{}{}
		}
	}
	testList1.merge(&testList2)

	if !reflect.DeepEqual(testList1, exp) {
		t.Errorf("merge: got %#v, want %#v", testList1, exp)
	}
}

func TestString(t *testing.T) {
	exp := `"a.applovin.com":{},
"a.glcdn.co":{},
"a.vserv.mobi":{},
"ad.leadboltapps.net":{},
"ad.madvertise.de":{},
"ad.where.com":{},
"ad1.adinfuse.com":{},
"ad2.adinfuse.com":{},
"adcontent.saymedia.com":{},
"adinfuse.com":{},
"admicro1.vcmedia.vn":{},
"admicro2.vcmedia.vn":{},
"admin.vserv.mobi":{},
"ads.adiquity.com":{},
"ads.admarvel.com":{},
"ads.admoda.com":{},
"ads.celtra.com":{},
"ads.flurry.com":{},
"ads.matomymobile.com":{},
"ads.mobgold.com":{},
"ads.mobilityware.com":{},
"ads.mopub.com":{},
`
	if got, want := act.String(), exp; got != want {
		t.Errorf("String(): mismatch\n got:\n%s\n want:\n%s", got, want)
	}
}

var (
	act = list{
		entry: entry{
			"a.applovin.com":         struct{}{},
			"a.glcdn.co":             struct{}{},
			"a.vserv.mobi":           struct{}{},
			"ad.leadboltapps.net":    struct{}{},
			"ad.madvertise.de":       struct{}{},
			"ad.where.com":           struct{}{},
			"ad1.adinfuse.com":       struct{}{},
			"ad2.adinfuse.com":       struct{}{},
			"adcontent.saymedia.com": struct{}{},
			"adinfuse.com":           struct{}{},
			"admicro1.vcmedia.vn":    struct{}{},
			"admicro2.vcmedia.vn":    struct{}{},
			"admin.vserv.mobi":       struct{}{},
			"ads.adiquity.com":       struct{}{},
			"ads.admarvel.com":       struct{}{},
			"ads.admoda.com":         struct{}{},
			"ads.celtra.com":         struct{}{},
			"ads.flurry.com":         struct{}{},
			"ads.matomymobile.com":   struct{}{},
			"ads.mobgold.com":        struct{}{},
			"ads.mobilityware.com":   struct{}{},
			"ads.mopub.com":          struct{}{},
		},
	}
	keyArray = [][]byte{
		[]byte("top.one.two.three.four.five.six.intellitxt.com"),
		[]byte("one.two.three.four.five.six.intellitxt.com"),
		[]byte("two.three.four.five.six.intellitxt.com"),
		[]byte("three.four.five.six.intellitxt.com"),
		[]byte("four.five.six.intellitxt.com"),
		[]byte("five.six.intellitxt.com"),
		[]byte("six.intellitxt.com"),
		[]byte("intellitxt.com"),
	}
)
