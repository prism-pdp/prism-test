package main

import(
    "encoding/json"
	"fmt"
	"strconv"
    "os"

	"github.com/dpduado/dpduado-test/harness/eval"
	"github.com/dpduado/dpduado-test/harness/helper"
)


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

func runEvalGenTag(_pathLogDir string) {
    evalReport := eval.NewEvalReport()
    evalReport.SetupReport("GenTags", "gentag-", "generate tags")

    evalReport.ProcTimeReport["GenTags"].Run(_pathLogDir)

    tmp, _ := json.MarshalIndent(evalReport, "", "\t")
    fmt.Println(string(tmp))
}

func main() {
    args := os.Args

    command := args[1]

	switch command {
	case "testdata":
		runTestdata(args[2], args[3], args[4])
    case "inflate":
        runInflateTestdata(args[2], args[3], args[4])
    case "eval-gentag":
        runEvalGenTag(args[2])
	default:
		fmt.Println("Unknown command (command:%s)", command)
	}
}
