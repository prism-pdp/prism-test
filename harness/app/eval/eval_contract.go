package eval

import (
	"bufio"
	"encoding/csv"
    "encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type EvalContract struct {
	Name string
	Series []int64
	Mean float64
	StdDev float64
}

type EvalContractReport struct {
	Name string // Log's filename
	EvalDataGasUsed map[string]*EvalContract
	EvalDataGasPrice map[string]*EvalContract
}

type EvalContractReportBundle struct {
	PathLogDir string
	PathResultDir string
	Reports []*EvalContractReport
}

func NewEvalContract(_name string) *EvalContract {
	obj := new(EvalContract)
	obj.Name = _name
	return obj
}

func NewEvalContractReport(_name string) *EvalContractReport {
	obj := new(EvalContractReport)
	obj.Name = _name
	obj.EvalDataGasUsed = make(map[string]*EvalContract)
	obj.EvalDataGasPrice = make(map[string]*EvalContract)
	return obj
}

func NewEvalContractReportBundle(_pathLogDir, _pathResultDir string) *EvalContractReportBundle {
	obj := new(EvalContractReportBundle)
	obj.PathLogDir = _pathLogDir
	obj.PathResultDir = _pathResultDir
	return obj
}

func (this *EvalContractReportBundle) Run() error {
	var err error

	dirEntries, err := os.ReadDir(this.PathLogDir)
	if err != nil { panic(err) }

	for _, e := range dirEntries {
		fileName := e.Name()

		if strings.HasPrefix(fileName, ".") {
			continue
		}

		filePath := filepath.Join(this.PathLogDir, fileName)

		f, err := os.Open(filePath)
		if err != nil { panic(err) }
		defer f.Close()

		repo := NewEvalContractReport(fileName)
		err = repo.Run(f)
		if err != nil { return err }

		this.Reports = append(this.Reports, repo)
	}

	return nil
}

func (this *EvalContractReport) Run(_file *os.File) error {
	scanner := bufio.NewScanner(_file)

	for scanner.Scan() {
		line := scanner.Text()
		_, msg, detail := helper.ParseLog(line)

		if msg == "Completed smart contract" {
			nameContract := detail["name"]
			if _, ok := this.EvalDataGasUsed[nameContract]; !ok {
				this.EvalDataGasUsed[nameContract] = NewEvalContract(nameContract)
			}
			if _, ok := this.EvalDataGasPrice[nameContract]; !ok {
				this.EvalDataGasPrice[nameContract] = NewEvalContract(nameContract)
			}
			e1 := this.EvalDataGasUsed[nameContract]
			e2 := this.EvalDataGasPrice[nameContract]

			gasUsed, err := strconv.ParseInt(detail["gasUsed"], 10, 64)
			if err != nil { return err }
			e1.Series = append(e1.Series, gasUsed)

			gasPrice, err := strconv.ParseInt(detail["gasPrice"], 10, 64)
			if err != nil { return err }
			e2.Series = append(e2.Series, gasPrice)
		}
	}

	for _, v1 := range this.EvalDataGasUsed {
		v1.CalcMean()
		v1.CalcStandardDeviation()
	}

	return nil
}

func (this *EvalContractReportBundle) Dump() error {
	for _, r := range this.Reports {
		err := r.Dump(this.PathResultDir)
		if err != nil { return err }
	}
	return nil
}

func (this *EvalContractReport) Dump(_pathDir string) error {
	var err error

	if err = this.DumpJson(_pathDir); err != nil {
		return err
	}

	if err = this.DumpCsv(_pathDir); err != nil {
		return err
	}

	return nil
}

func (this *EvalContractReport) DumpJson(_pathDir string) error {
	filePath := filepath.Join(_pathDir, "contract.json")

	tmp, err := json.MarshalIndent(this, "", "\t")
	if err != nil { return err }

	helper.WriteFile(filePath, tmp)

	return nil
}

func (this *EvalContractReport) DumpCsv(_pathDir string) error {
	filePath := filepath.Join(_pathDir, "contract.csv")
	target := []string{
		"RegisterFile",
		"AppendOwner",
		"SetChal",
		"SetProof",
		"SetAuditingResult",
	}

	file, err := os.Create(filePath)
	if err != nil { panic(err) }
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var header []string
	header = append(header, "Contract")
	for i, _ := range this.EvalDataGasUsed[target[0]].Series {
		header = append(header, strconv.Itoa(i+1))
	}
	header = append(header, "Mean")
	header = append(header, "Standard Deviation")
	if err := writer.Write(header); err != nil {
		return err
	}

	var records [][]string
	for _, v1 := range target {
		e := this.EvalDataGasUsed[v1]
		var r []string
		r = append(r, v1)
		for _, v2 := range e.Series {
			r = append(r, strconv.FormatInt(v2, 10))
		}
		r = append(r, strconv.FormatFloat(e.Mean, 'f', -1, 64))
		r = append(r, strconv.FormatFloat(e.StdDev, 'f', -1, 64))
		records = append(records, r)
	}

	return writer.WriteAll(records)
}

func (this *EvalContract) CalcMean() {
	this.Mean = helper.CalcMean(this.Series)
}

func (this *EvalContract) CalcStandardDeviation() {
	this.StdDev = helper.CalcStandardDeviation(this.Series, this.Mean)
}
