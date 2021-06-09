package glogp

import (
	logrus "github.com/sirupsen/logrus"
)

type LogRow struct {
	Level string `json:"level" cbor:"level"`
	Mem   string `json:"mem" cbor:"mem"`
	Msg   string `json:"msg" cbor:"msg"`
	Td    string `json:"td" cbor:"td"`
	Trace string `json:"trace" cbor:"trace"`
	Eid   string `json:"eid" cbor:"eid"`
	Bin   string `json:"bin" cbor:"bin"`
	Ts    string `json:"ts" cbor:"ts"`
}

func NewLogRow(e *logrus.Entry) *LogRow {
	row := &LogRow{
		Msg: e.Message,
		Ts:  e.Time.UTC().Format(LogTimeLayout),
	}
	switch e.Level {
	case logrus.DebugLevel:
		row.Level = "debug"
	case logrus.InfoLevel:
		row.Level = "info"
	case logrus.WarnLevel:
		row.Level = "warn"
	case logrus.ErrorLevel:
		row.Level = "error"
	}
	if val, ok := e.Data[string(LogFieldMemory)]; ok {
		row.Mem = val.(string)
	}
	if val, ok := e.Data[string(LogFieldTimedelta)]; ok {
		row.Td = val.(string)
	}
	if val, ok := e.Data[string(LogFieldTrace)]; ok {
		row.Trace = val.(string)
	}
	if val, ok := e.Data[string(LogFieldExecId)]; ok {
		row.Eid = val.(string)
	}
	if val, ok := e.Data[string(LogFieldBinary)]; ok {
		row.Bin = val.(string)
	}
	return row
}
