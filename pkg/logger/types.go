package logger

import "github.com/go-kratos/kratos/v2/log"

const TraceKey = "trace_id"

type JSONLogger struct {
	Logger  log.Logger
	TraceID string
}

type logEntry struct {
	Time    string `json:"time"`
	Caller  string `json:"caller"`
	TraceID string `json:"trace_id,omitempty"`
	Msg     string `json:"msg"`
	Level   string `json:"level"`
}
