package helper

import (
	"time"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"os"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

const escape = "\x1b"

const (
	NONE = iota
	RED
	GREEN
	YELLOW
	BLUE
	PURPLE
)

func GenXZ21Session(_server string, _contractAddr string, _privKey string) pdp.XZ21Session {
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

	return session
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

func PrintLog(_log string) {
	t := time.Now().Format(time.DateTime)
	fmt.Printf("[%s] %s\n", t, _log)
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