// Package edgeos provides methods and structures to retrieve, parse, and render EdgeOS configuration data and files.
package edgeos

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// ConfLoader handles the live EdgeOS configuration and static test files.
type ConfLoader interface {
	read() io.Reader
}

// Config contains the parsed adblock configuration and runtime environment.
type Config struct {
	*Env
	root *source
}

type ctr struct {
	*sync.RWMutex
	stat
}

type stat map[string]*stats

type stats struct {
	dropped   int32
	extracted int32
	kept      int32
}

const (
	agent        = `curl/7.64.1`
	allowNode    = "allow"
	blockNode    = "block"
	disabledNode = "disabled"
	fileNode     = "file"
	redirectNode = "dns-redirect-ip"
	rootNode     = "adblock"
	sourceNode   = "source"
	urlNode      = "url"

	// False and True are string representations used by EdgeOS.
	False = "false"
	True  = "true"
)

// ErrNoAdblockCfg is returned when no adblock configuration tree was parsed.
var ErrNoAdblockCfg = errors.New("no EdgeOS dns forwarding adblock configuration")

var sourceNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// GetTotalStats returns aggregate source-processing statistics.
func (c *Config) GetTotalStats() (dropped, extracted, kept int32) {
	for _, value := range c.stat {
		dropped += value.dropped
		extracted += value.extracted
		kept += value.kept
	}
	return dropped, extracted, kept
}

// NewContent resolves one configured content group for processing.
func (c *Config) NewContent(iface IFace) (Contenter, error) {
	switch iface {
	case AllowObj:
		return &content{Objects: c.explicit(allowKind), kind: iface}, nil
	case BlockObj:
		return &content{Objects: c.explicit(blockKind), kind: iface}, nil
	case FileObj:
		return &content{Objects: c.GetAll(fileNode), kind: iface}, nil
	case URLObj:
		return &content{Objects: c.GetAll(urlNode), kind: iface}, nil
	default:
		return nil, errors.New("invalid interface requested")
	}
}

func (c *Config) explicit(kind ntype) *Objects {
	objects := &Objects{Env: c.Env}
	if c.root == nil {
		return objects
	}

	values := c.root.allow
	if kind == blockKind {
		values = c.root.block
	}
	if len(values) == 0 {
		return objects
	}

	objects.src = append(objects.src, &source{
		Env:  c.Env,
		kind: kind,
		name: "explicit",
		ip:   c.redirectIP(),
		r:    strings.NewReader(strings.Join(values, "\n")),
	})
	return objects
}

// GetAll returns configured sources, optionally filtered by file or URL input.
func (c *Config) GetAll(input ...string) *Objects {
	objects := &Objects{Env: c.Env}
	if c.root == nil {
		return objects
	}
	for _, configured := range c.root.src {
		if len(input) == 0 || configured.input == input[0] {
			objects.src = append(objects.src, configured)
		}
	}
	return objects
}

// Files returns every dnsmasq fragment expected from the current configuration.
func (c *Config) Files() *CFile {
	objects := c.GetAll()
	if c.root != nil && len(c.root.block) > 0 {
		objects.src = append(objects.src, &source{
			Env:  c.Env,
			kind: blockKind,
			name: "explicit",
		})
	}
	return objects.Files()
}

// InSession reports whether the command is running in an EdgeOS configure session.
func (c *Config) InSession() bool {
	return os.ExpandEnv("$_OFR_CONFIGURE") == "ok"
}

// load reads configuration through the EdgeOS cli-shell-api.
func (c *Config) load(action string) ([]byte, error) {
	cmd := exec.Command(c.Bash) // #nosec G204 -- the executable is package configuration, not user input.
	request := fmt.Sprintf("%v %v %v", c.API, apiCMD(action, c.InSession()), c.mode())
	c.Debug(fmt.Sprintf("Loading configuration with %s", request))
	cmd.Stdin = strings.NewReader(request)
	return cmd.Output()
}

// Nodes returns the configured top-level nodes. The schema has one.
func (c *Config) Nodes() []string {
	if c.root == nil {
		return nil
	}
	return []string{rootNode}
}

// ProcessContent processes one or more content groups.
func (c *Config) ProcessContent(contents ...Contenter) error {
	if len(contents) == 0 {
		return errors.New("empty Contenter interface{} passed to ProcessContent()")
	}

	var wg sync.WaitGroup
	errCh := make(chan error)
	for _, configured := range contents {
		for _, item := range configured.GetList().src {
			wg.Add(1)
			go func(item *source) {
				defer wg.Done()
				if item.err != nil {
					errCh <- item.err
					return
				}
				c.Lock()
				if _, exists := c.stat[item.kind.String()]; !exists {
					c.stat[item.kind.String()] = &stats{}
				}
				c.Unlock()
				if err := item.process().writeFile(); err != nil {
					errCh <- err
				}
			}(item)
		}
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var messages []string
	for err := range errCh {
		messages = append(messages, err.Error())
	}
	if len(messages) > 0 {
		sort.Strings(messages)
		return errors.New(strings.Join(messages, "\n"))
	}
	return nil
}

// Adblock parses the service dns forwarding adblock configuration tree.
func (c *Config) Adblock(loader ConfLoader) error {
	scanner := bufio.NewScanner(loader.read())
	var current *source
	seenSources := make(map[string]struct{})
	inRoot := false

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		text := string(line)
		c.Debug(text)

		if text == rootNode+" {" {
			c.root = newSource()
			c.root.Env = c.Env
			c.root.name = rootNode
			inRoot = true
			continue
		}
		if !inRoot {
			continue
		}

		if strings.HasPrefix(text, sourceNode+" ") && strings.HasSuffix(text, " {") {
			if current != nil {
				return errors.New("nested adblock sources are invalid")
			}
			name := strings.TrimSuffix(strings.TrimPrefix(text, sourceNode+" "), " {")
			name = strings.Trim(name, "\"'")
			if !sourceNamePattern.MatchString(name) {
				return fmt.Errorf("invalid adblock source name %q", name)
			}
			if _, exists := seenSources[name]; exists {
				return fmt.Errorf("duplicate adblock source %q", name)
			}
			seenSources[name] = struct{}{}
			current = newSource()
			current.Env = c.Env
			current.kind = sourceKind
			current.name = name
			continue
		}
		if strings.HasSuffix(text, " {") {
			return fmt.Errorf("unsupported adblock configuration node %q", strings.TrimSuffix(text, " {"))
		}

		if text == "}" {
			if current != nil {
				if err := c.finishSource(current); err != nil {
					return err
				}
				current = nil
				continue
			}
			inRoot = false
			continue
		}

		key, value := splitConfigLine(text)
		if current != nil {
			switch key {
			case "description":
				current.desc = value
			case redirectNode:
				current.ip = value
			case fileNode:
				current.file = value
			case "prefix":
				current.prefix = value
			case urlNode:
				current.url = value
			default:
				return fmt.Errorf("unsupported setting %q in adblock source %q", key, current.name)
			}
			continue
		}

		switch key {
		case allowNode:
			c.root.allow = append(c.root.allow, strings.ToLower(value))
		case blockNode:
			c.root.block = append(c.root.block, strings.ToLower(value))
		case disabledNode:
			parsed, err := strToBool(value)
			if err != nil {
				return fmt.Errorf("invalid disabled value: %w", err)
			}
			c.Disabled = parsed
		case redirectNode:
			c.root.ip = value
		default:
			return fmt.Errorf("unsupported adblock setting %q", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	if current != nil {
		return fmt.Errorf("adblock source %q is not closed", current.name)
	}
	if c.root == nil {
		return fmt.Errorf("%w has been detected", ErrNoAdblockCfg)
	}
	return nil
}

func splitConfigLine(line string) (string, string) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return line, ""
	}
	return parts[0], strings.Trim(strings.TrimSpace(parts[1]), "\"'")
}

func (c *Config) finishSource(configured *source) error {
	switch {
	case configured.url != "" && configured.file != "":
		return fmt.Errorf("adblock source %q must set exactly one of url or file", configured.name)
	case configured.url != "":
		parsed, err := url.ParseRequestURI(configured.url)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("adblock source %q has an invalid HTTP(S) URL", configured.name)
		}
		configured.input = urlNode
	case configured.file != "":
		configured.input = fileNode
	default:
		return fmt.Errorf("adblock source %q must set exactly one of url or file", configured.name)
	}
	if configured.ip == "" {
		configured.ip = c.redirectIP()
	}
	c.root.src = append(c.root.src, configured)
	return nil
}

func (c *Config) redirectIP() string {
	if c.root != nil && c.root.ip != "" {
		return c.root.ip
	}
	return "0.0.0.0"
}

func (c *Config) mode() string {
	if c.InSession() {
		return "--show-working-only"
	}
	return ""
}

// ReloadDNS reloads the dnsmasq configuration.
func (c *Config) ReloadDNS() ([]byte, error) {
	bash := c.Bash
	service := c.DNSsvc
	c = nil                   // release memory before restarting dnsmasq on memory-constrained routers
	cmd := exec.Command(bash) // #nosec G204 -- the executable is package configuration, not user input.
	cmd.Stdin = strings.NewReader(service)
	return cmd.CombinedOutput()
}
