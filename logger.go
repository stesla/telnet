package telnet

type LogLevel int

const (
	ERROR LogLevel = 0 + iota
	WARN
	INFO
	DEBUG
)

func (ll LogLevel) String() string {
	switch ll {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger interface {
	Logf(level LogLevel, fmt string, v ...any)
	SetLevel(level LogLevel)
}

type NullLogger struct{}

func (NullLogger) Log(LogLevel, ...any)          {}
func (NullLogger) Logf(LogLevel, string, ...any) {}
func (NullLogger) SetLevel(LogLevel)             {}

type Log interface {
	Printf(fmt string, v ...any)
}

type LogLogger struct {
	log   Log
	level LogLevel
}

func NewLogLogger(log Log) Logger {
	return &LogLogger{log: log, level: WARN}
}

func (l *LogLogger) Logf(level LogLevel, fmt string, v ...any) {
	if level <= l.level {
		fmt = "[%s] " + fmt
		args := []any{level}
		args = append(args, v...)
		l.log.Printf(fmt, args...)
	}
}

func (l *LogLogger) SetLevel(level LogLevel) { l.level = level }
