package edgeos

import (
	"testing"

	"github.com/jonmeacham/edgeos-adblock/internal/tdata"
)

func TestConfigString(t *testing.T) {
	t.Run("full cfg", func(t *testing.T) {
		c := NewConfig(
			Dir("/tmp"),
			Ext("edgeos-adblock.conf"),
			Method("GET"),
		)

		if err := c.Blocklist(&CFGstatic{Cfg: tdata.Cfg}); err != nil {
			t.Fatal(err)
		}
		if got, want := c.String(), tdata.JSONcfg; got != want {
			t.Errorf("String(): got %q, want %q", got, want)
		}
	})

	t.Run("zero host sources", func(t *testing.T) {
		c := NewConfig(
			Dir("/tmp"),
			Ext("edgeos-adblock.conf"),
			Method("GET"),
		)

		if err := c.Blocklist(&CFGstatic{Cfg: tdata.ZeroHostSourcesCfg}); err != nil {
			t.Fatal(err)
		}
		if got, want := c.String(), tdata.JSONcfgZeroHostSources; got != want {
			t.Errorf("String(): got %q, want %q", got, want)
		}
	})
}
