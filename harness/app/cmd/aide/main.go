package main

import(
	"fmt"
	"strconv"
	"strings"
    "os"
)

func ParseSize(sizeStr string) (int64, error) {
    sizeStr = strings.TrimSpace(sizeStr) // 空白をトリム
    if len(sizeStr) < 2 {
        return 0, fmt.Errorf("invalid size format")
    }

    // 数値部分を取得
    numPart := sizeStr[:len(sizeStr)-1]
    unitPart := sizeStr[len(sizeStr)-1]

    // 数値をパース
    num, err := strconv.ParseFloat(numPart, 64)
    if err != nil {
        return 0, fmt.Errorf("invalid number: %s", numPart)
    }

    // 単位に応じてバイト数を計算
    switch strings.ToUpper(string(unitPart)) {
    case "K":
        return int64(num * 1024), nil
    case "M":
        return int64(num * 1024 * 1024), nil
    case "G":
        return int64(num * 1024 * 1024 * 1024), nil
    case "B":
        return int64(num), nil
    default:
        return 0, fmt.Errorf("unknown unit: %s", string(unitPart))
    }
}

func runTestdata(_path string, _unit string, _num string) {
    unit, err := ParseSize(_unit)
    if err != nil { panic(err) }

    buf := make([]byte, unit)
    for i := range unit {
        buf[i] = byte(i)
    }

    num, err := strconv.Atoi(_num)
    if err != nil { panic(err) }

    f, err := os.Create(_path)
    if err != nil { panic(err) }

    for range num {
        _, err := f.Write(buf)
        if err != nil { panic(err) }
    }

    err = f.Close()
    if err != nil { panic(err) }
}

func main() {
    args := os.Args

    command := args[1]

	switch command {
	case "testdata":
		runTestdata(args[2], args[3], args[4])
	default:
		fmt.Println("Unknown command (command:%s)", command)
	}
}
