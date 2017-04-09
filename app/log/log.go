package log

//go:generate go run $GOPATH/src/v2ray.com/core/tools/generrorgen/main.go -pkg log -path App,Log

import (
	"context"

	"v2ray.com/core/app"
	"v2ray.com/core/app/log/internal"
	"v2ray.com/core/common"
	"v2ray.com/core/common/errors"
)

var (
	streamLoggerInstance internal.LogWriter = internal.NewStdOutLogWriter()

	debugLogger   internal.LogWriter = streamLoggerInstance
	infoLogger    internal.LogWriter = streamLoggerInstance
	warningLogger internal.LogWriter = streamLoggerInstance
	errorLogger   internal.LogWriter = streamLoggerInstance
)

func SetLogLevel(level LogLevel) {
	debugLogger = new(internal.NoOpLogWriter)
	if level >= LogLevel_Debug {
		debugLogger = streamLoggerInstance
	}

	infoLogger = new(internal.NoOpLogWriter)
	if level >= LogLevel_Info {
		infoLogger = streamLoggerInstance
	}

	warningLogger = new(internal.NoOpLogWriter)
	if level >= LogLevel_Warning {
		warningLogger = streamLoggerInstance
	}

	errorLogger = new(internal.NoOpLogWriter)
	if level >= LogLevel_Error {
		errorLogger = streamLoggerInstance
	}
}

func InitErrorLogger(file string) error {
	logger, err := internal.NewFileLogWriter(file)
	if err != nil {
		return newError("failed to create error logger on file (", file, ")").Base(err)
	}
	streamLoggerInstance = logger
	return nil
}

// writeDebug outputs a debug log with given format and optional arguments.
func writeDebug(val ...interface{}) {
	debugLogger.Log(&internal.ErrorLog{
		Prefix: "[Debug]",
		Values: val,
	})
}

// writeInfo outputs an info log with given format and optional arguments.
func writeInfo(val ...interface{}) {
	infoLogger.Log(&internal.ErrorLog{
		Prefix: "[Info]",
		Values: val,
	})
}

// writeWarning outputs a warning log with given format and optional arguments.
func writeWarning(val ...interface{}) {
	warningLogger.Log(&internal.ErrorLog{
		Prefix: "[Warning]",
		Values: val,
	})
}

// writeError outputs an error log with given format and optional arguments.
func writeError(val ...interface{}) {
	errorLogger.Log(&internal.ErrorLog{
		Prefix: "[Error]",
		Values: val,
	})
}

func Trace(err error) {
	s := errors.GetSeverity(err)
	switch s {
	case errors.SeverityDebug:
		writeDebug(err)
	case errors.SeverityInfo:
		writeInfo(err)
	case errors.SeverityWarning:
		writeWarning(err)
	case errors.SeverityError:
		writeError(err)
	default:
		writeInfo(err)
	}
}

type Instance struct {
	config *Config
}

func New(ctx context.Context, config *Config) (*Instance, error) {
	return &Instance{config: config}, nil
}

func (*Instance) Interface() interface{} {
	return (*Instance)(nil)
}

func (g *Instance) Start() error {
	config := g.config
	if config.AccessLogType == LogType_File {
		if err := InitAccessLogger(config.AccessLogPath); err != nil {
			return err
		}
	}

	if config.ErrorLogType == LogType_None {
		SetLogLevel(LogLevel_Disabled)
	} else {
		if config.ErrorLogType == LogType_File {
			if err := InitErrorLogger(config.ErrorLogPath); err != nil {
				return err
			}
		}
		SetLogLevel(config.ErrorLogLevel)
	}

	return nil
}

func (*Instance) Close() {
	streamLoggerInstance.Close()
	accessLoggerInstance.Close()
}

func FromSpace(space app.Space) *Instance {
	v := space.GetApplication((*Instance)(nil))
	if logger, ok := v.(*Instance); ok && logger != nil {
		return logger
	}
	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
