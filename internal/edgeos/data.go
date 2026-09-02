package edgeos

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// ntype identifies normalized content behavior.
type ntype int

const (
	unknownKind ntype = iota
	allowKind
	blockKind
	sourceKind
)

func (n ntype) String() string {
	switch n {
	case allowKind:
		return allowNode
	case blockKind:
		return blockNode
	case sourceKind:
		return sourceNode
	default:
		return "unknown"
	}
}

// formatData renders sorted dnsmasq rules.
func formatData(format string, values *list) io.Reader {
	lines := make(sort.StringSlice, 0, len(values.entry))
	values.RLock()
	for value := range values.entry {
		lines = append(lines, fmt.Sprintf(format+"\n", value))
	}
	values.RUnlock()
	lines.Sort()
	return strings.NewReader(strings.Join(lines, ""))
}

// Iter retains the compact range helper used by tests and existing callers.
func Iter(i int) []struct{} { return make([]struct{}, i) }

func booltoStr(value bool) string { return strconv.FormatBool(value) }

func strToBool(value string) (bool, error) { return strconv.ParseBool(value) }
