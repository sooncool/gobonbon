package log_test

import (
	"gobonbon/log"
	l "log"
	"testing"
	"time"
)

func TestExample(t *testing.T) {
	name := "BonBon"
	log.Debug("My name is %v", name)
	log.Release("My name is %v", name)
	log.Error("My name is %v", name)
	// log.Fatal("My name is %v", logger)

	//测试文件名空
	// logger, err := log.New("release", "", l.LstdFlags)
	// if err != nil {
	// 	return
	// }
	// defer logger.Close()

	// logger.Debug("will not print")
	// logger.Error("will print error")
	// logger.Release("My name is %v", name)

	// log.Export(logger)

	// log.Debug("will not print")
	// log.Release("My name is %v", name)

	//测试文件
	logger, err := log.New("release", "log", l.LstdFlags)
	if err != nil {
		return
	}
	// defer logger.Close()
	logger.Debug("will not print")
	logger.Release("My logger is %v", name)
	logger.Release("My logger is %v", "ten second"+name)
	logger.Release("My logger is %v", "ten second"+name)
	logger.Release("My logger is %v", "ten second"+name)
	logger.Release("My logger is %v", "ten second"+name)
	logger.Release("My logger is %v", "ten second"+name)
	time.Sleep(2 * time.Second)
	logger.Release("My logger is %v", "ten second"+name)
	time.Sleep(2 * time.Second)
	log.Export(logger)

	log.Debug("will not print")
	log.Release("My name is %v", name+"end")
	time.Sleep(2 * time.Second)
}
