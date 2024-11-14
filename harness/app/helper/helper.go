package helper

import (
	"bufio"
	"bytes"
    "encoding/binary"
	"time"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/exp/slices"
	"math"
	"math/big"
	"math/rand"
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

func ParseSize(_sizeStr string) (int, int64, error) {
    sizeStr := strings.TrimSpace(_sizeStr)
    if len(sizeStr) < 2 {
        return 0, 0, fmt.Errorf("invalid size format")
    }

    numPart := sizeStr[:len(sizeStr)-1]
    unitPart := sizeStr[len(sizeStr)-1]

    num, err := strconv.Atoi(numPart)
    if err != nil {
        return 0, 0, fmt.Errorf("invalid number: %s", numPart)
    }

	var unitSize int64
    switch strings.ToUpper(string(unitPart)) {
    case "K":
		unitSize = int64(1024)
    case "M":
		unitSize = int64(1024 * 1024)
    case "G":
		unitSize = int64(1024 * 1024 * 1024)
    default:
        return 0, 0, fmt.Errorf("unknown unit: [%s]", string(unitPart))
    }
	return num, unitSize, nil
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

func ReadLines(_path string) ([]string, error) {
	var lines []string

	f, err := os.Open(_path)
	if err != nil { return lines, err }
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func WriteFile(_path string, _data []byte) {
	os.WriteFile(_path, _data, 0755)
}

func WriteFileUint16(_path string, _num int, _unitSize int64, _val uint16) error {
    buf := make([]byte, _unitSize)
    for i:= 0; i < len(buf); i += 2 {
        binary.BigEndian.PutUint16(buf[i:i+2], _val)
    }

    f, err := os.Create(_path)
    if err != nil { return err }

    for range _num {
        _, err := f.Write(buf)
		if err != nil { return err }
    }

    err = f.Close()
	if err != nil { return err }

	return nil
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

func Uniq(_array []string) []string {
	slices.Sort(_array)
	unique := slices.Compact(_array)
	return unique
}

func Remove(_array []string, _value string) []string {
	for i, v := range _array {
		if v == _value {
			return append(_array[:i], _array[i+1:]...)
		}
	}
	return _array
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

func CalcMean(_array []int64) float64 {
	var sum int64 = 0
	var count int64 = int64(len(_array))
	for _, v := range _array {
		sum += v
	}
	return float64(sum) / float64(count)
}

func CalcStandardDeviation(_array []int64, _mean float64) float64 {
	var variance float64
	for _, v := range _array {
		variance += math.Pow(float64(v) - _mean, 2)
	}
	variance /= float64(len(_array))
	return math.Sqrt(variance)
}

func ToggleBit(_data []byte, _pos uint32) (uint32, uint32) {
	index  := uint32(_pos / 8)
	offset := uint32(_pos % 8)

	_data[index] ^= (1 << offset)

	return index, offset
}

func MostFrequentValue(_data []byte, _limit int) uint16 {
    lut := make(map[uint16]int)

    // checks up the first 32 bytes
    for i := 0; i < _limit; i += 2 {
        var num uint16
        err := binary.Read(bytes.NewReader(_data[i:i+2]), binary.BigEndian, &num)
        if err != nil { panic(err) }

        if _, ok := lut[num]; !ok {
            lut[num] = 0
        }
        lut[num] += 1
    }

	var maxKey uint16
	var maxValue int = 0
	for k, v := range lut {
		if maxValue < v {
			maxKey = k
		}
	}

	return maxKey
}

func DrawLots(_probability float64) bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Float64() < _probability
}

func ToFloat64(_val string) float64 {
	val, err := strconv.ParseFloat(_val, 64)
	if err != nil { panic(err) }

	return val
}