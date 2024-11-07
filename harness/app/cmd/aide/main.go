package main

import(
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
    num, bufSize, err := helper.ParseSize(_size)
    if err != nil { panic(err) }

    val, err := strconv.Atoi(_val)
    if err != nil { panic(err) }

    buf := make([]byte, bufSize)
    for i := range bufSize {
        buf[i] = byte(val)
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

func runInflateTestdata(_pathIn string, _pathOut string, _scale string) {
    data, err := helper.ReadFile(_pathIn)
    if err != nil { panic(err) }

    scale, err := strconv.Atoi(_scale)
    if err != nil { panic(err) }

    for i := 0; i < scale; i++ {
        helper.AppendFile(_pathOut, data)
    }
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
    case "inflate":
        runInflateTestdata(args[1], args[2], args[3])
    case "eval-gentags":
        runEvalGenTag(args[1], args[2])
    case "eval-auditing":
        runEvalAuditing(args[1], args[2])
    case "show-account":
        runShowAccount(args[1])
	default:
		fmt.Printf("Unknown command (command:%s)\n", command)
	}
}
