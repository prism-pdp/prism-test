package eval

import (
	"bufio"
	// "encoding/csv"
    "encoding/json"
	// "fmt"
	// "math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	// "time"

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
	tmp, err := json.MarshalIndent(this, "", "\t")
	if err != nil { return err }

	filePath := filepath.Join(_pathDir, this.Name)
	helper.WriteFile(filePath, tmp)

	return nil
}
