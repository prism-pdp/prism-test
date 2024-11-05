package main

import(
    "bufio"
	"fmt"
	"strconv"
    "os"

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

func runParseLog(_path string) {
    file, err := os.Open(_path)
    if err != nil { panic(err) }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        dt, msg, detail := helper.ParseLog(line)
        fmt.Printf("%v -- %s -- %s\n", dt, msg, detail)
    }

    if err := scanner.Err(); err != nil {
        panic(err)
    }
}

func main() {
    args := os.Args

    command := args[1]

	switch command {
	case "testdata":
		runTestdata(args[2], args[3], args[4])
    case "parselog":
        runParseLog(args[2])
	default:
		fmt.Println("Unknown command (command:%s)", command)
	}
}
