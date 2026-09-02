package edgeos

import "testing"

func TestSourceFilename(t *testing.T) {
	s := &source{
		Env:  NewConfig(Dir("/tmp"), Ext("edgeos-adblock.conf")).Env,
		kind: sourceKind,
		name: "example",
	}
	if got, want := s.filename(), "/tmp/source.example.edgeos-adblock.conf"; got != want {
		t.Fatalf("filename() = %q, want %q", got, want)
	}
}
