package main

import(
    "encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pborman/getopt/v2"
    "math/rand"
	"strconv"
    "os"
    "path/filepath"
    "time"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/entity"
	"github.com/dpduado/dpduado-test/harness/eval"
	"github.com/dpduado/dpduado-test/harness/helper"
)

var baseclient client.BaseClient

func runTestdata(_path string, _size string, _val string) {
    num, unitSize, err := helper.ParseSize(_size)
    if err != nil { panic(err) }

    tmpVal, err := strconv.ParseUint(_val, 10, 16)
    if err != nil { panic(err) }
    value := uint16(tmpVal)

    // TODO: WriteFileUint16
    buf := make([]byte, unitSize)
    for i:= 0; i < len(buf); i += 2 {
        binary.BigEndian.PutUint16(buf[i:i+2], value)
    }

    f, err := os.Create(_path)
    if err != nil { panic(err) }

    for range num {
        _, err := f.Write(buf)
        if err != nil { panic(err) }
    }

    err = f.Close()
    if err != nil { panic(err) }
}

func runCorruption(_pathDir string, _damageRate string, _pathOutput string) {
    var err error

    damageRate, err := strconv.ParseFloat(_damageRate, 64)
    if err != nil { panic(err) }

    rand.Seed(time.Now().UnixNano())

	dirEntries, err := os.ReadDir(_pathDir)
	if err != nil { panic(err) }

    var corruptedFilePathList []string
    if helper.IsFile(_pathOutput) {
        corruptedFilePathList, err = helper.ReadAllLine(_pathOutput)
        if err != nil { panic(err) }
    }

	for _, e := range dirEntries {
		fileName := e.Name()
        if filepath.Ext(fileName) != ".dat" {
            continue
        }

        r1 := rand.Float64()
        if damageRate < r1 {
            continue
        }

		filePath := filepath.Join(_pathDir, fileName)

        data, err := helper.ReadFile(filePath)
        if err != nil { panic(err) }

        r2 := rand.Float64()
        pos := uint32(r2 * float64(len(data) * 8))

        index, offset := helper.ToggleBit(data, pos)
        helper.PrintLog("File corruption (file:%s, index:%d, offset:%d)", fileName, index, offset)

        helper.WriteFile(filePath, data)

        corruptedFilePathList = append(corruptedFilePathList, filePath)
    }

    corruptedFilePathList = helper.Uniq(corruptedFilePathList)

    if len(corruptedFilePathList) > 0 {
        s := fmt.Sprintf("%s\n", corruptedFilePathList[0])
        helper.WriteFile(_pathOutput, []byte(s))
    }
    if len(corruptedFilePathList) > 1 {
        for i := 1; i < len(corruptedFilePathList); i++ {
            s := fmt.Sprintf("%s\n", corruptedFilePathList[i])
            helper.AppendFile(_pathOutput, []byte(s))
        }
    }
}

func runListCoruptedFiles(_name string) {
    *helper.SimFlag = true

	var tpa entity.Auditor
	helper.LoadEntity(_name, &tpa)

    list := tpa.ListCorruptedFiles()
    for _, v := range list {
        fmt.Printf("%s ", v)
    }
}

func runRepair(_path string) {
    data, err := helper.ReadFile(_path)
    if err != nil { panic(err) }

    val := helper.MostFrequentValue(data, 32)

    helper.WriteFileUint16(_path, 1, int64(len(data)), val)

    helper.PrintLog("Repair (file:%s)", filepath.Base(_path))
}

func runRepairBatch(_path string) {
    if helper.IsFile(_path) == false {
        return
    }

    lines, err := helper.ReadAllLine(_path)
    if err != nil { panic(err) }

    for _, f := range lines {
        runRepair(f)
    }
}

func runWriteLog(_log string) {
    helper.PrintLog(_log)
}

func runEvalGenTag(_pathLogDir string, _pathResultDir string) {
    evalReport := eval.NewEvalProcTimeReport("gentags", "generate tags", _pathLogDir, _pathResultDir)

    evalReport.Run()

    err := evalReport.Dump()
    if err != nil { panic(err) }
}

func runEvalAuditing(_pathLogDir string, _pathResultDir string) {
    var err error

    evalGenProofReport := eval.NewEvalProcTimeReport("genproof", "generating proof", _pathLogDir, _pathResultDir)
    evalVerProofReport := eval.NewEvalProcTimeReport("verifyproof", "verifying proof", _pathLogDir, _pathResultDir)

    evalGenProofReport.Run()
    evalVerProofReport.Run()

    err = evalGenProofReport.Dump()
    if err != nil { panic(err) }

    err = evalVerProofReport.Dump()
    if err != nil { panic(err) }
}

func runEvalContract(_pathLogDir string, _pathResultDir string) {
    var err error

    bundleReport := eval.NewEvalContractReportBundle(_pathLogDir, _pathResultDir)
    err = bundleReport.Run()
    if err != nil { panic(err) }

    err = bundleReport.Dump()
    if err != nil { panic(err) }
}

func runShowAccount(_addr string) {
	baseclient = client.NewEthClient(*helper.OptServer, *helper.OptContractAddr, *helper.OptSenderPrivKey, common.HexToAddress(*helper.OptSenderAddr))
    account, err := baseclient.GetAccount(common.HexToAddress(_addr))
    if err != nil { panic(err) }

    fmt.Printf("%+v\n", account)
}

func main() {
	helper.SetupOpt()

	getopt.Parse()

	if *helper.HelpFlag {
		getopt.Usage()
		os.Exit(1)
	}

	args := getopt.Args()

    command := args[0]

	switch command {
	case "testdata":
		runTestdata(args[1], args[2], args[3])
    case "corruption":
		runCorruption(args[1], args[2], args[3])
    case "list-corrupted-files":
		runListCoruptedFiles(args[1])
    case "repair":
		runRepair(args[1])
    case "repair-batch":
		runRepairBatch(args[1])
    case "write-log":
		runWriteLog(args[1])
    case "eval-gentags":
        runEvalGenTag(args[1], args[2])
    case "eval-auditing":
        runEvalAuditing(args[1], args[2])
    case "eval-contract":
        runEvalContract(args[1], args[2])
    case "show-account":
        runShowAccount(args[1])
	default:
		fmt.Printf("Unknown command (command:%s)\n", command)
	}
}
