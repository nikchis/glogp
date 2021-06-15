package glogp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	logrus "github.com/sirupsen/logrus"
)

const LogPrinterDefFileDir = "../var/log"

type CustomLogger interface {
	Debug() *log.Logger
	Info() *log.Logger
	Warn() *log.Logger
	Error() *log.Logger
	Close()
}

type Printer struct {
	sync.RWMutex
	format         LogFormat
	level          LogLevel
	ts             int64
	file           *os.File
	clc            CustomLogger
	logFile        *logrus.Logger
	logStdout      *logrus.Logger
	logCustomDebug *logrus.Logger
	logCustomInfo  *logrus.Logger
	logCustomWarn  *logrus.Logger
	logCustomError *logrus.Logger
	wg             sync.WaitGroup
	stopRotation   chan bool
	closed         bool
}

func New(level LogLevel) *Printer {
	lFile, f, err := initLogFile(LogFormatJson, level, LogPrinterDefFileDir)
	if err != nil {
		return nil
	}
	lStdout, err := initLogStdout(LogFormatText, level)
	if err != nil {
		return nil
	}
	result := &Printer{format: LogFormatJson, level: level,
		logStdout: lStdout, file: f, logFile: lFile}
	result.stopRotation = make(chan bool, 1)
	result.startRotation()

	return result
}

func NewFile(format LogFormat, level LogLevel, dir string) (*Printer, error) {
	lFile, f, err := initLogFile(format, level, dir)
	if err != nil {
		return nil, err
	}
	result := &Printer{format: format, level: level, file: f, logFile: lFile}
	result.stopRotation = make(chan bool, 1)
	result.startRotation()

	return result, nil
}

func NewStd(format LogFormat, level LogLevel) (*Printer, error) {
	lStdout, err := initLogStdout(format, level)
	if err != nil {
		return nil, err
	}
	result := &Printer{format: format, level: level, logStdout: lStdout}

	return result, nil
}

func NewCustom(format LogFormat, level LogLevel, clc CustomLogger) (*Printer, error) {
	if clc == nil {
		return nil, fmt.Errorf("custom logger client is nil")
	}
	lCustomDebug, err := initLogCustom(format, level, clc.Debug())
	if err != nil {
		return nil, err
	}
	lCustomInfo, err := initLogCustom(format, level, clc.Info())
	if err != nil {
		return nil, err
	}
	lCustomWarn, err := initLogCustom(format, level, clc.Warn())
	if err != nil {
		return nil, err
	}
	lCustomError, err := initLogCustom(format, level, clc.Error())
	if err != nil {
		return nil, err
	}
	result := &Printer{
		format:         format,
		clc:            clc,
		level:          level,
		logCustomDebug: lCustomDebug,
		logCustomInfo:  lCustomInfo,
		logCustomWarn:  lCustomWarn,
		logCustomError: lCustomError,
	}

	return result, nil
}

func NewCustomAndStd(format LogFormat, level LogLevel, clc CustomLogger) (*Printer, error) {
	if clc == nil {
		return nil, fmt.Errorf("custom logger client is nil")
	}
	lCustomDebug, err := initLogCustom(format, level, clc.Debug())
	if err != nil {
		return nil, err
	}
	lCustomInfo, err := initLogCustom(format, level, clc.Info())
	if err != nil {
		return nil, err
	}
	lCustomWarn, err := initLogCustom(format, level, clc.Warn())
	if err != nil {
		return nil, err
	}
	lCustomError, err := initLogCustom(format, level, clc.Error())
	if err != nil {
		return nil, err
	}
	lStdout, err := initLogStdout(format, level)
	if err != nil {
		return nil, err
	}
	result := &Printer{
		clc:            clc,
		format:         format,
		level:          level,
		logStdout:      lStdout,
		logCustomDebug: lCustomDebug,
		logCustomInfo:  lCustomInfo,
		logCustomWarn:  lCustomWarn,
		logCustomError: lCustomError,
	}

	return result, nil
}

func (p *Printer) Debugf(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	p.push(message, LogLevelDebug)
}

func (p *Printer) Infof(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	p.push(message, LogLevelInfo)
}

func (p *Printer) Warnf(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	p.push(message, LogLevelWarn)
}

func (p *Printer) Errorf(format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	p.push(message, LogLevelError)
}

func (p *Printer) FileFormat() LogFormat {

	if p.logFile != nil {
		switch v := p.logFile.Formatter.(type) {
		case jsonFormatter:
			return v.format
		case textFormatter:
			return v.format
		case cborFormatter:
			return v.format
		}
	}
	return LogFormatNone
}

func (p *Printer) Level() LogLevel {
	return p.level
}

func (p *Printer) setFile(dir string) error {
	lFile, f, err := initLogFile(p.format, p.level, dir)
	if err != nil {
		return err
	}
	p.file = f
	p.logFile = lFile

	return nil
}

func (p *Printer) Close() {
	if p.closed {
		return
	}
	if p.clc != nil {
		p.clc.Close()
		p.clc = nil
	}
	if p.file != nil {
		p.stopRotation <- true
		p.wg.Wait()
		p.closeFile()
	}
	p.logStdout = nil
	p.closed = true
}

func (p *Printer) closeFile() {
	if p.logFile != nil {
		p.logFile.Writer().Close()
		p.logFile = nil
	}
	if p.file != nil {
		err := p.file.Close()
		if err != nil {
			msg := fmt.Sprintf("failed to close file: %v", err)
			if p.logStdout != nil {
				p.Errorf(msg)
			} else {
				fmt.Println(msg)
			}
		}
		p.file = nil
	}
}

func (p *Printer) push(message string, level LogLevel) {
	var srcExValue, srcInValue string

	if level > p.level {
		return
	}

	fields := logrus.Fields{}
	_, path, line, ok := runtime.Caller(4)
	if ok {
		paths := strings.Split(path, "/")
		if len(paths) > 1 {
			srcFileName := fmt.Sprintf("%s/%s", paths[len(paths)-2], paths[len(paths)-1])
			srcLineName := strconv.Itoa(line)
			srcExValue = fmt.Sprintf("%s:%s", srcFileName, srcLineName)
		}
	}
	_, path, line, ok = runtime.Caller(3)
	if ok {
		paths := strings.Split(path, "/")
		if len(paths) > 1 {
			srcFileName := fmt.Sprintf("%s/%s", paths[len(paths)-2], paths[len(paths)-1])
			srcLineName := strconv.Itoa(line)
			srcInValue = fmt.Sprintf("%s:%s", srcFileName, srcLineName)
		}
	}

	fields[string(LogFieldTrace)] = fmt.Sprintf("%s,%s", srcExValue, srcInValue)
	fields[string(LogFieldExecId)] = strings.ToUpper(strconv.FormatInt(LogExecId, 36))
	fields[string(LogFieldBinary)] = filepath.Base(os.Args[0])

	var tdelta_sec, tdelta_msec int64
	var ts int64 = time.Now().UTC().UnixNano()
	if p.ts > 0 {
		tdelta_sec = (ts - p.ts) / 1000000000
		tdelta_msec = ((ts - p.ts) % 1000000000) / 1000000
	}
	p.ts = ts
	fields[string(LogFieldTimedelta)] = fmt.Sprintf("%d.%03ds", tdelta_sec, tdelta_msec)

	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	const GB = 1073741824
	const MB = 1048576
	const KB = 1024
	switch {
	case m.Sys >= 10*GB:
		fields[string(LogFieldMemory)] = fmt.Sprintf("%dGB", m.Sys/GB)
	case m.Sys >= 10*MB:
		fields[string(LogFieldMemory)] = fmt.Sprintf("%dMB", m.Sys/MB)
	case m.Sys >= 10*KB:
		fields[string(LogFieldMemory)] = fmt.Sprintf("%dKB", m.Sys/KB)
	default:
		fields[string(LogFieldMemory)] = fmt.Sprintf("%dB", m.Sys)
	}

	message = strings.TrimRight(message, " \n")
	if p.clc != nil {
		switch level {
		case LogLevelDebug:
			p.logCustomInfo.WithFields(fields).Debug(message)
		case LogLevelInfo:
			p.logCustomInfo.WithFields(fields).Info(message)
		case LogLevelWarn:
			p.logCustomWarn.WithFields(fields).Warn(message)
		case LogLevelError:
			p.logCustomError.WithFields(fields).Error(message)
		}
	}
	if p.logFile != nil {
		switch level {
		case LogLevelDebug:
			p.logFile.WithFields(fields).Debug(message)
		case LogLevelInfo:
			p.logFile.WithFields(fields).Info(message)
		case LogLevelWarn:
			p.logFile.WithFields(fields).Warn(message)
		case LogLevelError:
			p.logFile.WithFields(fields).Error(message)
		}
	}
	if p.logStdout != nil {
		switch level {
		case LogLevelDebug:
			p.logStdout.WithFields(fields).Debug(message)
		case LogLevelInfo:
			p.logStdout.WithFields(fields).Info(message)
		case LogLevelWarn:
			p.logStdout.WithFields(fields).Warn(message)
		case LogLevelError:
			p.logStdout.WithFields(fields).Error(message)
		}
	}
}

func (p *Printer) startRotation() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-time.After(time.Second * 30):
				if fi, err := p.file.Stat(); err == nil {
					if fi.Size() > LogSizeMax {
						fpath := p.file.Name()
						fpathtmp := fmt.Sprintf("%s.1", fpath)
						if err := rotateLogArchives(fpath); err != nil {
							fpathtmp = fmt.Sprintf("%s.%s",
								fpath, time.Now().Format("2006-01-02.15-04"))
						}
						if _, err := os.Stat(fpathtmp); err == nil {
							if err = gzipLogFile(fpathtmp); err != nil {
								fmt.Println(err.Error())
								//Print.Errorf()
							} else {
								os.Remove(fpathtmp)
							}
						} else if os.IsNotExist(err) {
							p.Lock()
							p.closeFile()
							err = os.Rename(fpath, fpathtmp)
							p.setFile(filepath.Dir(fpath))
							p.Unlock()
							if err == nil {
								if err = gzipLogFile(fpathtmp); err != nil {
									fmt.Println(err.Error())
									//Print.Errorf(err.Error())
								} else {
									os.Remove(fpathtmp)
								}
							}
						}
					}
				}
			case <-p.stopRotation:
				close(p.stopRotation)
				return
			}
		}
	}()
}
