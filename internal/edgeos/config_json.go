package edgeos

import (
	"encoding/json"
	"fmt"
)

type jsonNodeSection struct {
	Disabled string                      `json:"disabled"`
	IP       string                      `json:"ip,omitempty"`
	Excludes []string                    `json:"excludes"`
	Includes []string                    `json:"includes"`
	Sources  []map[string]jsonSourceBody `json:"sources"`
}

type jsonSourceBody struct {
	Disabled    string `json:"disabled"`
	Description string `json:"description,omitempty"`
	IP          string `json:"ip,omitempty"`
	Prefix      string `json:"prefix,omitempty"`
	File        string `json:"file,omitempty"`
	URL         string `json:"url,omitempty"`
}

type jsonConfigDoc struct {
	Nodes []map[string]jsonNodeSection `json:"nodes"`
}

// String returns pretty-printed JSON for the blocklist configuration tree.
func (c *Config) String() string {
	b, err := json.MarshalIndent(jsonConfigDoc{Nodes: []map[string]jsonNodeSection{c.nodeSectionMap()}}, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": %q}`, err.Error())
	}
	return string(b)
}

func (c *Config) nodeSectionMap() map[string]jsonNodeSection {
	m := make(map[string]jsonNodeSection, len(c.tree))
	for _, pkey := range c.sortKeys() {
		node := c.tree[pkey]
		m[pkey] = jsonNodeSection{
			Disabled: booltoStr(node.disabled),
			IP:       node.ip,
			Excludes: node.exc,
			Includes: node.inc,
			Sources:  c.jsonSourceSlots(pkey),
		}
	}
	return m
}

func (c *Config) jsonSourceSlots(pkey string) []map[string]jsonSourceBody {
	srcs := c.tree[pkey].src
	if len(srcs) == 0 {
		return []map[string]jsonSourceBody{{}}
	}
	out := make([]map[string]jsonSourceBody, 0, len(srcs))
	for _, o := range srcs {
		body := jsonSourceBody{Disabled: booltoStr(o.disabled)}
		if o.desc != "" {
			body.Description = o.desc
		}
		if o.ip != "" {
			body.IP = o.ip
		}
		if o.prefix != "" {
			body.Prefix = o.prefix
		}
		if o.file != "" {
			body.File = o.file
		}
		if o.url != "" {
			body.URL = o.url
		}
		out = append(out, map[string]jsonSourceBody{o.name: body})
	}
	return out
}
