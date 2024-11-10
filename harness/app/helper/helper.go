package helper

import (
	"time"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

const escape = "\x1b"
const PathDumpDir = "./cache"

const (
	NONE = iota
	RED
	GREEN
	YELLOW
	BLUE
	PURPLE
)

const TimeFormat = "2006/01/02 15:04:05.000"

func GenXZ21Session(_server string, _contractAddr string, _privKey string) (*ethclient.Client, pdp.XZ21Session) {
	cl, err := ethclient.Dial(_server)
	if err != nil { panic(err) }

	contract, err := pdp.NewXZ21(common.HexToAddress(_contractAddr), cl)
	if err != nil { panic(err) }

	privKey, err := crypto.HexToECDSA(_privKey)
	if err != nil { panic(err) }

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(31337))
	if err != nil { panic(err) }

	session := pdp.XZ21Session{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: bind.TransactOpts{
			From: auth.From,
			Signer: auth.Signer,
		},
	}

	return cl, session
}

func Hex(_data []byte) string {
	return "0x" + hex.EncodeToString(_data)
}

func DecodeHex(_s string) ([]byte, error) {
	return hex.DecodeString(_s[2:])
}

func GetCreatorAddr(_prop *pdp.XZ21FileProperty) common.Address {
	return _prop.Creator
}

func PrintLog(format string, args ...interface{}) {
	t := time.Now().Format(TimeFormat)
	m := fmt.Sprintf(format, args...)
	log := fmt.Sprintf("[%s] %s\n", t, m)

	fmt.Printf("%s", log)

	if *OptPathLogFile != "" {
		err := AppendFile(*OptPathLogFile, []byte(log))
		if err != nil { panic(err) }
	}
}

func ParseLog(_log string) (time.Time, string, map[string]string) {
    var datetime time.Time
    var message, tmpDetail string
    var err error

    re1 := regexp.MustCompile(`\[(.*)\] (.*) \((.*)\)`)
    re2 := regexp.MustCompile(`\[(.*)\] (.*)`)

    match := re1.FindStringSubmatch(_log)

    if len(match) == 0 {
        match = re2.FindStringSubmatch(_log)
        datetime, err = time.Parse(TimeFormat, match[1])
        if err != nil { panic(err) }
        message = match[2]
        tmpDetail = ""
    } else {
        datetime, err = time.Parse(TimeFormat, match[1])
        if err != nil { panic(err) }
        message = match[2]
        tmpDetail = match[3]
    }

	detail := make(map[string]string)
	if tmpDetail != "" {
		arr := strings.Split(tmpDetail, ",")
		for _, a1 := range arr {
			a2 := strings.TrimSpace(a1)
			a3 := strings.Split(a2, ":")
			k := a3[0]
			v := a3[1]
			detail[k] = v
		}
	}

    return datetime, message, detail
}

func ParseSize(sizeStr string) (int, int64, error) {
    sizeStr = strings.TrimSpace(sizeStr)
    if len(sizeStr) < 2 {
        return 0, 0, fmt.Errorf("invalid size format")
    }

    numPart := sizeStr[:len(sizeStr)-1]
    unitPart := sizeStr[len(sizeStr)-1]

    num, err := strconv.Atoi(numPart)
    if err != nil {
        return 0, 0, fmt.Errorf("invalid number: %s", numPart)
    }

	var bufSize int64
    switch strings.ToUpper(string(unitPart)) {
    case "K":
		bufSize = int64(1024)
    case "M":
		bufSize = int64(1024 * 1024)
    case "G":
		bufSize = int64(1024 * 1024 * 1024)
    default:
        return 0, 0, fmt.Errorf("unknown unit: %s", string(unitPart))
    }
	return num, bufSize, nil
}

func colorText(_color int, _text string) string {
	return color(_color) + _text + color(NONE)
}

func color(c int) string {
	if c == NONE {
		return fmt.Sprintf("%s[%dm", escape, c)
	}

	return fmt.Sprintf("%s[3%dm", escape, c)
}

func IsEmptyFileProperty(_file *pdp.XZ21FileProperty) bool {
	isEmptySplitNum := (_file.SplitNum == 0)
	isEmptyCreator  := (_file.Creator.Cmp(common.BytesToAddress([]byte{0})) == 0)
	return (isEmptySplitNum && isEmptyCreator)
}

func IsFile(_path string) bool {
    info, err := os.Stat(_path)
    if err != nil {
        return false
    }
    return !info.IsDir()
}

func CalcDigest(_data []byte) [32]byte {
	return sha256.Sum256(_data)
}

func ReadFile(_path string) ([]byte, error) {
	data, err := os.ReadFile(_path)
	return data, err
}

func WriteFile(_path string, _data []byte) {
	os.WriteFile(_path, _data, 0755)
}

func AppendFile(_path string, _data []byte) error {
	// 追記モードでファイルを開く
	file, err := os.OpenFile(_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil { return err }
	defer file.Close()

	// ファイルに書き込み
	_, err = file.Write(_data)

	return err
}

func MakeDumpDirPath(_name string) string {
	return fmt.Sprintf("%s/%s", PathDumpDir, _name)
}

func MakeDumpFilePath(_name string, _filename string) string {
	pathDumpDir := MakeDumpDirPath(_name)
	return fmt.Sprintf("%s/%s", pathDumpDir, _filename)
}

func MakeDumpDir(_name string) (string, error) {
	path := MakeDumpDirPath(_name)
	err := os.MkdirAll(path, 0755)
	return path, err
}

func DumpEntity(_e IfEntity) {
	name := _e.GetName()

	_, err := MakeDumpDir(name)
	if err != nil { panic(err) }

	s, err := _e.ToJson()
	if err != nil { panic(err) }

	pathFile := MakeDumpFilePath(name, "dump.json")
	WriteFile(pathFile, []byte(s))
}

func LoadEntity(_name string, _e IfEntity) error {
	path := MakeDumpFilePath(_name, "dump.json")

	s, err := ReadFile(path)
	if err != nil { return err }

	_e.FromJson(s, *SimFlag)
	_e.AfterLoad()

	return nil
}