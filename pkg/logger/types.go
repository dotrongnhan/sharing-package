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
	UserIDKey          = "user_id"
	defaultCallerDepth = 3
)

type JSONLogger struct {
	Logger log.Logger
	// ContextFields là các field rút từ context theo key đã đăng ký (xem RegisterContextLogKeys).
	ContextFields map[string]string
	TraceID       string
	Depth         int // Thêm dòng này
}
