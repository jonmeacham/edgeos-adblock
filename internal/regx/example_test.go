package regx_test

import (
	"fmt"
	"testing"

	"github.com/jonmeacham/edgeos-adblock/internal/regx"
)

func TestOBJ(t *testing.T) {
	rxmap := regx.NewRegex()
	act := fmt.Sprint(rxmap)
	exp := `CMNT: ^(?:[\/*]+)(.*?)(?:[*\/]+)$
DESC: ^(?:description)+\s"?([^"]+)?"?$
DSBL: ^(?:disabled)+\s([\S]+)$
FLIP: ^(?:address=[/][.]{0,1}.*[/])(.*)$
FQDN: \b((?:(?:[^.-/]{0,1})[\p{L}\d-_]{1,63}[-]{0,1}[.]{1})+(?:[\p{L}]{2,63}))\b
HOST: ^(?:address=[/][.]{0,1})(.*)(?:[/].*)$
HTTP: (?:^(?:http|https){1}:)(?:\/|%2f){1,2}(.*)
IPBH: ^(?:dns-redirect-ip)+\s([\S]+)$
LBRC: [{]
LEAF: ^([\S]+)+\s([\S]+)\s[{]{1}$
MISC: ^([\w-]+)$
MLTI: ^((?:include|exclude)+)\s([\S]+)$
MPTY: ^$
NAME: ^([\w-]+)\s["']{0,1}(.*?)["']{0,1}$
NODE: ^([\w-]+)\s[{]{1}$
RBRC: [}]
SUFX: (?:#.*|\{.*|[/[].*)\z`
	if act != exp {
		t.Errorf("fmt.Sprint(rxmap): mismatch\n got:\n%s\n want:\n%s", act, exp)
	}
}
