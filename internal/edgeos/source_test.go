package edgeos

import (
	"testing"
)

func TestArea(t *testing.T) {
	tests := []struct {
		exp  string
		name string
		s    *source
	}{
		{
			name: roots,
			s: &source{
				nType: root,
			},
			exp: roots,
		},
		{
			name: PreDomns,
			s: &source{
				nType: preDomn,
			},
			exp: domains,
		},
		{
			name: hosts,
			s: &source{
				nType: host,
			},
			exp: hosts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, want := tt.s.area(), tt.exp; got != want {
				t.Errorf("area(): got %q, want %q", got, want)
			}
		})
	}
}
