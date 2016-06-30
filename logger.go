package main

import (
	"github.com/kovetskiy/lorg"
)

const logFormat = `${level:%s\::right:false} ${time} ${prefix}%s`

func getLogger() lorg.Logger {
	logger := lorg.NewLog()
	logger.SetFormat(lorg.NewFormat(logFormat))
	logger.SetLevel(lorg.LevelDebug)

	return logger
}
