package eval

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type EvalFrequency struct {
	DataRatio float64
	FileRatio float64
	DamageRate float64

	CorruptedFileList []string

	TotalCorruptedFileCount int
	TotalRepairedFileCount int
	TotalUnrepairedFileCount int

	HistoryCorruptedFileCount []int
}

type EvalFrequencyReport struct {
	PathLogDir string
	PathResultDir string

	EvalData map[string][]*EvalFrequency
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

func NewEvalFrequencyReport(_pathLogDir, _pathResultDir string) *EvalFrequencyReport {
	obj := new(EvalFrequencyReport)
	obj.PathLogDir = _pathLogDir
	obj.PathResultDir = _pathResultDir
	obj.EvalData = make(map[string][]*EvalFrequency)
	return obj
}

func (this *EvalFrequencyReport) Run() error {

	dirEntries, err := os.ReadDir(this.PathLogDir)
	if err != nil { panic(err) }

	for _, e := range dirEntries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		filePath := filepath.Join(this.PathLogDir, e.Name())

		lines, err := helper.ReadLines(filePath)
		if err != nil { return err }

		lineCount := len(lines)

		for i := 0; i < lineCount; i++ {
			l := lines[i]

			_, message, detail := helper.ParseLog(l)
			if message == "Start frequency evaluation" {
				r1 := helper.ToFloat64(detail["DataRatio"])
				r2 := helper.ToFloat64(detail["FileRatio"])
				r3 := helper.ToFloat64(detail["DamageRate"])
				e := NewEvalFrequency(r1, r2, r3)

				i += 1
				stepSize := this.runCore(i, lines, e, (detail["DataRatio"] == "0.2" && detail["FileRatio"] == "0.1"))
				if err != nil { return err }
				if stepSize < 0 { return fmt.Errorf("Invalid log lines") }

				// this.EvalData = append(this.EvalData, e)
				this.EvalData[detail["FileRatio"]] = append(this.EvalData[detail["FileRatio"]], e)
				i += stepSize
			}
		}
	}

	return nil
}

func (this *EvalFrequencyReport) runCore(_startIndex int, _lines []string, _e *EvalFrequency, _debug bool) int {
	lineCount := len(_lines)

	// save := 0
	for i := _startIndex; i < lineCount; i++ {
		l := _lines[i]
		_, message, _ := helper.ParseLog(l)
		if message == "Start cycle" {
			stepSize, corruptedFileList, repairedFileList := this.runCycle(i, _lines)
			if stepSize < 0 {
				panic(fmt.Errorf("Invalid log lines"))
			} else {
				i += stepSize
			}
			newCorruptedFileList := helper.SubSlices(corruptedFileList, _e.CorruptedFileList)
			_e.TotalCorruptedFileCount += len(newCorruptedFileList)
			_e.TotalRepairedFileCount += len(repairedFileList)

			_e.CorruptedFileList = helper.Uniq(append(_e.CorruptedFileList, corruptedFileList...))
			_e.CorruptedFileList = helper.SubSlices(_e.CorruptedFileList, repairedFileList)
			_e.HistoryCorruptedFileCount = append(_e.HistoryCorruptedFileCount, len(_e.CorruptedFileList))

		} else if message == "Finish frequency evaluation" {
			_e.TotalUnrepairedFileCount = len(_e.CorruptedFileList)
			return i - _startIndex
		}
	}

	panic(fmt.Errorf("Invalid sequence"))
}

func (this *EvalFrequencyReport) runCycle(_startIndex int, _lines []string) (int, []string, []string) {
	lineCount := len(_lines)

	var corruptedFileList []string
	var repairedFileList []string

	var i int
	for i = _startIndex; i < lineCount; i++ {
		l := _lines[i]

		_, message, detail := helper.ParseLog(l)

		if message == "File corruption" {
			corruptedFileList = append(corruptedFileList, detail["file"])
		} else if message == "Repair" {
			repairedFileList = append(repairedFileList, detail["file"])
		} else if message == "Finish cycle" {
			corruptedFileList = helper.Uniq(corruptedFileList)
			repairedFileList = helper.Uniq(repairedFileList)
			return i - _startIndex, corruptedFileList, repairedFileList
		}
	}

	panic(fmt.Errorf("Invalid sequence"))
}

func (this *EvalFrequencyReport) Dump() error {
	var err error

	if err = this.DumpJson(this.PathResultDir); err != nil {
		return err
	}

	if err = this.DumpCsv(this.PathResultDir); err != nil {
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

func (this *EvalFrequencyReport) DumpCsv(_pathDir string) error {
	filePath := filepath.Join(_pathDir, "frequency.csv")

	file, err := os.Create(filePath)
	if err != nil { panic(err) }
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{ "File Ratio", "Block Ratio" }
	for _, v := range this.EvalData {
		for i, _ := range v[0].HistoryCorruptedFileCount {
			header = append(header, strconv.Itoa(i))
		}
		break
	}

	var records [][]string
	keys := helper.SortedMapKeys(this.EvalData)
	for _, k := range keys {
		for _, v := range this.EvalData[k] {
			fr := fmt.Sprintf("%.1f", v.FileRatio)
			dr := fmt.Sprintf("%.1f", v.DataRatio)
			r := []string{ fr, dr }
			for _, c := range v.HistoryCorruptedFileCount {
				r = append(r, strconv.Itoa(c))
			}
			records = append(records, r)
		}
	}

	if err := writer.Write(header); err != nil {
		return err
	}
	return writer.WriteAll(records)
}
