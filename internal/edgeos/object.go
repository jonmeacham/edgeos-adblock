package edgeos

import "sort"

const notfound = -1

// Objects is a collection of configured sources.
type Objects struct {
	*Env
	src []*source
}

// Files returns the dnsmasq fragments expected from this configuration.
func (o *Objects) Files() *CFile {
	files := CFile{Env: o.Env}
	if o.Disabled {
		return &files
	}
	for _, configured := range o.src {
		files.Names = append(files.Names, configured.filename())
	}
	sort.Strings(files.Names)
	return &files
}

// Filter returns sources using the requested file or URL input.
func (o *Objects) Filter(input string) *Objects {
	filtered := &Objects{Env: o.Env}
	for _, configured := range o.src {
		if configured.input == input {
			filtered.src = append(filtered.src, configured)
		}
	}
	return filtered
}

// Find returns the position of a source by name.
func (o *Objects) Find(name string) int {
	for index, configured := range o.src {
		if configured.name == name {
			return index
		}
	}
	return notfound
}

// Names returns sorted source names.
func (o *Objects) Names() (names sort.StringSlice) {
	for _, configured := range o.src {
		names = append(names, configured.name)
	}
	sort.Sort(names)
	return names
}

// Len returns the number of sources.
func (o *Objects) Len() int { return len(o.src) }

func (o *Objects) Less(i, j int) bool { return o.src[i].name < o.src[j].name }
func (o *Objects) Swap(i, j int)      { o.src[i], o.src[j] = o.src[j], o.src[i] }
