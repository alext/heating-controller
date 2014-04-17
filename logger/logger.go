package logger

import (
	"fmt"
	"io"
	"log"
)

const (
	DEBUG int = iota
	INFO
	WARN
)

var (
	Level = INFO
)

func outputAbove(level int, msg ...interface{}) {
	if level >= Level {
		log.Println(msg...)
	}
}
func outputAbovef(level int, format string, v ...interface{}) {
	if level >= Level {
		log.Println(fmt.Sprintf(format, v...))
	}
}

func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func Debug(msg ...interface{}) {
	outputAbove(DEBUG, msg...)
}
func Debugf(format string, v ...interface{}) {
	outputAbovef(DEBUG, format, v...)
}

func Info(msg ...interface{}) {
	outputAbove(INFO, msg...)
}
func Infof(format string, v ...interface{}) {
	outputAbovef(INFO, format, v...)
}

func Warn(msg ...interface{}) {
	outputAbove(WARN, msg...)
}
func Warnf(format string, v ...interface{}) {
	outputAbovef(WARN, format, v...)
}

func Fatal(msg ...interface{}) {
	log.Fatalln(msg...)
}
func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}
