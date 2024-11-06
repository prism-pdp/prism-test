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

	start time.Time
}

type EvalProcTimeReport struct {
	TargetMsg string
	ReportName string
	PathLogDir string
	PathResultDir string
    ProcTime []*EvalProcTime
}

type EvalReport struct {
	ProcTimeReport []*EvalProcTimeReport
}

func NewEvalReport() *EvalReport {
	obj := new(EvalReport)
	return obj
}

func (this *EvalReport) SetupReport(_key string, _msg string, _pathLogDir string, _pathResultDir string) {
	obj := new(EvalProcTimeReport)
	obj.TargetMsg = _msg
	obj.ReportName = _key
	obj.PathLogDir = _pathLogDir
	obj.PathResultDir = _pathResultDir
	this.ProcTimeReport = append(this.ProcTimeReport, obj)
}

func (this *EvalReport) Run() {
	for _, v := range this.ProcTimeReport {
		v.Run()
	}
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

		this.ProcTime = append(this.ProcTime, evalProcTime)
	}
}

func (this *EvalReport) Dump() error {
	for _, v := range this.ProcTimeReport {
		err := v.Dump()
		if err != nil { return err }
	}
	return nil
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
	for i, _ := range this.ProcTime {
		header = append(header, strconv.Itoa(i+1))
	}
	header = append(header, "Mean")

	if err := writer.Write(header); err != nil {
		return err
	}

	var records [][]string
	for _, v1 := range this.ProcTime {
		var r []string
		r = append(r, strconv.Itoa(v1.BlockNum))
		for _, v2 := range v1.Series {
			r = append(r, strconv.FormatInt(v2, 10))
		}
		r = append(r, strconv.FormatFloat(v1.Mean, 'f', -1, 64))
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

func getBlockNum(_filename string) (int, error) {
	if strings.HasPrefix(_filename, "gentags") {
		return strconv.Atoi(_filename[8:12])
	}
	return 0, fmt.Errorf("Invalid filename (%s)", _filename)
}