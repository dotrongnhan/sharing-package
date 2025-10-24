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

func NewLoggerWith(ctx context.Context, keyvals ...interface{}) *log.Helper {
	// 1. Lấy traceID từ context
	traceID, _ := ctx.Value(TraceKey).(string)

	// 2. Tạo logger GỐC (*JSONLogger)
	rawLogger := NewJSONLogger(traceID)

	// 3. Dùng log.With toàn cục để thêm các keyvals (Cách này thêm 1 lớp stack)
	loggerWithFields := log.With(rawLogger, keyvals...)

	// 4. Bọc lại bằng Helper
	return log.NewHelper(loggerWithFields)
}

func getCallerInfo() string {
	pcs := make([]uintptr, 20)
	n := runtime.Callers(3, pcs) // Bắt đầu từ 3 (bỏ qua Callers, getCallerInfo, JSONLogger.Log)
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		if frame.File == "" {
			break
		}

		frameDir := filepath.Dir(frame.File)

		// Lọc bỏ file của logger và thư viện
		if frameDir == loggerPackageDir ||
			strings.Contains(frame.File, "go/pkg/mod") ||
			strings.Contains(frame.File, "/vendor/") {

			if !more {
				break
			}
			continue
		}

		// --- Đã tìm thấy frame của người dùng ---

		// Dùng sync.Once để tìm project root DỰA TRÊN file của frame
		// Đây là mấu chốt để tìm đúng go.mod
		initRootOnce.Do(func() {
			findProjectRoot(frame.File)
		})

		// Sử dụng projectRoot đã được tính toán
		if projectRoot != "" {
			if rel, err := filepath.Rel(projectRoot, frame.File); err == nil {
				return fmt.Sprintf("%s:%d", rel, frame.Line)
			}
		}

		// Fallback:
		return fmt.Sprintf("%s:%d", frame.File, frame.Line)
	}
	return "unknown"
}

func (l *JSONLogger) Log(level log.Level, keyvals ...interface{}) error {
	entry := logEntry{
		Time:    time.Now().Format(time.RFC3339),
		Level:   level.String(),
		TraceID: l.TraceID,
		Caller:  getCallerInfo(),
	}

	// Lặp qua TẤT CẢ keyvals
	if len(keyvals) > 1 {
		for i := 0; i < len(keyvals)-1; i += 2 {
			key, ok := keyvals[i].(string)
			if !ok {
				continue
			}
			val := keyvals[i+1]

			switch key {
			case "msg":
				entry.Msg = fmt.Sprintf("%v", val)
			case "input":
				entry.Input = val
				// Bạn có thể thêm các case khác nếu muốn (ví dụ: "error")
			}
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

func init() {
	// Chỉ cần lấy thư mục của package logger
	_, file, _, ok := runtime.Caller(0)
	if ok {
		loggerPackageDir = filepath.Dir(file)
	}
}
