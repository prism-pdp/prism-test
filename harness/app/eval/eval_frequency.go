package eval

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type EvalFrequency struct {
	DataRatio float64
	FileRatio float64
	DamageRate float64

	TotalCorruptedFileCount int
	TotalRepairedFileCount int
	TotalUnrepairedFileCount int
}

type EvalFrequencyReport struct {
	PathLogFile string
	PathResultDir string

	EvalData []*EvalFrequency
}

func NewEvalFrequency(_dataRatio, _fileRatio, _damageRate float64) *EvalFrequency {
	obj := new(EvalFrequency)
	obj.DataRatio = _dataRatio
	obj.FileRatio = _fileRatio
	obj.DamageRate = _damageRate
	obj.TotalCorruptedFileCount = 0
	obj.TotalRepairedFileCount = 0
	obj.TotalUnrepairedFileCount = 0
	return obj
}

func NewEvalFrequencyReport(_pathLogFile, _pathResultDir string) *EvalFrequencyReport {
	obj := new(EvalFrequencyReport)
	obj.PathLogFile = _pathLogFile
	obj.PathResultDir = _pathResultDir
	return obj
}

func (this *EvalFrequencyReport) Run() error {
	lines, err := helper.ReadLines(this.PathLogFile)
	if err != nil { return err }

	lineCount := len(lines)

	for i := 0; i < lineCount; i++ {
		l := lines[i]

		fmt.Println(l)

		_, message, detail := helper.ParseLog(l)
		if message == "Start frequency evaluation" {
			r1 := helper.ToFloat64(detail["DataRatio"])
			r2 := helper.ToFloat64(detail["FileRatio"])
			r3 := helper.ToFloat64(detail["DamageRate"])
			e := NewEvalFrequency(r1, r2, r3)

			i += 1
			stepSize := this.runCore(i, lines, e)
			if stepSize < 0 { return fmt.Errorf("Invalid log lines") }
			this.EvalData = append(this.EvalData, e)
			i += stepSize
		}
	}

	return nil
}

func (this *EvalFrequencyReport) runCore(_startIndex int, _lines []string, _e *EvalFrequency) int {
	lineCount := len(_lines)

	var corruptedFileList []string

	var i int
	for i = _startIndex; i < lineCount; i++ {
		l := _lines[i]

		_, message, detail := helper.ParseLog(l)

		if message == "File corruption" {
			corruptedFileList = append(corruptedFileList, detail["file"])
			corruptedFileList = helper.Uniq(corruptedFileList)
		} else if message == "Repair" {
			corruptedFileList = helper.Remove(corruptedFileList, detail["file"])
		} else if message == "Finish frequency evaluation" {
			_e.TotalUnrepairedFileCount = len(corruptedFileList)
			return i - _startIndex
		}
	}

	return -1 // error
}

func (this *EvalFrequencyReport) Dump() error {
	var err error

	if err = this.DumpJson(this.PathResultDir); err != nil {
		return err
	}

	return nil
}

func (this *EvalFrequencyReport) DumpJson(_pathDir string) error {
	filePath := filepath.Join(_pathDir, "frequency.json")

	tmp, err := json.MarshalIndent(this, "", "\t")
	if err != nil { return err }

	helper.WriteFile(filePath, tmp)

	return nil
}