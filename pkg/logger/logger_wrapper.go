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
	"runtime/debug"
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

	anchor := getProjectAnchor()
	if anchor != "" {
		searchString := anchor + "/"
		idx := strings.Index(file, searchString)

		if idx != -1 {
			relativePath := file[idx+len(searchString):]
			return fmt.Sprintf("%s:%d", relativePath, line)
		}
	}

	dir := filepath.Dir(file)
	projectRoot := findProjectRootForDir(dir)
	if projectRoot != "" {
		relativePath, err := filepath.Rel(projectRoot, file)
		if err == nil {
			return fmt.Sprintf("%s:%d", relativePath, line)
		}
	}

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

// --- (Helper 1) Phương pháp Production (debug.ReadBuildInfo) ---
func getProjectAnchor() string {
	anchorOnce.Do(func() {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			projectAnchor = ""
			return
		}

		anchor := info.Main.Path

		if anchor == "command-line-arguments" || anchor == "" {
			projectAnchor = ""
			return
		}
		projectAnchor = anchor
	})
	return projectAnchor
}

// --- (Helper 2) Phương pháp Local 'go run' (Tìm go.mod) ---
func findProjectRootForDir(dir string) string {
	if root, ok := fileDirCache.Load(dir); ok {
		return root.(string)
	}

	currentDir := dir
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			fileDirCache.Store(dir, currentDir)
			return currentDir
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			break
		}
		currentDir = parentDir
	}

	fileDirCache.Store(dir, "")
	return ""
}
