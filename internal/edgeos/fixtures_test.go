package edgeos

import (
	"os"
	"testing"
)

func fixtureConfig(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../testdata/config.test.boot")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const (
	fixtureDisabled = `adblock {
		disabled true
		dns-redirect-ip 0.0.0.0
	}`
	fixtureSingleSource = `adblock {
		disabled false
		dns-redirect-ip 0.0.0.0
		source example {
			prefix "0.0.0.0 "
			url https://example.invalid/hosts.txt
		}
	}`
	fixtureNoAdblock = `Configuration under specified path is empty`
)
