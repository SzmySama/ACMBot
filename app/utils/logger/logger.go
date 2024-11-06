package logger

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	Reset     = "\033[0m"
	Red       = "\033[31m"
	Yellow    = "\033[33m"
	Green     = "\033[32m"
	Cyan      = "\033[36m"
	Magenta   = "\033[35m"
	TimeColor = "\033[34m" // Blue
	FuncColor = "\033[35m" // Magenta
)

func Init() {

}

func init() {
	logrus.SetFormatter(&LogFormatter{})
}

type LogFormatter struct{}

func (f *LogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var color string
	switch entry.Level {
	case logrus.TraceLevel:
		color = Magenta
	case logrus.DebugLevel:
		color = Green
	case logrus.InfoLevel:
		color = Cyan
	case logrus.WarnLevel:
		color = Yellow
	case logrus.ErrorLevel:
		color = Red
	case logrus.FatalLevel:
		color = Red
	case logrus.PanicLevel:
		color = Red
	}

	skip := 9

	funcName := getCallerInfo(skip)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	logMsg := fmt.Sprintf("%s[%s]%s %s%s()%s: %s%s%s", TimeColor, timestamp, Reset, FuncColor, funcName, Reset, color, entry.Message, Reset)
	return []byte(logMsg + "\n"), nil
}

func getCallerInfo(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return runtime.FuncForPC(pc).Name()
}
