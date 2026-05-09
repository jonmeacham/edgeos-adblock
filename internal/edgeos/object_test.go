package edgeos

import (
	"reflect"
	"sort"
	"testing"

	"github.com/jonmeacham/edgeos-adblock/internal/tdata"
)

func TestObjectsAddObj(t *testing.T) {
	c := NewConfig(
		Dir("/tmp"),
		Ext("edgeos-adblock.conf"),
	)

	if err := c.Blacklist(&CFGstatic{Cfg: tdata.Cfg}); err != nil {
		t.Fatal(err)
	}

	o, err := c.NewContent(FileObj)
	if err != nil {
		t.Fatal(err)
	}

	exp := o

	o.GetList().addObj(c, rootNode)

	if !reflect.DeepEqual(o, exp) {
		t.Errorf("addObj: got %+v, want %+v", o, exp)
	}
}

func TestObjectString(t *testing.T) {
	c := NewConfig(
		Dir("/tmp"),
		Ext("edgeos-adblock.conf"),
	)

	if err := c.Blacklist(&CFGstatic{Cfg: tdata.Cfg}); err != nil {
		t.Fatal(err)
	}

	act := c.GetAll()
	if act.Find("hageziPro") < 0 {
		t.Errorf("Find(hageziPro): expected configured URL source")
	}
	if got, want := act.Find("@#$%"), -1; got != want {
		t.Errorf("Find: got %d, want %d", got, want)
	}
}

func TestSortObject(t *testing.T) {
	act := &Objects{
		src: []*source{
			{name: "eagle"},
			{name: "aardvark"},
			{name: "dog"},
			{name: "crab"},
			{name: "beetle"},
		},
	}

	exp := &Objects{
		src: []*source{
			{name: "aardvark"},
			{name: "beetle"},
			{name: "crab"},
			{name: "dog"},
			{name: "eagle"},
		},
	}

	sort.Sort(act)
	if !reflect.DeepEqual(act, exp) {
		t.Errorf("sort: got %+v, want %+v", act, exp)
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		ltype string
		exp   sort.StringSlice
	}{
		{ltype: urls, exp: urlsOnly},
		{ltype: files, exp: filesOnly},
		{ltype: hosts, exp: sort.StringSlice(nil)},
	}

	c := NewConfig(
		Dir("/tmp"),
		Ext("edgeos-adblock.conf"),
	)

	if err := c.Blacklist(&CFGstatic{Cfg: tdata.Cfg}); err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.ltype, func(t *testing.T) {
			act := c.GetAll().Filter(tt.ltype)
			if !reflect.DeepEqual(act.Names(), tt.exp) {
				t.Errorf("Names(): got %#v, want %#v", act.Names(), tt.exp)
			}
		})
	}
}

func TestGetLtypeDesc(t *testing.T) {
	if got, want := getLtypeDesc(""), "pre-configured unknown ltype"; got != want {
		t.Errorf("getLtypeDesc(\"\"): got %q, want %q", got, want)
	}
	if got, want := getLtypeDesc("Hyperbolic-frisbee-throwers"), "pre-configured Hyperbolic frisbee throwers"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

var (
	filesOnly = sort.StringSlice{"tasty"}
	urlsOnly  = sort.StringSlice{"hageziPro"}
)
