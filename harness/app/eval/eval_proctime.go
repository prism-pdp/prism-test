package eval

import (
	"bufio"
	"encoding/csv"
    "encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type EvalProcTime struct {
	Name string
	BlockNum int
	Series []int64
	Mean float64
	StdDev float64

	start time.Time
}

type EvalProcTimeReport struct {
	ReportName string
	TargetMsg string
	PathLogDir string
	PathResultDir string
    ProcTime []*EvalProcTime
}

func NewEvalProcTimeReport(_reportName, _targetMsg, _pathLogDir, _pathResultDir string) *EvalProcTimeReport {
	obj := new(EvalProcTimeReport)
	obj.ReportName = _reportName
	obj.TargetMsg = _targetMsg
	obj.PathLogDir = _pathLogDir
	obj.PathResultDir = _pathResultDir
	return obj
}

func (this *EvalProcTimeReport) Run() {
	var err error

	dirEntries, err := os.ReadDir(this.PathLogDir)
	if err != nil { panic(err) }

	for _, e := range dirEntries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		filePath := filepath.Join(this.PathLogDir, e.Name())

		f, err := os.Open(filePath)
		if err != nil { panic(err) }
		defer f.Close()

		evalProcTime := new(EvalProcTime)
		evalProcTime.Name = e.Name()
		evalProcTime.BlockNum, err = getBlockNum(e.Name())
		if err != nil { panic(err) }

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			evalProcTime.CalcDuration(this.TargetMsg, line)
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		evalProcTime.CalcMean()

		evalProcTime.CalcStandardDeviation()

		this.ProcTime = append(this.ProcTime, evalProcTime)
	}
}

func (this *EvalProcTimeReport) Dump() error {
	var err error

	err = this.DumpJson(this.PathResultDir)
	if err != nil { return err }

	err = this.DumpCsv(this.PathResultDir)
	if err != nil { return err }

	return nil
}

func (this *EvalProcTimeReport) DumpJson(_pathDir string) error {
    tmp, err := json.MarshalIndent(this, "", "\t")
	if err != nil { return err }
	filePath := filepath.Join(_pathDir, this.ReportName + ".json")
	helper.WriteFile(filePath, tmp)

	return nil
}

func (this *EvalProcTimeReport) DumpCsv(_pathDir string) error {
	filePath := filepath.Join(_pathDir, this.ReportName + ".csv")

	file, err := os.Create(filePath)
	if err != nil { panic(err) }
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var header []string
	header = append(header, "Blocks")
	header = append(header, "Mean")
	header = append(header, "Standard Deviation")
	for i, _ := range this.ProcTime[0].Series {
		header = append(header, strconv.Itoa(i+1))
	}

	if err := writer.Write(header); err != nil {
		return err
	}

	var records [][]string
	for _, v1 := range this.ProcTime {
		var r []string
		r = append(r, strconv.Itoa(v1.BlockNum))
		r = append(r, strconv.FormatFloat(v1.Mean, 'f', -1, 64))
		r = append(r, strconv.FormatFloat(v1.StdDev, 'f', -1, 64))
		for _, v2 := range v1.Series {
			r = append(r, strconv.FormatInt(v2, 10))
		}
		records = append(records, r)
	}

	return writer.WriteAll(records)
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
	this.Mean = helper.CalcMean(this.Series)
}

func (this *EvalProcTime) CalcStandardDeviation() {
	this.StdDev = helper.CalcStandardDeviation(this.Series, this.Mean)
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

func getBlockNum(_filename string) (int, error) {
	if strings.HasPrefix(_filename, "gentags") {
		return strconv.Atoi(_filename[8:12])
	} else if strings.HasPrefix(_filename, "auditing") {
		val, err := strconv.ParseFloat(_filename[9:12], 32)
		if err != nil { return 0, err }
		return int(val * 1000.0), nil
	}
	return 0, fmt.Errorf("Invalid filename (%s)", _filename)
}