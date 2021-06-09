package glogp

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	logrus "github.com/sirupsen/logrus"
)

type LogFormat uint8

const (
	LogFormatNone LogFormat = iota
	LogFormatText
	LogFormatJson
	LogFormatCbor
)

type LogField string

const (
	LogFieldLevel     LogField = "level"
	LogFieldMemory    LogField = "mem"
	LogFieldMessage   LogField = "msg"
	LogFieldTimedelta LogField = "td"
	LogFieldTrace     LogField = "trace"
	LogFieldExecId    LogField = "eid"
	LogFieldBinary    LogField = "bin"
)

const (
	LogTextColorNone       = "\033[0m"
	LogTextColorBlack      = "\033[0;30m"
	LogTextColorRed        = "\033[0;31m"
	LogTextColorGreen      = "\033[0;32m"
	LogTextColorYellow     = "\033[0;33m"
	LogTextColorBlue       = "\033[0;34m"
	LogTextColorPurple     = "\033[0;35m"
	LogTextColorCyan       = "\033[0;36m"
	LogTextColorWhite      = "\033[0;37m"
	LogTextColorNoneBold   = "\033[1m"
	LogTextColorBlackBold  = "\033[1;30m"
	LogTextColorRedBold    = "\033[1;31m"
	LogTextColorGreenBold  = "\033[1;32m"
	LogTextColorYellowBold = "\033[1;33m"
	LogTextColorBlueBold   = "\033[1;34m"
	LogTextColorPurpleBold = "\033[1;35m"
	LogTextColorCyanBold   = "\033[1;36m"
	LogTextColorWhiteBold  = "\033[1;37m"
)

type cborFormatter struct {
	format LogFormat
}

func (f cborFormatter) Format(e *logrus.Entry) ([]byte, error) {

	row := NewLogRow(e)

	bytes, err := cbor.Marshal(&row)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields to CBOR: %v", err)
	}

	return append(bytes, '\n'), nil
}

type jsonFormatter struct {
	format LogFormat
}

func (f jsonFormatter) Format(e *logrus.Entry) ([]byte, error) {

	row := NewLogRow(e)

	bytes, err := json.Marshal(&row)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON: %v", err)
	}

	return append(bytes, '\n'), nil
}

type textFormatter struct {
	format LogFormat
}

func (f textFormatter) Format(e *logrus.Entry) ([]byte, error) {

	row := NewLogRow(e)
	var str string
	switch e.Level {
	case logrus.ErrorLevel:
		str += LogTextColorRedBold
	case logrus.WarnLevel:
		str += LogTextColorYellowBold
	case logrus.DebugLevel:
		str += LogTextColorCyan
	default:
		str += LogTextColorNone
	}
	str = fmt.Sprintf("%s%-5s mem=%-5s msg=\"", str, row.Level, row.Mem)

	switch e.Level {
	case logrus.ErrorLevel:
		str += LogTextColorRedBold
	case logrus.WarnLevel:
		str += LogTextColorYellowBold
	case logrus.DebugLevel:
		str += LogTextColorCyanBold
	default:
		str += LogTextColorNoneBold
	}
	str += row.Msg

	switch e.Level {
	case logrus.ErrorLevel:
		str += LogTextColorRedBold
	case logrus.WarnLevel:
		str += LogTextColorYellowBold
	case logrus.DebugLevel:
		str += LogTextColorCyan
	default:
		str += LogTextColorNone
	}

	str = fmt.Sprintf("%s\"  td=%s  trace=[%s]  eid=%s  bin=%s  %s%s\n",
		str, row.Td, row.Trace, row.Eid, row.Bin, row.Ts, LogTextColorNone)

	return []byte(str), nil
}

func setFormat(logger *logrus.Logger, format LogFormat) {
	switch format {
	case LogFormatText:
		logger.SetFormatter(&textFormatter{format})
	case LogFormatJson:
		logger.SetFormatter(&jsonFormatter{format})
	case LogFormatCbor:
		logger.SetFormatter(&cborFormatter{format})
	}
}
