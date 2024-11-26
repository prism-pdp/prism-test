package eval

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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
	HistoryRepairedFileCount []int
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

		_, message, detail := helper.ParseLog(l)
		if message == "Start frequency evaluation" {
			r1 := helper.ToFloat64(detail["DataRatio"])
			r2 := helper.ToFloat64(detail["FileRatio"])
			r3 := helper.ToFloat64(detail["DamageRate"])
			e := NewEvalFrequency(r1, r2, r3)

			i += 1
			stepSize := this.runCore(i, lines, e)
			if err != nil { return err }
			if stepSize < 0 { return fmt.Errorf("Invalid log lines") }
			this.EvalData = append(this.EvalData, e)
			i += stepSize
		}
	}

	return nil
}

func (this *EvalFrequencyReport) runCore(_startIndex int, _lines []string, _e *EvalFrequency) int {
	lineCount := len(_lines)

	var saveCorruptedFileList []string
	var saveRepairedFileList []string

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
			saveCorruptedFileList = append(saveCorruptedFileList, newCorruptedFileList...)
			saveRepairedFileList = append(saveRepairedFileList, repairedFileList...)
			_e.HistoryCorruptedFileCount = append(_e.HistoryCorruptedFileCount, len(saveCorruptedFileList))
			_e.HistoryRepairedFileCount = append(_e.HistoryRepairedFileCount, len(saveRepairedFileList))

			_e.CorruptedFileList = helper.Uniq(append(_e.CorruptedFileList, corruptedFileList...))
			_e.CorruptedFileList = helper.SubSlices(_e.CorruptedFileList, repairedFileList)

			// repairedFileCount := len(repairedFileList) + save
			// _e.HistoryRepairedFileCount = append(_e.HistoryRepairedFileCount, repairedFileCount)
			// save = repairedFileCount
		} else if message == "Finish frequency evaluation" {
			_e.TotalCorruptedFileCount = _e.HistoryCorruptedFileCount[len(_e.HistoryCorruptedFileCount)-1]
			_e.TotalRepairedFileCount = _e.HistoryRepairedFileCount[len(_e.HistoryRepairedFileCount)-1]
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

	if err = this.DumpCsvHeatMap(this.PathResultDir); err != nil {
		return err
	}

	if err = this.DumpCsvHistory(this.PathResultDir); err != nil {
		return err
	}

	return nil
}

func (this *EvalFrequencyReport) DumpJson(_pathDir string) error {
	fileName := filepath.Base(this.PathLogFile)
	filePath := filepath.Join(_pathDir, fileName + ".json")

	tmp, err := json.MarshalIndent(this, "", "\t")
	if err != nil { return err }

	helper.WriteFile(filePath, tmp)

	return nil
}

func (this *EvalFrequencyReport) DumpCsvHistory(_pathDir string) error {
	fileName := filepath.Base(this.PathLogFile)
	filePath := filepath.Join(_pathDir, fileName + "-history.csv")

	file, err := os.Create(filePath)
	if err != nil { panic(err) }
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"Type", "File Ratio", "Block Ratio",
		 "1",  "2",  "3",  "4",  "5",  "6",  "7",  "8",  "9",  "10",
		"11", "12", "13", "14", "15", "16", "17", "18", "19",  "20",
		"21", "22", "23", "24", "25", "26", "27", "28", "29",  "30",
		"31", "32", "33", "34", "35", "36", "37", "38", "39",  "40",
		"41", "42", "43", "44", "45", "46", "47", "48", "49",  "50",
		"51", "52", "53", "54", "55", "56", "57", "58", "59",  "60",
		"61", "62", "63", "64", "65", "66", "67", "68", "69",  "70",
		"71", "72", "73", "74", "75", "76", "77", "78", "79",  "80",
		"81", "82", "83", "84", "85", "86", "87", "88", "89",  "90",
		"91", "92", "93", "94", "95", "96", "97", "98", "99", "100",
	}

	var records [][]string
	for _, v := range this.EvalData {
		fr := fmt.Sprintf("%.1f", v.FileRatio)
		dr := fmt.Sprintf("%.1f", v.DataRatio)
		r1 := []string{ "Corruption", fr, dr }
		r2 := []string{ "Repair", fr, dr }
		for _, c := range v.HistoryCorruptedFileCount {
			r1 = append(r1, strconv.Itoa(c))
		}
		for _, c := range v.HistoryRepairedFileCount {
			r2 = append(r2, strconv.Itoa(c))
		}
		records = append(records, r1)
		records = append(records, r2)
	}

	if err := writer.Write(header); err != nil {
		return err
	}
	return writer.WriteAll(records)
}

func (this *EvalFrequencyReport) DumpCsvHeatMap(_pathDir string) error {
	fileName := filepath.Base(this.PathLogFile)
	filePath := filepath.Join(_pathDir, fileName + "-heatmap.csv")

	file, err := os.Create(filePath)
	if err != nil { panic(err) }
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"FileRatio",
		"DataRatio:0.1",
		"DataRatio:0.2",
		"DataRatio:0.3",
		"DataRatio:0.4",
		"DataRatio:0.5",
		"DataRatio:0.6",
		"DataRatio:0.7",
		"DataRatio:0.8",
		"DataRatio:0.9",
		"DataRatio:1.0",
	}

	records := make([][]string, 10)
	for i := 0; i < len(records); i++ {
		records[i] = make([]string, len(header))
		records[i][0] = fmt.Sprintf("%.1f", float64(i + 1) * 0.1)
	}

	for _, v := range this.EvalData {
		dr := int(v.DataRatio * 10 - 1) + 1
		fr := int(v.FileRatio * 10 - 1)
		records[fr][dr] = helper.IntToString(v.TotalUnrepairedFileCount)
	}

	if err := writer.Write(header); err != nil {
		return err
	}
	return writer.WriteAll(records)
}