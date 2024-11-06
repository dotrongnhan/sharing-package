package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func NewContextWithTraceID(ctx context.Context) context.Context {
	traceID := uuid.NewString()
	return context.WithValue(ctx, TraceKey, traceID)
}
func GenerateTraceID() string {
	uUid := uuid.NewString()
	traceID := fmt.Sprintf("%s", strings.ReplaceAll(uUid, "-", ""))
	return traceID
}
func NewBackgroundContextWithTraceID(serviceName string) context.Context {
	return NewContextWithTraceID(context.Background())
}

func NewLogger(ctx context.Context) *log.Helper {
	traceID, _ := ctx.Value(TraceKey).(string)
	logger := NewJSONLogger(traceID)
	return log.NewHelper(logger)
}
func NewJSONLogger(traceID string) *JSONLogger {
	return &JSONLogger{
		Logger:  log.NewStdLogger(os.Stdout),
		TraceID: traceID,
	}
}

func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3) // Adjust stack depth as needed
	if !ok {
		return "unknown"
	}

	// Lấy đường dẫn gốc từ thư mục chạy (root project)
	// filepath.Abs("") trả về đường dẫn tuyệt đối của thư mục hiện tại.
	rootPath, err := filepath.Abs("")
	if err != nil {
		return "unknown"
	}

	// Chuyển đường dẫn file thành đường dẫn tương đối so với rootPath
	relativePath, err := filepath.Rel(rootPath, file)
	if err != nil {
		return fmt.Sprintf("%s:%d", file, line) // Trả về đường dẫn gốc nếu không thể lấy tương đối
	}

	return fmt.Sprintf("%s:%d", relativePath, line)
}

func (l *JSONLogger) Log(level log.Level, keyvals ...interface{}) error {
	entry := logEntry{
		Time:    time.Now().Format(time.RFC3339),
		Level:   level.String(),
		TraceID: l.TraceID,
		Caller:  getCallerInfo(),
	}

	// Xử lý keyvals để thêm vào message nếu có các giá trị bổ sung
	if len(keyvals) > 1 {
		for i := 0; i < len(keyvals)-1; i += 2 {
			if keyvals[i] == "msg" {
				entry.Msg = fmt.Sprintf("%v", keyvals[i+1])
			}
			// Có thể xử lý thêm các keyvals khác nếu cần thiết
		}
	}

	// Chuyển entry sang JSON
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Ghi trực tiếp chuỗi JSON mà không thêm bất kỳ key-value nào khác
	_, err = fmt.Fprintln(os.Stdout, string(b))
	return err
}
