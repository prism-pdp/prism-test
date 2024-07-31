package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"fmt"
	"math/big"
	"os"
	"reflect"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/Nik-U/pbc"
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

var data []byte
var hash [32]byte
var chunkSize uint32

var param pdp.PairingParam

// var keySU1 pdp.PairingKey
// var keySU2 pdp.PairingKey
// var keySU3 pdp.PairingKey
var ramSP  *RamSP
var ramSU1 *RamSU
var ramSU2 *RamSU
var ramSU3 *RamSU

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

	data = []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
		0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49,
		0x50, 0x51, 0x52, 0x53,
	}

	hash = sha256.Sum256(data)

	chunkSize = uint32(10)

	param.Gen()

	ramSP  = GenRamSP()
	ramSU1 = GenRamSU(&param, getAddress(SU1))
	ramSU2 = GenRamSU(&param, getAddress(SU2))
	ramSU3 = GenRamSU(&param, getAddress(SU3))
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
	xz21Para := param.ToXZ21Para()
	_, err := ctxTable[SM].Session.RegisterPara(
		xz21Para.Pairing,
		xz21Para.G,
		xz21Para.U,
	)

	return xz21Para, err
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

func runSetupPhase() {
	var err error

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

func getRamSU(_entity int) *RamSU {
	switch _entity {
	case SU1:
		return ramSU1
	case SU2:
		return ramSU2
	case SU3:
		return ramSU3
	}
	return nil
}

func uploadAlg(_privKey *pbc.Element, _data []byte) (*pdp.Metadata, error) {
	chunk, err := pdp.SplitData(data, chunkSize)
	if err != nil { return nil, err }

	meta := pdp.GenMetadata(&param, _privKey, chunk)

	return meta, nil
}

func runUploadPhase(_entity int) {

	ram := getRamSU(_entity)

	// =================================================
	// SU
	// =================================================
	isFound, err := ctxTable[_entity].Session.SearchFile(hash)
	fmt.Println(isFound)
	if err != nil { panic(err) }
	if isFound {
		fmt.Println(colorText(GREEN, "runUploadPhase_New: dedup"))
	} else {
		fmt.Println(colorText(GREEN, "runUploadPhase_New: new"))
	}

	meta, err := uploadAlg(ram.key.PrivateKey, data)
	if err != nil { panic(err) }

	//-- Send data and meta to SP --//

	// =================================================
	// SP
	// =================================================
	isFound, err = ctxTable[SP].Session.SearchFile(hash)
	if err != nil { panic(err) }
	if isFound {
		//-- Do dedup challenge --//
		ramSP.AppendOwner(_entity, hash, meta)
	} else {
		ramSP.NewFile(_entity, hash, data, meta)
		ctxTable[SP].Session.RegisterFile(hash, ram.addr)
	}
}

func main() {

	setup(os.Args[1], os.Args[2])

	command := os.Args[3]

	switch command {
	case "setup":
		runSetupPhase()
	case "upload":
		fmt.Println("A")
		runUploadPhase(SU1)
		fmt.Println("B")
		runUploadPhase(SU2)
	}
}