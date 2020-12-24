package logger

import (
	"fmt"
	"github.com/dengpju/higo-utils/utils"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

var Logrus *Logger
var once sync.Once

func init() {
	once.Do(func() {
		Logrus = New()
	})
}

type Logger struct {
	*logrus.Logger
	root string // 主目录
	file string
}

func New() *Logger {
	return &Logger{Logger: logrus.New(), root: "", file: "log"}
}

func (this Logger) Init() {
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
}

func (this Logger) Root(path string) Logger {
	this.root = path
	return this
}

func (this Logger) File(file string) Logger {
	this.file = file
	return this
}

// 输出换行debug调用栈
func PrintlnStack() {
	ds := fmt.Sprintf("%s", debug.Stack())
	dss := strings.Split(ds, "\n")
	Logrus.Info(fmt.Sprintf("=== DEBUG STACK Bigin goroutine %d ===", utils.GoroutineID()))
	for _, b := range dss {
		Logrus.Info(strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%s", b), "\t"), "\n"))
	}
	Logrus.Info(fmt.Sprintf("=== DEBUG STACK End goroutine %d ===", utils.GoroutineID()))
}
