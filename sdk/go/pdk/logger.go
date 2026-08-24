package pdk

import (
	"encoding/json"
	"os"
	"time"
)

// Logger escribe una línea JSON por evento a stdout — internal/plugins ya
// redirige stdout/stderr del proceso a BaseDir/logs/<name>.log, así que
// cualquier formato serviría para no perder el mensaje, pero JSON permite
// que una futura vista de logs en el dashboard los parsee en vez de tener
// que adivinar el formato de cada plugin.
type Logger struct {
	plugin string
}

// NewLogger crea un logger identificado con name — normalmente pdk.Name().
func NewLogger(name string) *Logger {
	return &Logger{plugin: name}
}

func (l *Logger) write(level, msg string, fields map[string]any) {
	entry := map[string]any{
		"ts":     time.Now().UTC().Format(time.RFC3339),
		"level":  level,
		"plugin": l.plugin,
		"msg":    msg,
	}
	for k, v := range fields {
		entry[k] = v
	}
	_ = json.NewEncoder(os.Stdout).Encode(entry)
}

func (l *Logger) Info(msg string, fields map[string]any)  { l.write("info", msg, fields) }
func (l *Logger) Warn(msg string, fields map[string]any)  { l.write("warn", msg, fields) }
func (l *Logger) Error(msg string, fields map[string]any) { l.write("error", msg, fields) }
