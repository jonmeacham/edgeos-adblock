package edgeos

import (
	"testing"
)

func TestChkWeb(t *testing.T) {
	tests := []struct {
		exp  bool
		port int
		site string
	}{
		{exp: true, site: "www.google.com", port: 443},
		{exp: true, site: "yahoo.com", port: 80},
		{exp: true, site: "bing.com", port: 443},
		{exp: false, site: "bigtop.@@@", port: 80},
	}
	for _, tt := range tests {
		t.Run(tt.site, func(t *testing.T) {
			got := ChkWeb(tt.site, tt.port)
			if got != tt.exp {
				t.Errorf("ChkWeb(%q, %d): got %v, want %v", tt.site, tt.port, got, tt.exp)
			}
		})
	}
}
