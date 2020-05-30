package easy_circuitbreaker

import "log"

type LevelLogger interface {
	Debug(msg string)
	Debugf(format string, v ...interface{})
	Info(msg string)
	Infof(format string, v ...interface{})
	Warn(msg string)
	Warnf(format string, v ...interface{})
	Error(msg string)
	Errorf(format string, v ...interface{})
}

type DefaultLogger struct{}

func (logger DefaultLogger) Debug(msg string) {
	log.Println(msg)
}

func (logger DefaultLogger) Debugf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (logger DefaultLogger) Info(msg string) {
	log.Println(msg)
}

func (logger DefaultLogger) Infof(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (logger DefaultLogger) Warn(msg string) {
	log.Println(msg)
}

func (logger DefaultLogger) Warnf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (logger DefaultLogger) Error(msg string) {
	log.Println(msg)
}

func (logger DefaultLogger) Errorf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
