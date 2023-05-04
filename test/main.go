package main

import (
	"fmt"
	"github.com/dunpju/higo-logger/logger"
	"github.com/dunpju/higo-utils/utils/runtimeutil"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover...:", r)
			id, _ := runtimeutil.GoroutineID()
			logger.LoggerStack(r, id)
		}
	}()
	logger.Logrus.Init()
	logger.Logrus.Info("dddd1111")
	err()
}

func err() {
	panic("panic test")
}
