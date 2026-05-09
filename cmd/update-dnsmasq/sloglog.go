package main

import (
	"fmt"
	"io"
	"log/slog"
	"log/syslog"
	"os"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/term"

	e "github.com/jonmeacham/edgeos-adblock/internal/edgeos"
)

// inTerminalHook is set by tests to simulate a TTY (nil uses stdin probe).
var inTerminalHook func() bool

func inTerminal() bool {
	if inTerminalHook != nil {
		return inTerminalHook()
	}
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func setLogFile(goos string) string {
	if goos == "darwin" {
		return fmt.Sprintf("/tmp/%s.log", prog)
	}
	return fmt.Sprintf("/var/log/%s.log", prog)
}

var logFile = setLogFile(runtime.GOOS)

// slogLogger implements edgeos.Logger using log/slog (stdlib).
type slogLogger struct {
	mu          sync.Mutex
	level       slog.LevelVar
	fileSyslog  io.Writer
	stderrOn    bool
	log         *slog.Logger
	handlerOpts *slog.HandlerOptions
}

func (l *slogLogger) Debug(args ...any) {
	l.log.Debug(strings.TrimSpace(fmt.Sprint(args...)))
}

func (l *slogLogger) Info(args ...any) {
	l.log.Info(strings.TrimSpace(fmt.Sprint(args...)))
}

func (l *slogLogger) Infof(format string, args ...any) {
	l.log.Info(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Warning(args ...any) {
	l.log.Warn(strings.TrimSpace(fmt.Sprint(args...)))
}

func (l *slogLogger) Warningf(format string, args ...any) {
	l.log.Warn(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Error(args ...any) {
	l.log.Error(strings.TrimSpace(fmt.Sprint(args...)))
}

func (l *slogLogger) Errorf(format string, args ...any) {
	l.log.Error(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Noticef(format string, args ...any) {
	l.log.Info(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Criticalf(format string, args ...any) {
	l.log.Error(fmt.Sprintf(format, args...))
}

func newSlogLogger() *slogLogger {
	l := &slogLogger{}
	l.level.Set(slog.LevelInfo)
	l.handlerOpts = &slog.HandlerOptions{Level: &l.level}

	fd, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		fd = os.Stderr
	}

	writers := []io.Writer{fd}
	if sysWr, err := syslog.New(syslog.LOG_DAEMON|syslog.LOG_INFO, prog+": "); err == nil {
		writers = append(writers, sysWr)
	} else {
		fmt.Fprint(os.Stderr, err.Error())
	}

	l.fileSyslog = io.MultiWriter(writers...)
	l.rebuildUnlocked()
	return l
}

// rebuildUnlocked composes file/syslog and optional stderr via slog.NewMultiHandler (Go 1.26+).
func (l *slogLogger) rebuildUnlocked() {
	h := []slog.Handler{slog.NewTextHandler(l.fileSyslog, l.handlerOpts)}
	if l.stderrOn {
		h = append(h, slog.NewTextHandler(os.Stderr, l.handlerOpts))
	}
	l.log = slog.New(slog.NewMultiHandler(h...))
}

// attachStderr mirrors legacy "screen" logging when stderr is a TTY (-v / -debug).
func (l *slogLogger) attachStderr() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.stderrOn || !inTerminal() {
		return
	}
	l.stderrOn = true
	l.rebuildUnlocked()
}

func (l *slogLogger) setDebug(on bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if on {
		l.level.Set(slog.LevelDebug)
	} else {
		l.level.Set(slog.LevelInfo)
	}
}

// screenLog enables stderr logging on interactive terminals (legacy behavior).
func screenLog(_ string) {
	if logLogger != nil {
		logLogger.attachStderr()
	}
}

var (
	logLogger *slogLogger
	log       e.Logger

	logErrorf  func(f string, args ...any)
	logFatalf  func(f string, args ...any)
	logInfo    func(args ...any)
	logNoticef func(f string, args ...any)
	logPrintf  func(f string, args ...any)
)

func initLoggingVars() {
	logErrorf = func(f string, args ...any) { log.Errorf(f, args...) }
	logFatalf = func(f string, args ...any) { log.Criticalf(f, args...); exitCmd(1) }
	logInfo = func(args ...any) { log.Info(args...) }
	logNoticef = func(f string, args ...any) { log.Noticef(f, args...) }
	logPrintf = func(f string, args ...any) { log.Infof(f, args...) }
}

func initLogging() {
	logLogger = newSlogLogger()
	log = logLogger
	initLoggingVars()
}

func init() {
	initLogging()
}
