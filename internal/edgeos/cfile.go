package edgeos

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// CFile holds an array of file names
type CFile struct {
	*Env
	Names []string
}

// readDir returns a listing of dnsmasq adblock configuration files
func (c *CFile) readDir(pattern string) ([]string, error) {
	f, err := filepath.Glob(pattern)
	c.Debug(fmt.Sprintf("Files: %v\n: %v", pattern, f))
	return f, err
}

// Remove deletes a CFile array of file names
func (c *CFile) Remove() error {
	d, err := c.readDir(filepath.Join(c.Dir, "*."+c.Ext))
	if err != nil {
		return err
	}
	keep := make(map[string]struct{}, len(c.Names))
	for _, name := range c.Names {
		keep[name] = struct{}{}
	}
	stale := make([]string, 0, len(d))
	for _, name := range d {
		if _, exists := keep[name]; !exists {
			stale = append(stale, name)
		}
	}
	c.Debug(fmt.Sprintf("Removing: %v", stale))
	return purgeFiles(stale)
}

// String implements string method
func (c *CFile) String() string {
	return strings.Join(c.Strings(), "\n")
}

// Strings returns a sorted array of strings.
func (c *CFile) Strings() []string {
	sort.Strings(c.Names)
	return c.Names
}
