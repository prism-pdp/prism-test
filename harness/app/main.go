package main

import (
	// "encoding/json"
	"fmt"
	"math/big"
	"os"
	// "io/ioutil"
	"crypto/ecdsa"

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

func getAddress(_entity int) common.Address {
	tmp := fmt.Sprintf("ADDRESS_%d", _entity)
	return common.HexToAddress(os.Getenv(tmp))
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

type Context struct {
	PrivKey *ecdsa.PrivateKey
	Auth *bind.TransactOpts
	Session pdp.XZ21Session
}

var ctxTable [6]*Context

func setupContext(_entity int, _contract *pdp.XZ21) {
	var err error

	var ctx Context

	ctx.PrivKey, err = crypto.HexToECDSA(getPrivKey(_entity))
	if err != nil { panic(err) }

	ctx.Auth, err = bind.NewKeyedTransactorWithChainID(ctx.PrivKey, big.NewInt(31337))
	if err != nil { panic(err) }

	ctx.Session = pdp.XZ21Session{
		Contract: _contract,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: bind.TransactOpts{
			From: ctx.Auth.From,
			Signer: ctx.Auth.Signer,
		},
	}

	ctxTable[_entity] = &ctx
}

func setup(_server string, _contractAddr string) {
	cl, err := ethclient.Dial(_server)
	if err != nil { panic(err) }

	contract, err := pdp.NewXZ21(common.HexToAddress(_contractAddr), cl)
	if err != nil { panic(err) }

	setupContext(SM, contract)
	setupContext(SP, contract)
	setupContext(TPA, contract)
	setupContext(SU1, contract)
	setupContext(SU2, contract)
	setupContext(SU3, contract)

	fmt.Println(ctxTable)
}

func main() {
	setup(os.Args[1], os.Args[2])

	// =================================================
	// Check addresses
	// =================================================
	ctxSM := ctxTable[SM]
	fmt.Println(ctxTable)
	addrSM, err := ctxSM.Session.AddrSM()
	if err != nil { panic(err) }
	fmt.Println("Address of SM: " + addrSM.Hex())

	addrSP, err := ctxSM.Session.AddrSP()
	if err != nil { panic(err) }
	fmt.Println("Address of SP: " + addrSP.Hex())

	addrTPA, err := ctxSM.Session.AddrTPA()
	if err != nil { panic(err) }
	fmt.Println("Address of TPA: " + addrTPA.Hex())

	// =================================================
	// Enroll service users
	// =================================================

	addrSU1 := getAddress(SU1)
	_, err = ctxSM.Session.EnrollAccount(addrSU1, "PUBKEY_1")
	if err != nil { panic(err) }

	addrSU2 := getAddress(SU2)
	_, err = ctxSM.Session.EnrollAccount(addrSU2, "PUBKEY_2")
	if err != nil { panic(err) }

	addrSU3 := getAddress(SU3)
	_, err = ctxSM.Session.EnrollAccount(addrSU3, "PUBKEY_3")
	if err != nil { panic(err) }

	// =================================================
	// Upload phase (New file)
	// =================================================

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
