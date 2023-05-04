package logger

import (
	"bytes"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	Eof = "EOFEOF"
)

var (
	Logrus *Logger
	once   sync.Once
)

func init() {
	once.Do(func() {
		Logrus = New()
	})
}

type Logger struct {
	*logrus.Logger
	root   string // 主目录
	file   string
	isInit bool
}

func New() *Logger {
	return &Logger{Logger: logrus.New(), root: "", file: "log"}
}

func (this *Logger) IsInit(isInit bool) *Logger {
	this.isInit = isInit
	return this
}

func (this *Logger) Init() {
	if this.isInit {
		return
	}
	// 日志文件
	path := fmt.Sprintf("%sruntime%slogs", this.root, string(os.PathSeparator))

	// 目录不存在，并创建
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if os.Mkdir(path, os.ModePerm) != nil {
		}
	}

	fileName := path + fmt.Sprintf("%s%s", string(os.PathSeparator), this.file)

	// 设置日志级别
	Logrus.SetLevel(logrus.DebugLevel)

	// 设置时间格式
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true // INFO[2020-09-28 13:20:14]
	Logrus.SetFormatter(customFormatter)

	// 设置 rotatelogs
	logWriter, _ := rotatelogs.New(
		// 分割后的文件名称
		fileName+".%Y%m%d.log",
		// 生成软链，指向最新日志文件
		rotatelogs.WithLinkName(fileName),
		// 设置最大保存时间(7天)
		rotatelogs.WithMaxAge(7*24*time.Hour),
		// 设置日志切割时间间隔(1天)
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logWriter,
		logrus.FatalLevel: logWriter,
		logrus.DebugLevel: logWriter,
		logrus.WarnLevel:  logWriter,
		logrus.ErrorLevel: logWriter,
		logrus.PanicLevel: logWriter,
	}

	lfHook := lfshook.NewHook(writeMap, &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	// Hook
	Logrus.AddHook(lfHook)
	this.isInit = true
}

func (this *Logger) Root(path string) *Logger {
	this.root = path
	return this
}

func (this *Logger) File(file string) *Logger {
	this.file = file
	return this
}

// 记录
func LoggerStack(err interface{}, goroutineID uint64) {
	strChan := make(chan string, 1000)
	PrintStackTrace(err, goroutineID, strChan)
	for {
		select {
		case v := <-strChan:
			if Eof == v {
				break
			}
			Logrus.Info(v)
		}
	}
}

// 打印堆栈信息
func PrintStackTrace(err interface{}, goroutineID uint64, strChan chan string) string {
	buf := new(bytes.Buffer)
	s := fmt.Sprintf("=== DEBUG STACK Bigin goroutine id %d ===", goroutineID)
	strChan <- s
	_, _ = fmt.Fprintf(buf, s+"\n")
	s = fmt.Sprintf("%v", err)
	strChan <- s
	_, _ = fmt.Fprintf(buf, s+"\n")
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		s = fmt.Sprintf("%s:%d (0x%x)", file, line, pc)
		strChan <- s
		_, _ = fmt.Fprintf(buf, s+"\n")
	}
	s = fmt.Sprintf("=== DEBUG STACK End goroutine id %d ===", goroutineID)
	_, _ = fmt.Fprintf(buf, s+"\n")
	strChan <- s
	strChan <- Eof
	return buf.String()
}
