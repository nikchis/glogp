package glogp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	logrus "github.com/sirupsen/logrus"
)

const LogTimeLayout = "2006-01-02T15:04:05.000"

var LogExecId int64 = time.Now().UTC().UnixNano()
var LogSizeMax int64 = 16777216 // 16MB
var LogArchivesMax int = 9

// Printer for general purpose tasks {Debug, Info, Warn, Error, ...}
var Print *Printer = New(LogLevelDebug)

func Debugf(format string, a ...interface{}) {
	if Print != nil {
		Print.Debugf(format, a...)
	}
}

func Infof(format string, a ...interface{}) {
	if Print != nil {
		Print.Infof(format, a...)
	}
}

func Warnf(format string, a ...interface{}) {
	if Print != nil {
		Print.Warnf(format, a...)
	}
}

func Errorf(format string, a ...interface{}) {
	if Print != nil {
		Print.Errorf(format, a...)
	}
}

func initLogFile(format LogFormat, level LogLevel, dir string) (*logrus.Logger, *os.File, error) {
	return initLogFileWithName(format, level, dir, filepath.Base(os.Args[0]))
}

func initLogFileWithName(format LogFormat, level LogLevel,
	dir, fname string) (*logrus.Logger, *os.File, error) {
	if err := mkdirForce(dir); err != nil {
		return nil, nil, err
	}
	fpath := fmt.Sprintf("%s/%s.log", dir, fname)
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, err
	}
	lFile := logrus.New()
	lFile.Out = f
	setFormat(lFile, format)
	setLevel(lFile, level)

	return lFile, f, nil
}

func initLogCustom(format LogFormat, level LogLevel, stdLog *log.Logger) (*logrus.Logger, error) {
	lCustom := logrus.New()
	lCustom.SetOutput(stdLog.Writer())
	setFormat(lCustom, format)
	setLevel(lCustom, level)

	return lCustom, nil
}

func initLogStdout(format LogFormat, level LogLevel) (*logrus.Logger, error) {
	lStdout := logrus.New()
	lStdout.Out = os.Stdout
	setFormat(lStdout, format)
	setLevel(lStdout, level)

	return lStdout, nil
}
