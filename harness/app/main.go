package main

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/entity"
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
var chunkNum uint32

var ledger entity.FakeLedger

var sm entity.Manager
var sp entity.Provider
var su1 entity.User
var su2 entity.User
var su3 entity.User

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

func setup() {
	data = []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
		0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49,
		0x50, 0x51, 0x52, 0x53,
	}

	chunkNum = uint32(5)
}

func runSetupPhase(_server string, _contractAddr string) {
	var err error

	ledger = entity.GenFakeLedger()
	sm  = entity.GenManager(_server, _contractAddr, getPrivKey(SM))
	sp  = entity.GenProvider(_server, _contractAddr, getPrivKey(SP), &ledger)
	su1 = entity.GenUser(_server, _contractAddr, getAddress(SU1), getPrivKey(SU1), &sm.Param, &ledger)
	su2 = entity.GenUser(_server, _contractAddr, getAddress(SU2), getPrivKey(SU2), &sm.Param, &ledger)
	su3 = entity.GenUser(_server, _contractAddr, getAddress(SU3), getPrivKey(SU3), &sm.Param, &ledger)

	// =================================================
	// Register param
	// =================================================
	err = sm.RegisterPara()
	if err != nil { panic(err) }
	fmt.Println(colorText(GREEN, "registerPara: OK"))

	// =================================================
	// Enroll user accounts
	// =================================================
	ledger.RegisterAccount(su1.Addr, &su1.PublicKeyData)
	err = sm.EnrollUser(su1.Addr, su1.PublicKeyData.Key)
	if err != nil { panic(err) }
	fmt.Println(colorText(GREEN, "Enroll SU1: OK"))

	ledger.RegisterAccount(su2.Addr, &su2.PublicKeyData)
	err = sm.EnrollUser(su2.Addr, su2.PublicKeyData.Key)
	if err != nil { panic(err) }
	fmt.Println(colorText(GREEN, "Enroll SU2: OK"))

	ledger.RegisterAccount(su3.Addr, &su3.PublicKeyData)
	err = sm.EnrollUser(su3.Addr, su3.PublicKeyData.Key)
	if err != nil { panic(err) }
	fmt.Println(colorText(GREEN, "Enroll SU3: OK"))
}

func runUploadPhase(_su *entity.User) {
	// SU checks whether data is uploaded.
	isUploaded := _su.IsUploaded(data)

	// Processing differs depending on whether the file has already been uploaded or not.
	if isUploaded {
		// SP generates a challenge for deduplication.
		chalData, id := sp.GenDedupChallen(data, _su.Addr)

		// SP sends the challenge to SU.

		// SU generates a proof to prove ownership of the data to be uploaded.
		proofData := _su.GenDedupProof(&chalData, data, chunkNum)

		// SP verifies the proof.
		isVerified := sp.VerifyDedupProof(id, &chalData, &proofData)
		if isVerified {
			sp.AppendOwner(_su, data)
			fmt.Println(colorText(GREEN, "Append: OK"))
		} else {
			fmt.Println(colorText(RED, "Append: NG"))
		}
	} else {
		// SU uploads the file.
		tag := _su.PrepareUpload(data, chunkNum)

		// SP accepts the file.
		sp.UploadNewFile(data, &tag, _su.Addr, &_su.PublicKeyData)

		fmt.Println(colorText(GREEN, "New file: OK"))
	}
}

// func runAuditingPhase(_su *entity.User) {
// 	// SU generates challenge and requests to audit the file
// 	// TODO: Search for his/her own files and generate a challenge by specifying the ID of the file.
// 	chalData := _su.GenAuditChallen(data)

// 	//
// }

func main() {

	setup()

	command := os.Args[3]

	switch command {
	case "setup":
		runSetupPhase(os.Args[1], os.Args[2])
		ledger.Dump("./cache/setup-ledger.json")
		sp.Dump("./cache/setup-sp.json")
		su1.Dump("./cache/setup-su1.json")
		su2.Dump("./cache/setup-su2.json")
		su3.Dump("./cache/setup-su3.json")
	case "upload":
		ledger = entity.LoadFakeLedger("./cache/setup-ledger.json")
		sp = entity.LoadProvider("./cache/setup-sp.json", os.Args[1], os.Args[2], getPrivKey(SP), &ledger)
		su1 = entity.LoadUser("./cache/setup-su1.json", os.Args[1], os.Args[2], getPrivKey(SU1), &ledger)
		su2 = entity.LoadUser("./cache/setup-su2.json", os.Args[1], os.Args[2], getPrivKey(SU2), &ledger)
		su3 = entity.LoadUser("./cache/setup-su3.json", os.Args[1], os.Args[2], getPrivKey(SU3), &ledger)

		runUploadPhase(&su1)
		runUploadPhase(&su2)

		ledger.Dump("./cache/upload-ledger.json")
		sp.Dump("./cache/upload-sp.json")
	}
}