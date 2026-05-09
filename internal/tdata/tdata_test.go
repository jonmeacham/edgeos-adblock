package tdata_test

import (
	"testing"

	"github.com/jonmeacham/edgeos-adblock/internal/tdata"
)

func TestTdataCfg(t *testing.T) {
	exp := map[string]string{
		"cfg":          tdata.Cfg,
		"cfg2":         tdata.CfgPartial,
		"cfg3":         tdata.CfgMimimal,
		"none":         tdata.CfgDeleted,
		"fileManifest": tdata.FileManifest,
		"default":      "",
	}

	for k := range exp {
		t.Run(k, func(t *testing.T) {
			act := tdata.Get(k)
			if act != exp[k] {
				t.Errorf("Get(%q): got %q, want %q", k, act, exp[k])
			}
		})
	}
}
