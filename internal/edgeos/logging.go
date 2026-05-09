package edgeos

// Logger is the logging surface used by this package. It is implemented in
// the main module with log/slog (see sloglog.go).
type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Warning(args ...any)
	Warningf(format string, args ...any)
	Error(args ...any)
	Errorf(format string, args ...any)
	Noticef(format string, args ...any)
	Criticalf(format string, args ...any)
}
