package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/entity"
	"github.com/dpduado/dpduado-test/harness/helper"
	"github.com/dpduado/dpduado-test/harness/types"
)

var mode string
var command string
var server string
var contractAddr string

var data1 []byte
var data2 []byte
var chunkNum uint32

var sm *entity.Manager
var sp *entity.Provider
var tpa *entity.Auditor
var su1 *entity.User
var su2 *entity.User
var su3 *entity.User

var clientTable map[types.EntityType]client.BaseClient

var ledger client.FakeLedger

type Account struct {
	Address string `json:'Address'`
	PrivKey string `json:'PrivKey'`
}

func getAddress(_mode string, _entity types.EntityType) common.Address {
	var addr common.Address

	switch _mode {
	case "eth":
		tmp := fmt.Sprintf("ADDRESS_%d", _entity)
		addr = common.HexToAddress(os.Getenv(tmp))
	case "sim":
		tmp := fmt.Sprintf("0x100%d", int(_entity))
		addr = common.HexToAddress(tmp)
	}

	return addr
}

func getPrivKey(_mode string, _entity types.EntityType) string {
	var key string

	switch mode {
	case "eth":
		tmp := fmt.Sprintf("PRIVKEY_%d", _entity)
		key = os.Getenv(tmp)
	default:
		key = "NA"
	}

	return key
}

// func getName(_entity EntityType) string {
// 	switch _entity {
// 	case SM:
// 		return "SM"
// 	case SP:
// 		return "SP"
// 	case TPA:
// 		return "TPA"
// 	case SU1:
// 		return "SU1"
// 	case SU2:
// 		return "SU2"
// 	case SU3:
// 		return "SU3"
// 	}
// 	return "NA"
// }

func setup(_opts []string) {
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

	mode = _opts[1]
	command = _opts[2]

	if mode == "eth" {
		server = _opts[3]
		contractAddr = _opts[4]
	} else {
		server = ""
		contractAddr = ""
	}

	clientOpts := client.NewClientOpts(mode, server, contractAddr)

	for _, e := range types.EntityList {
		clientOpts.AddrTable[e] = getAddress(mode, e)
		clientOpts.PrivKeyTable[e] = getPrivKey(mode, e)
	}

	if mode == "sim" {
		// make fake ledger
		if command == "setup" {
			ledger = client.GenFakeLedger()
		} else {
			ledger = client.LoadFakeLedger("./cache/fake-ledger.json")
		}
		clientOpts.Ledger = &ledger
	}

	clientTable = make(map[types.EntityType]client.BaseClient)
	for _, e := range types.EntityList {
		clientTable[e] = client.NewClient(mode, e, &clientOpts)
	}

	sm = entity.MakeManager("./cache/sm.json", clientTable[types.SM], "SM")
	sp = entity.MakeProvider("./cache/sp.json", clientTable[types.SP], "SP")
	tpa = entity.MakeAuditor("./cache/tpa.json", clientTable[types.TPA], "TPA")
	su1 = entity.MakeUser("./cache/su1.json", clientTable[types.SU1], &sm.Param, "SU1")
	su2 = entity.MakeUser("./cache/su2.json", clientTable[types.SU2], &sm.Param, "SU2")
	su3 = entity.MakeUser("./cache/su3.json", clientTable[types.SU3], &sm.Param, "SU3")
}

func runSetupPhase() {
	helper.PrintLog("Start Setup Phase")

	// =================================================
	// Register param
	// =================================================
	sm.RegisterParam()
	helper.PrintLog("Register Parameter: OK")

	// =================================================
	// Enroll user accounts
	// =================================================
	sm.EnrollUser(su1.Addr, su1.PublicKeyData.Key)
	helper.PrintLog("Enroll Service User 1: OK")

	sm.EnrollUser(su2.Addr, su2.PublicKeyData.Key)
	helper.PrintLog("Enroll Service User 2: OK")

	sm.EnrollUser(su3.Addr, su3.PublicKeyData.Key)
	helper.PrintLog("Enroll Service User 3: OK")

	helper.PrintLog("Finish Setup Phase")
}

func runUploadPhase(_su *entity.User, _data []byte) {
	helper.PrintLog("Start Upload Phase")

	// SU checks whether data is uploaded.
	isUploaded := _su.IsUploaded(_data)

	// Processing differs depending on whether the file has already been uploaded or not.
	if isUploaded {
		// SP generates a challenge for deduplication.
		chalData := sp.GenDedupChal(_data, _su.Addr)

		// SP sends the challenge to SU.

		// SU generates a proof to prove ownership of the data to be uploaded.
		proofData := _su.GenDedupProof(&chalData, _data, chunkNum)

		// SP verifies the proof.
		isRegistered := sp.RegisterOwnerToFile(_su, _data, &chalData, &proofData)
		if isRegistered {
			helper.PrintLog("Append Owner: OK")
		} else {
			helper.PrintLog("Append Owner: NG")
		}
	} else {
		// SU uploads the file.
		tag := _su.PrepareUpload(_data, chunkNum)

		// SP accepts the file.
		err := sp.UploadNewFile(_data, &tag, _su.Addr, &_su.PublicKeyData)
		if err != nil { panic(err) }

		helper.PrintLog("Upload New file: OK")
	}

	helper.PrintLog("Finish Upload Phase")
}

func runUploadAuditingChal(_su *entity.User) {
	helper.PrintLog(fmt.Sprintf("Start upload auditing chal (entity:%s)", _su.Name))

	// SU gets the list of his/her files.
	fileList := _su.GetFileList()
	// SU generates challenge and requests to audit each file
	for i, f := range fileList {
		helper.PrintLog(fmt.Sprintf("Upload auditing chal (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(fileList)))
		chalData := _su.GenAuditingChal(f)
		_su.UploadAuditingChal(f, &chalData)
	}

	helper.PrintLog(fmt.Sprintf("Finish upload auditing chal (entity:%s)", _su.Name))
}

func runUploadAuditingProof() {
	helper.PrintLog(fmt.Sprintf("Start upload auditing proof (entity:%s)", sp.Name))

	// SP gets challenge from blockchain.
	fileList, chalDataList := sp.DownloadAuditingChal()
	for i, h := range fileList {
		helper.PrintLog(fmt.Sprintf("Download auditing chal (file:%s, index:%d/%d)", helper.Hex(h[:]), i+1, len(fileList)))
	}

	// For test
	if len(fileList) != 2 { panic(fmt.Errorf("Invalid fileList size (expect:2, actual:%d)", len(fileList))) }

	for i, f := range fileList {
		helper.PrintLog(fmt.Sprintf("Upload auditing proof (entity:%s, file:%s, index:%d/%d)", sp.Name, helper.Hex(f[:]), i+1, len(fileList)))
		proofData := sp.GenAuditingProof(f, &chalDataList[i])
		sp.UploadAuditingProof(f, &proofData)
	}

	helper.PrintLog(fmt.Sprintf("Finish upload auditing proof (entity:%s)", sp.Name))
}

func runVerifyAuditingProof() {
	helper.PrintLog(fmt.Sprintf("Start verify auditing proof (entity:%s)", tpa.Name))

	// TPA gets challenge and proof from blockchain.
	fileList, reqList := tpa.GetAuditingReqList()
	for i, f := range fileList {
		helper.PrintLog(fmt.Sprintf("Download auditing req (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(fileList)))
	}

	if len(fileList) != 2 { panic(fmt.Errorf("Invalid fileList size (expect:2, actual:%d)", len(fileList))) }

	for i, f := range fileList {
		// TPA gets M (list of hash of chunks) from SP.
		file := sp.SearchFile(f)
		chunk, _ := pdp.SplitData(file.Data, file.TagData.Size)

		ここ
		chal := reqList[i].ChalData.Import(&params)
		hashChunks := pdp.HashChunks(chunk, &reqList[i].ChalData)

		// TPA verifies proof.
		result, err := tpa.VerifyAuditingProof(&file.TagData, hashChunks, &reqList[i].ChalData, &reqList[i].ProofData, file.Owners[0])
		if err != nil { panic(err) }

		helper.PrintLog(fmt.Sprintf("Upload auditing result (file:%s, result:%t)", helper.Hex(f[:]), result))
		tpa.UploadAuditingResult(f, result)
	}

	helper.PrintLog(fmt.Sprintf("Finish verify auditing proof (entity:%s)", tpa.Name))
}

func runAuditingPhase() {
	helper.PrintLog("Start Auditing Phase")

	// 1st
	runUploadAuditingChal(su1)
	runUploadAuditingChal(su2)
	runUploadAuditingProof()
	runVerifyAuditingProof()

	time.Sleep(3 * time.Second) // TODO: Implement WaitMined into all contracts.

	// 2nd
	runUploadAuditingChal(su1)
	runUploadAuditingChal(su2)
	runUploadAuditingProof()
	runVerifyAuditingProof()

	helper.PrintLog("Finish Auditing Phase")
}

func main() {

	setup(os.Args)

	switch command {
	case "setup":
		runSetupPhase()
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