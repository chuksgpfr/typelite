package logger

// Logger is optional. If nil, typelite will not log.
type Logger interface {
	Printf(format string, args ...any)
}
