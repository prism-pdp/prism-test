package main

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/entity"
	"github.com/dpduado/dpduado-test/harness/session"
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

var data1 []byte
var data2 []byte
var chunkNum uint32

var sm *entity.Manager
var sp *entity.Provider
var tpa *entity.Auditor
var su1 *entity.User
var su2 *entity.User
var su3 *entity.User

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

func setup() {
	data1 = []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39,
		0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49,
		0x50, 0x51, 0x52, 0x53,
	}
	data2 = []byte{
		0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F,
		0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
		0x2A, 0x2B, 0x2C, 0x2D, 0x2E, 0x2F,
		0x3A, 0x3B, 0x3C, 0x3D, 0x3E, 0x3F,
		0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x4F,
		0x5A, 0x5B, 0x5C, 0x5D,
	}
	chunkNum = uint32(5)
}

func runSetupPhase(_server string, _contractAddr string) {
	// =================================================
	// Register param
	// =================================================
	sm.RegisterPara()
	fmt.Println(colorText(GREEN, "registerPara: OK"))

	// =================================================
	// Enroll user accounts
	// =================================================
	sm.EnrollUser(su1.Addr, su1.PublicKeyData.Key)
	fmt.Println(colorText(GREEN, "Enroll SU1: OK"))

	sm.EnrollUser(su2.Addr, su2.PublicKeyData.Key)
	fmt.Println(colorText(GREEN, "Enroll SU2: OK"))

	sm.EnrollUser(su3.Addr, su3.PublicKeyData.Key)
	fmt.Println(colorText(GREEN, "Enroll SU3: OK"))
}

func runUploadPhase(_su *entity.User, _data []byte) {
	// SU checks whether data is uploaded.
	isUploaded := _su.IsUploaded(_data)

	// Processing differs depending on whether the file has already been uploaded or not.
	if isUploaded {
		// SP generates a challenge for deduplication.
		chalData, id := sp.GenDedupChallen(_data, _su.Addr)

		// SP sends the challenge to SU.

		// SU generates a proof to prove ownership of the data to be uploaded.
		proofData := _su.GenDedupProof(&chalData, _data, chunkNum)

		// SP verifies the proof.
		isVerified := sp.VerifyDedupProof(id, &chalData, &proofData)
		if isVerified {
			sp.AppendOwner(_su, _data)
			fmt.Println(colorText(GREEN, "Append: OK"))
		} else {
			fmt.Println(colorText(RED, "Append: NG"))
		}
	} else {
		// SU uploads the file.
		tag := _su.PrepareUpload(_data, chunkNum)

		// SP accepts the file.
		sp.UploadNewFile(_data, &tag, _su.Addr, &_su.PublicKeyData)

		fmt.Println(colorText(GREEN, "New file: OK"))
	}
}

func runUploadChallen(_su *entity.User) {
	// SU gets the list of his/her files.
	fileList := _su.FetchFileList()
	// SU generates challenge and requests to audit each file
	for _, f := range fileList {
		chalData := _su.GenAuditChallen(f)
		result := _su.UploadChallen(f, &chalData)
		if result {
			fmt.Println(colorText(GREEN, "Upload chal: OK"))
		} else {
			fmt.Println(colorText(GREEN, "Upload chal: Skip (Under auditing)"))
		}
	}
}

func runUploadProof() {
	// SP gets challenge from blockchain.
	hashList, chalDataList := sp.DownloadChallen()
	for i, h := range hashList {
		proofData := sp.GenAuditProof(h, &chalDataList[i])
		sp.UploadProof(h, &proofData)
	}
}

func runVerifyAuditProof() {
	// TPA gets challenge and proof from blockchain.
	hashList, chalDataList, proofDataList := tpa.DownloadAuditProof()
	for i, h := range hashList {
		// TPA gets M (list of hash of chunks) from SP.
		f := sp.SearchFile(h)
		chunk, _ := pdp.SplitData(f.Data, f.TagData.Size)
		hashChunks := pdp.HashChunks(chunk)

		// TPA verifies proof.
		result, err := tpa.VerifyAuditProof(&f.TagData, hashChunks, &chalDataList[i], &proofDataList[i], f.Owners[0])
		if err != nil { panic(err) }
		if result {
			fmt.Println(colorText(GREEN, "Verify proof: OK"))
		} else {
			fmt.Println(colorText(RED, "Verify proof: NG"))
		}
	}
}

func runAuditingPhase() {
	runUploadChallen(su1)
	runUploadChallen(su2)

	runUploadProof()

	runVerifyAuditProof()
}

func main() {

	setup()

	server := os.Args[1]
	contractAddr := os.Args[2]
	command := os.Args[3]
	mode := os.Args[4]

	var ledger session.FakeLedger
	if mode == "sim" {
		if command == "setup" {
			ledger = session.GenFakeLedger()
		} else {
			ledger = session.LoadFakeLedger("./cache/fake-ledger.json")
		}
	}

	sessionSM  := session.NewSession(mode, &ledger, getAddress(SM))
	sessionSP  := session.NewSession(mode, &ledger, getAddress(SP))
	sessionTPA := session.NewSession(mode, &ledger, getAddress(TPA))
	sessionSU1 := session.NewSession(mode, &ledger, getAddress(SU1))
	sessionSU2 := session.NewSession(mode, &ledger, getAddress(SU2))
	sessionSU3 := session.NewSession(mode, &ledger, getAddress(SU3))

	if command == "setup" {
		sm  = entity.GenManager(server, contractAddr, getPrivKey(SM), sessionSM)
		sp  = entity.GenProvider(server, contractAddr, getPrivKey(SP), sessionSP)
		tpa = entity.GenAuditor(sessionTPA)
		su1 = entity.GenUser(server, contractAddr, getAddress(SU1), getPrivKey(SU1), &sm.Param, sessionSU1)
		su2 = entity.GenUser(server, contractAddr, getAddress(SU2), getPrivKey(SU2), &sm.Param, sessionSU2)
		su3 = entity.GenUser(server, contractAddr, getAddress(SU3), getPrivKey(SU3), &sm.Param, sessionSU3)
	} else {
		sm = entity.LoadManager("./cache/sm.json", os.Args[1], os.Args[2], getPrivKey(SM), sessionSM)
		sp = entity.LoadProvider("./cache/sp.json", os.Args[1], os.Args[2], getPrivKey(SP), sessionSP)
		tpa = entity.LoadAuditor("./cache/tpa.json", sessionTPA)
		su1 = entity.LoadUser("./cache/su1.json", os.Args[1], os.Args[2], getPrivKey(SU1), sessionSU1)
		su2 = entity.LoadUser("./cache/su2.json", os.Args[1], os.Args[2], getPrivKey(SU2), sessionSU2)
		su3 = entity.LoadUser("./cache/su3.json", os.Args[1], os.Args[2], getPrivKey(SU3), sessionSU3)
	}

	switch command {
	case "setup":
		runSetupPhase(os.Args[1], os.Args[2])
	case "upload":
		runUploadPhase(su1, data1)
		runUploadPhase(su2, data1)
		runUploadPhase(su2, data2)
	case "audit":
		runAuditingPhase()
	}

	sm.Dump("./cache/sm.json")
	sp.Dump("./cache/sp.json")
	tpa.Dump("./cache/tpa.json")
	su1.Dump("./cache/su1.json")
	su2.Dump("./cache/su2.json")
	su3.Dump("./cache/su3.json")
	if mode == "sim" {
		ledger.Dump("./cache/fake-ledger.json")
	}
}