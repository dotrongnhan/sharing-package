package logger

import (
	"github.com/go-kratos/kratos/v2/log"
	"sync"
)

const (
	TraceKey         = "trace_id"
	TraceIDHeaderKey = "X-Trace-ID"
	LoggerWithInput  = "input"
)

type JSONLogger struct {
	Logger  log.Logger
	TraceID string
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
	loggerPackageDir string
	projectRoot      string
	initRootOnce     sync.Once // <-- Dùng lại sync.Once
)
