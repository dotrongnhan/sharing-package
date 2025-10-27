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
	logger := NewJSONLogger(traceID, defaultCallerDepth)
	return log.NewHelper(logger)
}

func NewJSONLogger(traceID string, depth int) *JSONLogger {
	return &JSONLogger{
		Logger:  log.NewStdLogger(os.Stdout),
		TraceID: traceID,
		Depth:   depth,
	}
}

func NewLoggerWith(ctx context.Context, keyvals ...interface{}) *log.Helper {
	traceID, _ := ctx.Value(TraceKey).(string)

	// Sửa: Dùng defaultCallerDepth + 1 (vì có thêm 1 lớp log.With)
	rawLogger := NewJSONLogger(traceID, defaultCallerDepth+1)

	loggerWithFields := log.With(rawLogger, keyvals...)
	return log.NewHelper(loggerWithFields)
}

func getCallerInfo(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		return "unknown"
	}

	// Lấy thư mục của file
	dir := filepath.Dir(file)

	// Tìm thư mục gốc dự án cho thư mục này
	projectRoot := findProjectRootForDir(dir)

	if projectRoot != "" {
		// Nếu tìm thấy gốc, tính toán đường dẫn tương đối
		relativePath, err := filepath.Rel(projectRoot, file)
		if err == nil {
			return fmt.Sprintf("%s:%d", relativePath, line)
		}
	}

	// Fallback: Nếu không tìm thấy go.mod hoặc có lỗi,
	// chỉ trả về tên file (an toàn hơn đường dẫn tuyệt đối)
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

func (l *JSONLogger) Log(level log.Level, keyvals ...interface{}) error {
	entryMap := make(map[string]interface{})

	// Thêm các trường cố định
	entryMap[TimeKey] = time.Now().Format(time.RFC3339)
	entryMap[LevelKey] = level.String()
	entryMap[TraceKey] = l.TraceID
	entryMap[CallerKey] = getCallerInfo(l.Depth)

	// Lặp qua TẤT CẢ keyvals và thêm vào map
	if len(keyvals) > 1 {
		for i := 0; i < len(keyvals)-1; i += 2 {
			key, ok := keyvals[i].(string)
			if !ok {
				continue
			}
			val := keyvals[i+1]

			// Xử lý "msg" đặc biệt để đảm bảo nó là string
			if key == MsgKey {
				entryMap[MsgKey] = fmt.Sprintf("%v", val)
			} else {
				entryMap[key] = val
			}
		}
	}

	b, err := json.Marshal(entryMap)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, string(b))
	return err
}

func findProjectRootForDir(dir string) string {
	// 1. Kiểm tra cache trước
	if root, ok := fileDirCache.Load(dir); ok {
		return root.(string)
	}

	// 2. Không có trong cache, bắt đầu tìm kiếm
	currentDir := dir
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		// Kiểm tra xem file go.mod có tồn tại không
		if _, err := os.Stat(goModPath); err == nil {
			// Đã tìm thấy! Lưu vào cache và trả về.
			fileDirCache.Store(dir, currentDir)
			return currentDir
		}

		// Đi lên thư mục cha
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Đã đến thư mục gốc của hệ thống (ví dụ: "/")
			break
		}
		currentDir = parentDir
	}

	// 3. Không tìm thấy. Lưu chuỗi rỗng vào cache để không tìm lại.
	fileDirCache.Store(dir, "")
	return ""
}
