package logger

import (
	"github.com/go-kratos/kratos/v2/log"
	"sync"
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

type logEntry struct {
	Time    string      `json:"time"`
	Caller  string      `json:"caller"`
	TraceID string      `json:"trace_id,omitempty"`
	Msg     string      `json:"msg"`
	Level   string      `json:"level"`
	Input   interface{} `json:"input,omitempty"`
}

var (
	fileDirCache sync.Map
)
