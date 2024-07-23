package main

import (
	// "encoding/json"
	"fmt"
	"math/big"
	"os"
	// "io/ioutil"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	// "github.com/ethereum/go-ethereum/core/types"
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

const (
	SM = 0
	SP
	TPA
	SU1
	SU2
	SU3
)

type Account struct {
	Address string `json:'Address'`
	PrivKey string `json:'PrivKey'`
}

func getAddress(_entity int) string {
	tmp := fmt.Sprintf("ADDRESS_%d", _entity)
	return os.Getenv(tmp)
}

func getPrivKey(_entity int) string {
	tmp := fmt.Sprintf("PRIVKEY_%d", _entity)
	return os.Getenv(tmp)
}

/*
func increment(_session *pdp.BaseCounterSession) (*types.Transaction, error) {
	return _session.Increment()
}

func setCount(_session *pdp.BaseCounterSession, _number *big.Int) (*types.Transaction, error) {
	return _session.SetCount(_number)
}

func getNumber(_session *pdp.BaseCounterSession) (*big.Int, error) {
	return _session.Count()
}
*/


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
	server := os.Args[1]
	contractAddr := os.Args[2]

	cl, err := ethclient.Dial(server)
	if err != nil { panic(err) }

	contract, err := pdp.NewXZ21(common.HexToAddress(contractAddr), cl)
	if err != nil { panic(err) }

	keyECDSA, err := crypto.HexToECDSA(getPrivKey(SM))
	if err != nil { panic(err) }

	auth, err := bind.NewKeyedTransactorWithChainID(keyECDSA, big.NewInt(31337))
	if err != nil { panic(err) }

	//session := pdp.XZ21Session{
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

	addrSM, err := session.GetAddrSM()
	if err != nil { panic(err) }

	fmt.Println(addrSM)

	/*
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
	*/
}
