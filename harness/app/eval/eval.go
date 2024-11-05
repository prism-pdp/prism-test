package eval

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type EvalProcTime struct {
	Series []int64
	Mean float64

	start time.Time
}

type EvalProcTimeReport struct {
    ProcTime map[string]*EvalProcTime
	TargetMsg string
	LogFilePrefix string
}

type EvalReport struct {
	ProcTimeReport map[string]*EvalProcTimeReport
}

func NewEvalReport() *EvalReport {
	obj := new(EvalReport)
	obj.ProcTimeReport = make(map[string]*EvalProcTimeReport)
	return obj
}

func (this *EvalReport) SetupReport(_key string, _logFilePrefix string, _msg string) {
	this.ProcTimeReport[_key] = new(EvalProcTimeReport)
	this.ProcTimeReport[_key].ProcTime = make(map[string]*EvalProcTime)
	this.ProcTimeReport[_key].TargetMsg = _msg
	this.ProcTimeReport[_key].LogFilePrefix = _logFilePrefix
}

func (this *EvalProcTimeReport) Run(_pathLogDir string) {

	files, err := os.ReadDir(_pathLogDir)
	if err != nil { panic(err) }

	for _, file := range files {
		if false == strings.HasPrefix(file.Name(), this.LogFilePrefix) {
			continue
		}

		filePath := filepath.Join(_pathLogDir, file.Name())

		f, err := os.Open(filePath)
		if err != nil { panic(err) }
		defer f.Close()

		this.ProcTime[file.Name()] = new(EvalProcTime)

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			this.ProcTime[file.Name()].CalcDuration(this.TargetMsg, line)
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		this.ProcTime[file.Name()].CalcMean()
	}
}

func (this *EvalProcTime) CalcDuration(_targetMsg string, _log string) {
	datetime, msg, _ := helper.ParseLog(_log)

	match, start := checkMsg(_targetMsg, msg)
	if match {
		err := this.update(start, datetime)
		if err != nil { panic(err) }
	}
}

func (this *EvalProcTime) CalcMean() {
	var sum int64 = 0
	var count int64 = int64(len(this.Series))
	for _, v := range this.Series {
		sum += v
	}
	this.Mean = float64(sum) / float64(count)
}

func (this *EvalProcTime) update(_flagStart bool, _datetime time.Time) error {
	if this.start.IsZero() && _flagStart {
		this.start = _datetime
	} else if !this.start.IsZero() && !_flagStart {
		diff := _datetime.UnixMilli() - this.start.UnixMilli()
		this.Series = append(this.Series, diff)

		this.start = time.Time{} // zero clear
	} else {
		return fmt.Errorf("Invalid log sequences")
	}
	return nil
}

func checkMsg(_expected string, _actual string) (bool, bool) {
	match := false
	flagStart := false
	if "Start " + _expected == _actual {
		match = true
		flagStart = true
	} else if "Finish " + _expected == _actual {
		match = true
	}
	return match, flagStart
}