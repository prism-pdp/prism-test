package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	pdp "github.com/dpduado/dpduado-go"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
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

type Config struct {
	Server       string `json:'server'`
	PrivKey      string `json:'privKey'`
	ContractAddr string `json:'contractAddr'`
}

func increment(_session *pdp.BaseCounterSession) (*types.Transaction, error) {
	return _session.Increment()
}

func setCount(_session *pdp.BaseCounterSession, _number *big.Int) (*types.Transaction, error) {
	return _session.SetCount(_number)
}

func getNumber(_session *pdp.BaseCounterSession) (*big.Int, error) {
	return _session.Count()
}

func readConf(_path string, _conf *Config) error {
	f, err := os.Open(_path)
	if err != nil { return err }
	defer f.Close()

	err = json.NewDecoder(f).Decode(_conf)
	if err != nil { return err }

	return nil
}

func color(c int) string {
	if c == NONE {
		return fmt.Sprintf("%s[%dm", escape, c)
	}

	return fmt.Sprintf("%s[3%dm", escape, c)
}

func colorText(_color int, _text string) string {
	return color(_color) + _text + color(NONE)
}

func main() {
	var conf Config
	err := readConf(os.Args[1], &conf)
	if err != nil { panic(err) }

	cl, err := ethclient.Dial(conf.Server)
	if err != nil { panic(err) }

	counter, err := pdp.NewBaseCounter(common.HexToAddress(conf.ContractAddr), cl)
	if err != nil { panic(err) }

	keyECDSA, err := crypto.HexToECDSA(conf.PrivKey)
	if err != nil { panic(err) }

	auth, err := bind.NewKeyedTransactorWithChainID(keyECDSA, big.NewInt(31337))
	if err != nil { panic(err) }

	session := pdp.BaseCounterSession{
		Contract: counter,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: bind.TransactOpts{
			From: auth.From,
			Signer: auth.Signer,
		},
	}

	test_data := big.NewInt(10)

	_, err = setCount(&session, test_data)
	if err != nil { panic(err) }

	_, err = increment(&session)
	if err != nil { panic(err) }

	ans, err := getNumber(&session)
	if err != nil { panic(err) }

	expected := new(big.Int).Add(test_data, big.NewInt(1))
	if expected.Cmp(ans) == 0 {
		fmt.Println(colorText(GREEN, "Success"))
	} else {
		fmt.Println(colorText(RED, "Failure"))
	}
}
