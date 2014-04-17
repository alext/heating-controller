package logger

import (
	"fmt"
	"io"
	"log"
	"os"
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

func SetDestination(dest interface{}) error {
	switch dest := dest.(type) {
	case io.Writer:
		log.SetOutput(dest)
	case string:
		if dest == "STDERR" {
			log.SetOutput(os.Stderr)
		} else if dest == "STDOUT" {
			log.SetOutput(os.Stdout)
		} else {
			file, err := os.OpenFile(dest, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
			if err != nil {
				return err
			}
			log.SetOutput(file)
		}
	default:
		return fmt.Errorf("Invalid log destination %T(%v)", dest, dest)
	}
	return nil
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