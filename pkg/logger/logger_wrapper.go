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
	"slices"
	"strings"
	"sync"
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

// contextLogKeys là các key được phép rút từ context vào log. Chỉ định danh sách tường minh
// vì Go không cho duyệt context, và để giá trị nhạy cảm/nặng (scope, transaction) không lọt log.
var contextLogKeys = []string{UserIDKey}

var contextLogKeysMu sync.RWMutex

// RegisterContextLogKeys khai báo thêm key cần log. Gọi lúc khởi động app, trước khi phục vụ
// request — danh sách là global nên đăng ký giữa chừng sẽ làm log không đồng nhất.
func RegisterContextLogKeys(keys ...string) {
	contextLogKeysMu.Lock()
	defer contextLogKeysMu.Unlock()

	for _, key := range keys {
		if key == "" || slices.Contains(contextLogKeys, key) {
			continue
		}
		contextLogKeys = append(contextLogKeys, key)
	}
}

// extractContextFields rút các key đã đăng ký. Chỉ nhận string: value kiểu khác (slice, struct,
// *gorm.DB…) thường là dữ liệu nội bộ, ghi ra log vừa nặng vừa dễ lộ.
func extractContextFields(ctx context.Context) map[string]string {
	contextLogKeysMu.RLock()
	defer contextLogKeysMu.RUnlock()

	fields := make(map[string]string, len(contextLogKeys))
	for _, key := range contextLogKeys {
		value, ok := ctx.Value(key).(string)
		if !ok || value == "" {
			continue
		}
		fields[key] = value
	}
	return fields
}

func NewLogger(ctx context.Context) *log.Helper {
	traceID, _ := ctx.Value(TraceKey).(string)
	logger := NewJSONLogger(traceID, defaultCallerDepth)
	logger.ContextFields = extractContextFields(ctx)
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
	rawLogger.ContextFields = extractContextFields(ctx)

	loggerWithFields := log.With(rawLogger, keyvals...)
	return log.NewHelper(loggerWithFields)
}

func getCallerInfo(depth int) string {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		return "unknown"
	}

	// --- Phương pháp 1: Ưu tiên Production (Theo yêu cầu của bạn) ---
	// Tìm "go/src/backend/"
	const prodAnchor = "go/src/backend/"
	idx := strings.Index(file, prodAnchor)
	if idx != -1 {
		// Cắt chuỗi để lấy mọi thứ SAU "go/src/backend/"
		relativePath := file[idx+len(prodAnchor):]
		return fmt.Sprintf("%s:%d", relativePath, line)
	}

	// --- Phương pháp 2: Thử kiểu Local (Cách bình thường) ---
	// (Chỉ chạy nếu Phương pháp 1 thất bại)
	rootPath, err := filepath.Abs("")
	if err == nil {
		relativePath, err := filepath.Rel(rootPath, file)
		if err == nil && !strings.HasPrefix(relativePath, "..") {
			return fmt.Sprintf("%s:%d", relativePath, line)
		}
	}

	// --- Phương pháp 3: Fallback (Dự phòng) ---
	// Nếu cả 2 đều thất bại, trả về tên file
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

func (l *JSONLogger) Log(level log.Level, keyvals ...interface{}) error {
	entryMap := make(map[string]interface{})

	// Thêm các trường cố định
	entryMap[TimeKey] = time.Now().Format(time.RFC3339)
	entryMap[LevelKey] = level.String()
	entryMap[TraceKey] = l.TraceID
	entryMap[CallerKey] = getCallerInfo(l.Depth)

	// Đặt trước vòng keyvals để caller truyền tay cùng key vẫn ghi đè được.
	for key, value := range l.ContextFields {
		entryMap[key] = value
	}

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
