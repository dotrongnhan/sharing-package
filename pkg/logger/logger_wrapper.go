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

func NewLoggerWithInput(ctx context.Context, input interface{}) *log.Helper {
	traceID, _ := ctx.Value(TraceKey).(string)
	logger := NewJSONLoggerWithInput(traceID, input)
	return log.NewHelper(logger)
}

func NewJSONLoggerWithInput(traceID string, input interface{}) *JSONLogger {
	return &JSONLogger{
		Logger:  log.NewStdLogger(os.Stdout),
		TraceID: traceID,
		Input:   input,
	}
}

func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3) // Adjust stack depth as needed
	if !ok {
		return "unknown"
	}

	// Khởi tạo projectRoot (chỉ chạy 1 lần)
	// Lần đầu tiên chạy, nó sẽ dùng 'file' (từ main.go) để tìm root
	initRootOnce.Do(func() {
		findProjectRoot(file) // Truyền file của hàm gọi vào
	})

	// Tính toán đường dẫn tương đối
	// Nếu đã tìm được projectRoot, hãy tạo đường dẫn tương đối
	if projectRoot != "" {
		if relPath, err := filepath.Rel(projectRoot, file); err == nil {
			return fmt.Sprintf("%s:%d", relPath, line)
		}
	}

	// Fallback: Nếu không tìm được root, trả về đường dẫn đầy đủ
	return fmt.Sprintf("%s:%d", file, line)
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

	if l.Input != nil {
		entry.Input = l.Input
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

// findProjectRoot sẽ chạy 1 LẦN DUY NHẤT
// sử dụng file của hàm gọi đầu tiên để tìm gốc
func findProjectRoot(callerFile string) {
	// Bắt đầu từ thư mục chứa file gọi log, đi ngược lên trên
	// để tìm go.mod
	dir := filepath.Dir(callerFile)
	for {
		// Nếu tìm thấy go.mod, chúng ta xem đây là thư mục gốc
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			projectRoot = dir // Lưu lại đường dẫn tuyệt đối của gốc
			return
		}

		// Đi lên thư mục cha
		parent := filepath.Dir(dir)

		// Nếu đã lên đến thư mục gốc (root /) mà không thấy
		if parent == dir {
			return
		}
		dir = parent
	}
}
