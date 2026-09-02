package main

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	e "github.com/jonmeacham/edgeos-adblock/internal/edgeos"

	"flag"
)

// opts holds CLI flags. Certain flags are hidden from default help (visible=false),
// matching legacy mflag behavior.
type opts struct {
	*flag.FlagSet
	visible map[string]bool

	Dbug    *bool
	DNSdir  *string
	File    *string
	Help    *bool
	Verb    *bool
	Version *bool
}

func (o *opts) regBool(name string, value bool, usage string, visible bool) *bool {
	p := new(bool)
	o.BoolVar(p, name, value, usage)
	o.visible[name] = visible
	return p
}

func (o *opts) regString(name string, value string, usage string, visible bool) *string {
	p := new(string)
	o.StringVar(p, name, value, usage)
	o.visible[name] = visible
	return p
}

// cleanArgs removes flags when code is being tested.
func cleanArgs(args []string) (r []string) {
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "-test"), strings.HasPrefix(a, "-convey"):
			continue
		default:
			r = append(r, a)
		}
	}
	return r
}

// getCFG returns a e.ConfLoader.
func (o *opts) getCFG(c *e.Config) e.ConfLoader {
	if *o.File == "" {
		return &e.CFGcli{Config: c}
	}

	rd, err := e.GetFile(*o.File)
	if err != nil {
		logFatalf("cannot open configuration file %s!", *o.File)
	}
	f, err := io.ReadAll(rd)
	if err != nil {
		logFatalf("cannot read configuration file %s!", *o.File)
	}
	return &e.CFGstatic{Config: c, Cfg: string(f)}
}

// isFlagZeroValue mirrors flag.PrintDefaults logic (stdlib keeps this unexported).
func isFlagZeroValue(fl *flag.Flag, value string) bool {
	typ := reflect.TypeOf(fl.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Pointer {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	if value == z.Interface().(flag.Value).String() {
		return true
	}
	switch value {
	case "false", "", "0":
		return true
	default:
		return false
	}
}

func printFlagUsage(w io.Writer, f *flag.Flag) {
	var b strings.Builder
	fmt.Fprintf(&b, "  -%s", f.Name)

	name, usage := flag.UnquoteUsage(f)
	if len(name) > 0 {
		b.WriteString(" ")
		b.WriteString(name)
	}
	if len(b.String()) <= 4 {
		b.WriteString("\t")
	} else {
		b.WriteString("\n    \t")
	}
	b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))

	if !isFlagZeroValue(f, f.DefValue) {
		t := reflect.TypeOf(f.Value).String()
		if strings.Contains(t, "string") && strings.Contains(t, "flag.") {
			fmt.Fprintf(&b, " (default %q)", f.DefValue)
		} else {
			fmt.Fprintf(&b, " (default %v)", f.DefValue)
		}
	}
	fmt.Fprint(w, b.String(), "\n")
}

// getOpts returns command line flags and values or displays help.
func getOpts() *opts {
	fs := flag.NewFlagSet(prog, flag.ExitOnError)
	o := &opts{
		FlagSet: fs,
		visible: make(map[string]bool),
	}

	o.DNSdir = o.regString("dir", "/etc/dnsmasq.d", "Override dnsmasq directory", true)
	o.Dbug = o.regBool("debug", false, "Enable Debug mode", false)
	o.File = o.regString("f", "", "`<file>` # Load a config.boot file", true)
	o.Help = o.regBool("h", false, "Display help", true)
	o.Verb = o.regBool("v", false, "Verbose display", true)
	o.Version = o.regBool("version", false, "Show version", true)

	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", prog)
		fs.VisitAll(func(f *flag.Flag) {
			if !o.visible[f.Name] {
				return
			}
			printFlagUsage(fs.Output(), f)
		})
	}

	return o
}

func (o *opts) initEdgeOS() *e.Config {
	dnsmasq := "/bin/systemctl restart dnsmasq"
	if _, err := os.Stat("/bin/systemctl"); os.IsNotExist(err) {
		dnsmasq = "/etc/init.d/dnsmasq restart"
	}
	return e.NewConfig(
		e.API("/bin/cli-shell-api"),
		e.Bash("/bin/bash"),
		e.Disabled(false),
		e.Dbug(*o.Dbug),
		e.Dir(*o.DNSdir),
		e.DNSsvc(dnsmasq),
		e.Ext("edgeos-adblock.conf"),
		e.InCLI("inSession"),
		e.Method("GET"),
		e.Prefix("address="),
		e.SetLogger(log),
		e.Timeout(30*time.Second),
	)
}

// setArgs retrieves arguments entered on the command line.
func (o *opts) setArgs() {
	if o.Parse(cleanArgs(os.Args[1:])) != nil {
		exitCmd(0)
	}

	if *o.Dbug {
		screenLog("")
		e.Dbug(*o.Dbug)
		if logLogger != nil {
			logLogger.setDebug(true)
		}
	}

	if *o.Help {
		o.Usage()
		exitCmd(0)
	}

	if *o.Verb {
		screenLog("")
	}

	if *o.Version {
		fmt.Printf(
			" Build Information:\n"+
				"   Version:\t\t\t%s\n"+
				"   Date:\t\t\t%s\n"+
				"   CPU:\t\t\t\t%v\n"+
				"   OS:\t\t\t\t%v\n"+
				"   Git hash:\t\t\t%v\n\n"+
				" This software comes with ABSOLUTELY NO WARRANTY.\n"+
				" %s is free software, and you are\n"+
				" welcome to redistribute it under the terms of\n"+
				" the MIT License.\n",
			version,
			build,
			architecture,
			hostOS,
			githash,
			prog,
		)
		exitCmd(0)
	}
}
