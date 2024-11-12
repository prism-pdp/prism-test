package main

import(
    "encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pborman/getopt/v2"
	"strconv"
    "os"

	"github.com/dpduado/dpduado-test/harness/client"
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
