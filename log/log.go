package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const (
	DebugLevel   = 0
	ReleaseLevel = 1
	ErrorLevel   = 2
	FatalLevel   = 3
)

const (
	PrintDebugLevel   = "[debug  ] "
	PrintReleaseLevel = "[release] "
	PrintErrorLevel   = "[error  ] "
	PrintFatalLevel   = "[fatal  ] "
)

type Logger struct {
	level      int
	baseLogger *log.Logger
	baseFile   *os.File //文件句柄
	todaydate  string
	msgQueue   chan string // 所有的日志先到这来
	closed     bool
}

// New 创建一个自己的日志对象。
// pathname:文件夹路径名称
// strLevel: 打印等级。DEBUG, INFO, ERROR
// flag定义日志的属性（时间、文件等等）
func New(strLevel string, pathname string, flag int) (*Logger, error) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = DebugLevel
	case "release":
		level = ReleaseLevel
	case "error":
		level = ErrorLevel
	case "fatal":
		level = FatalLevel
	default:
		return nil, errors.New("unknown level: " + strLevel)
	}

	// logger
	var baseLogger *log.Logger
	var baseFile *os.File

	if pathname != "" {
		now := time.Now()

		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second())

		file, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			return nil, err
		}

		baseLogger = log.New(file, "", flag)
		baseFile = file
	} else {
		baseLogger = log.New(os.Stdout, "", flag)
	}

	//文件夹路径名称为空，不创建日志文件，直接终端打印信息
	logger := new(Logger)
	logger.level = level
	logger.baseLogger = baseLogger
	logger.baseFile = baseFile
	logger.todaydate = time.Now().Format("2006-01-02")
	logger.msgQueue = make(chan string, 1000)
	logger.closed = false

	// 启动日志切换
	go logger.logworker(pathname)

	return logger, nil
}

func (logger *Logger) logworker(pathname string) {
	for !logger.closed {
		msg := <-logger.msgQueue
		logger.baseLogger.Output(3, msg)

		//跨日改时间，后台启动
		nowDate := time.Now().Format("2006-01-02")
		if nowDate != logger.todaydate {
			// logger.Debug("doRotate run %v %v", nowDate, logger.todaydate)
			logger.doRotate(pathname)
		}
	}
}

// 日志按天切换文件
func (logger *Logger) doRotate(pathname string) {
	// 首先关闭文件句柄，把当前日志改名为昨天，再创建新的文件句柄，将这个文件句柄赋值给log对象
	defer func() {
		rec := recover()
		if rec != nil {
			fmt.Printf("doRotate %v", rec)
		}
	}()

	if logger.baseFile == nil {
		return
	}

	prefile := logger.baseFile

	_, err := prefile.Stat()
	if err == nil {
		err := prefile.Close()
		if err != nil {
			fmt.Printf("doRotate rename err %v", err)
			return
		}
	}

	if pathname != "" {
		now := time.Now()
		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second())
		nextfile, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			return
		}
		logger.baseFile = nextfile

		fmt.Println("newLogger use MultiWriter")
		multi := io.MultiWriter(nextfile, os.Stdout)
		logger.baseLogger.SetOutput(multi)
	}

	// 更新标记，这个标记决定是否会启动文件切换
	nowDate := time.Now().Format("2006-01-02")
	logger.todaydate = nowDate
}

// It's dangerous to call the method on logging
func (logger *Logger) Close() {
	if logger.baseFile != nil {
		logger.baseFile.Close() //关闭文件
	}

	logger.baseLogger = nil
	logger.baseFile = nil
	logger.closed = true
}

func (logger *Logger) doPrintf(level int, printLevel string, format string, a ...interface{}) {
	if level < logger.level {
		return
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}

	format = printLevel + format
	logger.msgQueue <- fmt.Sprintf(format, a...)
	// logger.baseLogger.Output(3, fmt.Sprintf(format, a...))

	if level == FatalLevel {
		os.Exit(1)
	}
}

func (logger *Logger) Debug(format string, a ...interface{}) {
	logger.doPrintf(DebugLevel, PrintDebugLevel, format, a...)
}

func (logger *Logger) Release(format string, a ...interface{}) {
	logger.doPrintf(ReleaseLevel, PrintReleaseLevel, format, a...)
}

func (logger *Logger) Error(format string, a ...interface{}) {
	logger.doPrintf(ErrorLevel, PrintErrorLevel, format, a...)
}

func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(FatalLevel, PrintFatalLevel, format, a...)
}

var gLogger, _ = New("debug", "", log.LstdFlags)

// It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		gLogger = logger
	}
}

func Debug(format string, a ...interface{}) {
	gLogger.doPrintf(DebugLevel, PrintDebugLevel, format, a...)
}

func Release(format string, a ...interface{}) {
	gLogger.doPrintf(ReleaseLevel, PrintReleaseLevel, format, a...)
}

func Error(format string, a ...interface{}) {
	gLogger.doPrintf(ErrorLevel, PrintErrorLevel, format, a...)
}

func Fatal(format string, a ...interface{}) {
	gLogger.doPrintf(FatalLevel, PrintFatalLevel, format, a...)
}

func Close() {
	gLogger.Close()
}
