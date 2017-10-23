package log

// I'm trying out Uber's zap just for fun.
import "go.uber.org/zap"

// Logger is shared by...
var Logger *zap.SugaredLogger

// SetLogger will set the logger instance used by the rest of the Redis package.
// This should be set prior to the server startup.
func SetLogger(l *zap.SugaredLogger) {
	Logger = l
}

// NewLogger creates a new SugaredLogger. #Sync() should be deferred by the callee.
func NewLogger() *zap.SugaredLogger {
	l, err := zap.NewProduction()
	if err != nil {
		panic("Failed creating a logger, something is very wrong!")
	}
	return l.Sugar()
}
