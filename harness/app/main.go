package main

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"reflect"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	SM = iota
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
}

func checkSuperAccount() error {
	ctx := ctxTable[SM]

	addrSM1 := getAddress(SM)
	addrSM2, err := ctx.Session.AddrSM()
	if err != nil { return err }
	if addrSM1 != addrSM2 {
		return fmt.Errorf("Invalid SM address (%s, %s)", addrSM1.Hex(), addrSM2.Hex())
	}

	addrSP1 := getAddress(SP)
	addrSP2, err := ctx.Session.AddrSP()
	if err != nil { panic(err) }
	if addrSP1 != addrSP2 {
		return fmt.Errorf("Invalid SP address (%s, %s)", addrSP1.Hex(), addrSP2.Hex())
	}

	addrTPA1 := getAddress(TPA)
	addrTPA2, err := ctx.Session.AddrTPA()
	if err != nil { panic(err) }
	if addrTPA1 != addrTPA2 {
		return fmt.Errorf("Invalid TPA address (%s, %s)", addrTPA1.Hex(), addrTPA2.Hex())
	}
	
	return nil
}

func registerPara() (*pdp.XZ21Para, error) {
	var tmp pdp.PairingParam
	tmp.Gen()

	para := tmp.ToXZ21Para()
	_, err := ctxTable[SM].Session.RegisterPara(
		para.Pairing,
		para.G,
		para.U,
	)

	return para, err
}

func checkPara(_para *pdp.XZ21Para) error {
	para, err := ctxTable[SM].Session.GetPara()
	if err != nil { return err }

	if para.Pairing != _para.Pairing {
		return fmt.Errorf("Invalid pairing")
	}
	if !reflect.DeepEqual(para.U, _para.U) {
		return fmt.Errorf("Invalid U")
	}
	if !reflect.DeepEqual(para.G, _para.G) {
		return fmt.Errorf("Invalid G")
	}

	return nil
}

func enrollUserAccount() error {
	var err error

	ctx := ctxTable[SM]

	addrSU1 := getAddress(SU1)
	_, err = ctx.Session.EnrollAccount(addrSU1, "PUBKEY_1")
	if err != nil { return err }

	addrSU2 := getAddress(SU2)
	_, err = ctx.Session.EnrollAccount(addrSU2, "PUBKEY_2")
	if err != nil { return err }

	addrSU3 := getAddress(SU3)
	_, err = ctx.Session.EnrollAccount(addrSU3, "PUBKEY_3")
	if err != nil { return err }

	return nil
}

func checkUserAccount() error {
	var err error

	addrSU1 := getAddress(SU1)
	accountSU1, err := ctxTable[SU1].Session.GetAccount(addrSU1)
	if err != nil { return err }
	if len(accountSU1.PubKey) == 0 { return fmt.Errorf("SU1 is missing.") }

	addrSU2 := getAddress(SU2)
	accountSU2, err := ctxTable[SU2].Session.GetAccount(addrSU2)
	if err != nil { return err }
	if len(accountSU2.PubKey) == 0 { return fmt.Errorf("SU2 is missing.") }

	addrSU3 := getAddress(SU3)
	accountSU3, err := ctxTable[SU3].Session.GetAccount(addrSU3)
	if err != nil { return err }
	if len(accountSU3.PubKey) == 0 { return fmt.Errorf("SU3 is missing.") }

	return nil
}

func main() {
	var err error

	setup(os.Args[1], os.Args[2])

	// =================================================
	// Check addresses
	// =================================================
	err = checkSuperAccount()
	if err != nil { panic(err) }
	fmt.Println(colorText(GREEN, "checkSuperAccount: ok"))

	// =================================================
	// Register param
	// =================================================
	para, err := registerPara()
	if err != nil { panic(err) }
	fmt.Println(colorText(GREEN, "registerPara: ok"))

	// =================================================
	// Check param
	// =================================================
	err = checkPara(para)
	if err != nil { panic(err) }
	fmt.Println(colorText(GREEN, "checkPara: ok"))

	// =================================================
	// Enroll user accounts
	// =================================================
	err = enrollUserAccount()
	if err != nil { panic(err) }
	fmt.Println(colorText(GREEN, "enrollUserAccount: ok"))

	// =================================================
	// Check user accounts
	// =================================================
	err = checkUserAccount()
	if err != nil { panic(err) }
	fmt.Println(colorText(GREEN, "checkUserAccount: ok"))

	// =================================================
	// Upload phase (New file)
	// =================================================
}
