package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	e "github.com/jonmeacham/edgeos-adblock/internal/edgeos"
)

var (
	// updated by go build -ldflags
	architecture = "UNKNOWN"
	build        = "UNKNOWN"
	githash      = "UNKNOWN"
	hostOS       = "UNKNOWN"
	version      = "UNKNOWN"
	// ----------------------------

	exitCmd      = os.Exit
	initEnvirons = initEnv
	prog         = progName(os.Args[0])
	bkpCfgFile   = "/config/user-data/edgeos-adblock.failover.cfg"
	stdCfgFile   = "/config/config.boot"
)

// progName returns the executable basename without its extension (legacy behavior).
func progName(argv0 string) string {
	base := filepath.Base(argv0)
	if ext := filepath.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return base
}

func main() {
	// Memory profiling
	// defer profile.Start(profile.MemProfile).Stop()

	objex := []e.IFace{
		e.PreRObj,
		e.PreDObj,
		e.PreHObj,
		e.ExRtObj,
		e.ExDmObj,
		e.ExHtObj,
		e.FileObj,
		e.URLdObj,
		e.URLhObj,
	}

	if os.Geteuid() != 0 {
		fmt.Printf("%s must be run as sudo\n", prog)
		logErrorf("%s must be run as sudo", prog)
		exitCmd(0)
	}
	c, err := initEnvirons()
	if err != nil {
		logErrorf("Cannot continue due to error: %s", err.Error())
		exitCmd(0)
	}

	c.Debug(fmt.Sprintf("Dumping commandline args: %v", os.Args[1:]))
	c.Debug(fmt.Sprintf("Dumping env variables: %v", c))
	logNoticef("%v", "Starting edgeos-adblock update...")

	if !e.ChkWeb("www.google.com", 443) {
		logFatalf("%s", "No internet access, aborting edgeos-adblock update!")
	}

	logInfo("Checking for stale downloaded lists...")
	if err = removeStaleFiles(c); err != nil {
		logFatalf("%v", err.Error())
	}

	// _, _ = context.WithTimeout(context.Background(), c.Timeout)

	if !c.Disabled {
		if err := processObjects(c, objex); err != nil {
			logErrorf("%v", err.Error())
		}
	}

	dropped, extracted, kept := c.GetTotalStats()
	if kept+dropped != 0 {
		c.Log.Noticef("Total entries found: %d", extracted)
		c.Log.Noticef("Total entries extracted %d", kept)
		c.Log.Noticef("Total entries dropped %d", dropped)
	}

	reloadDNS(c)

	logNoticef("%v", "edgeos-adblock update completed......")
}

// files returns an empty *e.CFile string array
func files(c *e.Config) *e.CFile {
	return &e.CFile{Names: []string{}, Env: c.Env}
}

func initEnv() (c *e.Config, err error) {
	o := getOpts()
	o.setArgs()
	c = o.initEdgeOS()
	if *o.File == "" {
		if *o.Safe {
			*o.File = stdCfgFile
		}
		if c, err = loadConfig(c, o); err != nil {
			if _, err = os.Stat(bkpCfgFile); !os.IsNotExist(err) && *o.Safe {
				*o.File = bkpCfgFile
			}
		}
	}

	return loadConfig(c, o)
}

func loadConfig(c *e.Config, o *opts) (*e.Config, error) {
	var err error

	if err = c.Blocklist(o.getCFG(c)); err != nil {
		fmt.Fprintf(os.Stderr, "Removing stale dnsmasq edgeos-adblock files, because %v\n", err.Error())
		if err = files(c).Remove(); err != nil {
			fmt.Fprintf(os.Stderr, "%v", err.Error())
		}
		reloadDNS(c)
		exitCmd(0)
	}

	return c, err
}

// processObjects processes local sources, downloads Internet sources and creates
// dnsmasq configuration files
func processObjects(c *e.Config, objects []e.IFace) error {
	for _, o := range objects {
		ct, err := c.NewContent(o)
		if err != nil {
			return err
		}
		if err = c.ProcessContent(ct); err != nil {
			return err
		}
	}
	return nil
}

// reloadDNS reloads the latest processed dnsmasq configuration files
func reloadDNS(c *e.Config) {
	if b, err := c.ReloadDNS(); err != nil {
		logErrorf("ReloadDNS(): %v\n error: %v\n", string(b), err.Error())
		exitCmd(1)
	}
	logPrintf("%s", "Successfully restarted dnsmasq")
}

// removeStaleFiles deletes redundant files
func removeStaleFiles(c *e.Config) error {
	if err := c.GetAll().Files().Remove(); err != nil {
		return fmt.Errorf("problem removing stale files: %v", err.Error())
	}
	return nil
}
