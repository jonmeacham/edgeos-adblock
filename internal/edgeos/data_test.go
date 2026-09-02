package edgeos

import (
	"io"
	"sync"
	"testing"
)

func TestFormatData(t *testing.T) {
	values := &list{RWMutex: &sync.RWMutex{}, entry: make(entry)}
	values.set([]byte("two.example"))
	values.set([]byte("one.example"))
	got, err := io.ReadAll(formatData("address=/%v/0.0.0.0", values))
	if err != nil {
		t.Fatal(err)
	}
	want := "address=/one.example/0.0.0.0\naddress=/two.example/0.0.0.0\n"
	if string(got) != want {
		t.Fatalf("formatData() = %q, want %q", got, want)
	}
}

func TestContentKinds(t *testing.T) {
	for kind, want := range map[ntype]string{
		unknownKind: "unknown",
		allowKind:   allowNode,
		blockKind:   blockNode,
		sourceKind:  sourceNode,
	} {
		if got := kind.String(); got != want {
			t.Errorf("%v.String() = %q, want %q", int(kind), got, want)
		}
	}
}
