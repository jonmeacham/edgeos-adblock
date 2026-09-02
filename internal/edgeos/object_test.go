package edgeos

import (
	"reflect"
	"sort"
	"testing"
)

func TestObjectsFindSortAndFilter(t *testing.T) {
	objects := &Objects{
		src: []*source{
			{name: "remote", input: urlNode},
			{name: "local", input: fileNode},
		},
	}
	if got := objects.Find("remote"); got != 0 {
		t.Fatalf("Find(remote) = %d, want 0", got)
	}
	if got := objects.Find("missing"); got != notfound {
		t.Fatalf("Find(missing) = %d, want %d", got, notfound)
	}
	wantFiles := sort.StringSlice{"local"}
	if got := objects.Filter(fileNode).Names(); !reflect.DeepEqual(got, wantFiles) {
		t.Fatalf("file sources = %#v, want %#v", got, wantFiles)
	}
	sort.Sort(objects)
	wantNames := sort.StringSlice{"local", "remote"}
	if got := objects.Names(); !reflect.DeepEqual(got, wantNames) {
		t.Fatalf("sorted names = %#v, want %#v", got, wantNames)
	}
}

func TestConfiguredFragmentNames(t *testing.T) {
	c := NewConfig(Dir("/tmp"), Ext("edgeos-adblock.conf"))
	if err := c.Adblock(&CFGstatic{Cfg: fixtureConfig(t)}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/tmp/block.explicit.edgeos-adblock.conf",
		"/tmp/source.hageziPro.edgeos-adblock.conf",
		"/tmp/source.tasty.edgeos-adblock.conf",
	}
	if got := c.Files().Strings(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Files() = %#v, want %#v", got, want)
	}
}
