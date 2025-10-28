package logger

import (
	"github.com/go-kratos/kratos/v2/log"
)

const (
	TraceKey           = "trace_id"
	TraceIDHeaderKey   = "X-Trace-ID"
	InputKey           = "input"
	CallerKey          = "caller"
	TimeKey            = "time"
	MsgKey             = "msg"
	LevelKey           = "level"
	defaultCallerDepth = 3
)

type JSONLogger struct {
	Logger  log.Logger
	TraceID string
	Depth   int // Thêm dòng này
}
