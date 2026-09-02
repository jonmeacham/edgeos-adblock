package edgeos

import "io"

// IFace identifies a configured content group.
type IFace int

const (
	Invalid IFace = iota + 100
	AllowObj
	BlockObj
	FileObj
	URLObj
)

func (i IFace) String() string {
	switch i {
	case AllowObj:
		return allowNode
	case BlockObj:
		return blockNode
	case FileObj:
		return fileNode
	case URLObj:
		return urlNode
	default:
		return "unknown"
	}
}

// Contenter resolves one configured group into sources ready for processing.
type Contenter interface {
	GetList() *Objects
}

type content struct {
	*Objects
	kind IFace
}

type bList struct {
	file string
	r    io.Reader
	size int
}

func (c *content) GetList() *Objects {
	switch c.kind {
	case FileObj:
		c.loadFiles()
	case URLObj:
		c.downloadURLs()
	}
	return c.Objects
}

func (c *content) loadFiles() {
	responses := make(chan *source, len(c.src))
	for _, configured := range c.src {
		configured.Env = c.Env
		go func(configured *source) {
			configured.r, configured.err = GetFile(configured.file)
			responses <- configured
		}(configured)
	}
	c.collect(responses)
}

func (c *content) downloadURLs() {
	responses := make(chan *source, len(c.src))
	for _, configured := range c.src {
		configured.Env = c.Env
		go func(configured *source) {
			responses <- download(configured)
		}(configured)
	}
	c.collect(responses)
}

func (c *content) collect(responses chan *source) {
	for range c.src {
		response := <-responses
		if index := c.Find(response.name); index != notfound {
			c.src[index] = response
		}
	}
	close(responses)
}
