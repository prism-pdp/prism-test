package main

import (
	"fmt"
	"os"

	"github.com/pborman/getopt/v2"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/entity"
	"github.com/dpduado/dpduado-test/harness/helper"
)

var (
	helpFlag *bool
	simFlag *bool
	optServer *string
	optContractAddr *string
	optSenderAddr *string
	optSenderPrivKey *string
)

var (
	server string
	contractAddr string
	senderAddr string
	senderPrivKey string
)

var command string

var data1 []byte
var data2 []byte
var chunkNum uint32

var ledger client.FakeLedger

type Account struct {
	Address string `json:'Address'`
	PrivKey string `json:'PrivKey'`
}

func toString(_opt *string) string {
	if _opt == nil {
		return ""
	}
	return *_opt
}

func makePath(_name string) string {
	return "./cache/" + _name + ".json"
}

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

	command = _opts[0]

	// server := toString(optServer)
	// contractAddr:= toString(optContractAddr)
	senderAddr = toString(optSenderAddr)
	senderPrivKey = toString(optSenderPrivKey)

	if *simFlag {
		// make fake ledger
		if command == "setup" {
			ledger = client.GenFakeLedger()
		} else {
			ledger = client.LoadFakeLedger("./cache/fake-ledger.json")
		}
	}
}

func runSetupPhase(_smAddr string, _smPrivKey string, _spAddr string, _spPrivKey string) {
	helper.PrintLog("Start Setup Phase")

	sm := entity.GenManager("SM", _smAddr, _smPrivKey)
	if *simFlag {
		sm.SetupSimClient(&ledger)
	}

	// =================================================
	// Register param
	// =================================================
	sm.RegisterParam()
	helper.PrintLog("Register Parameter: OK")

	sp := entity.GenProvider("SP", _spAddr, _spPrivKey)
	if *simFlag {
		sp.SetupSimClient(&ledger)
	}

	sm.Dump("./cache/sm.json")
	sp.Dump("./cache/sp.json")

	helper.PrintLog("Finish Setup Phase")
}

func runEnrollUser(_name string, _addr string, _privKey string) {
	helper.PrintLog(fmt.Sprintf("enroll service user (name:%s)", _name))

	sm := entity.LoadManager("./cache/sm.json")
	if *simFlag {
		sm.SetupSimClient(&ledger)
	}

	su := entity.GenUser(_addr, _privKey, sm.GetParam(), _name)
	sm.EnrollUser(su)

	path := makePath(_name)
	su.Dump(path)
}

func runEnrollAuditor(_name string, _addr string) {
	helper.PrintLog("enroll auditor (name:%s)", _name)

	sm := entity.LoadManager("./cache/sm.json")
	if *simFlag {
		sm.SetupSimClient(&ledger)
	}

	tpa := entity.GenAuditor(_name, _addr)
	sm.EnrollAuditor(tpa)

	path := makePath(_name)
	tpa.Dump(path)
}

func runUploadPhase(_name string, _data []byte) {
	var path string

	helper.PrintLog("Start Upload Phase")

	path = makePath("sp")
	sp := entity.LoadProvider(path)
	if *simFlag {
		sp.SetupSimClient(&ledger)
	}

	path = makePath(_name)
	su := entity.LoadUser(path)
	if *simFlag {
		su.SetupSimClient(&ledger)
	}

	// SU checks whether data is uploaded.
	isUploaded := su.IsUploaded(_data)

	// Processing differs depending on whether the file has already been uploaded or not.
	if isUploaded {
		// SP generates a challenge for deduplication.
		chalData := sp.GenDedupChal(_data, su.Addr)

		// SP sends the challenge to SU.

		// SU generates a proof to prove ownership of the data to be uploaded.
		proofData := su.GenDedupProof(&chalData, _data, chunkNum)

		// SP verifies the proof.
		isRegistered := sp.RegisterOwnerToFile(su, _data, &chalData, &proofData)
		if isRegistered {
			helper.PrintLog("Append Owner: OK")
		} else {
			helper.PrintLog("Append Owner: NG")
		}
	} else {
		// SU uploads the file.
		tag := su.PrepareUpload(_data, chunkNum)

		// SP accepts the file.
		err := sp.UploadNewFile(_data, &tag, su.Addr, &su.PublicKeyData)
		if err != nil { panic(err) }

		helper.PrintLog("Upload New file: OK")
	}

	path = makePath("sp")
	sp.Dump(path)

	path = makePath(_name)
	su.Dump(path)

	helper.PrintLog("Finish Upload Phase")
}

func runUploadAuditingChal(_name string) {
	path := makePath(_name)
	su := entity.LoadUser(path)
	if *simFlag {
		su.SetupSimClient(&ledger)
	}

	helper.PrintLog("Start upload auditing chal (entity:%s)", su.Name)

	// SU gets the list of his/her files.
	fileList := su.GetFileList()
	// SU generates challenge and requests to audit each file
	for i, f := range fileList {
		helper.PrintLog("Upload auditing chal (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(fileList))
		chalData := su.GenAuditingChal(f)
		su.UploadAuditingChal(f, &chalData)
	}

	path = makePath(_name)
	su.Dump(path)

	helper.PrintLog("Finish upload auditing chal (entity:%s)", su.Name)
}

func runUploadAuditingProof() {
	path := makePath("sp")
	sp := entity.LoadProvider(path)
	if *simFlag {
		sp.SetupSimClient(&ledger)
	}

	helper.PrintLog("Start upload auditing proof (entity:%s)", sp.Name)

	// SP gets challenge from blockchain.
	fileList, chalDataList := sp.DownloadAuditingChal()
	for i, h := range fileList {
		helper.PrintLog("Download auditing chal (file:%s, index:%d/%d)", helper.Hex(h[:]), i+1, len(fileList))
	}

	for i, f := range fileList {
		helper.PrintLog("Upload auditing proof (entity:%s, file:%s, index:%d/%d)", sp.Name, helper.Hex(f[:]), i+1, len(fileList))
		proofData := sp.GenAuditingProof(f, &chalDataList[i])
		sp.UploadAuditingProof(f, &proofData)
	}

	sp.Dump(path)

	helper.PrintLog("Finish upload auditing proof (entity:%s)", sp.Name)
}

func runVerifyAuditingProof(_name string) {
	pathSP := makePath("sp")
	sp := entity.LoadProvider(pathSP)
	if *simFlag {
		sp.SetupSimClient(&ledger)
	}

	pathTPA := makePath(_name)
	tpa := entity.LoadAuditor(pathTPA)
	if *simFlag {
		tpa.SetupSimClient(&ledger)
	}

	helper.PrintLog("Start verify auditing proof (entity:%s)", tpa.Name)

	// TPA gets challenge and proof from blockchain.
	fileList, reqDataList := tpa.GetAuditingReqList()
	for i, f := range fileList {
		helper.PrintLog("Download auditing req (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(fileList))
	}

	for i, f := range fileList {
		// TPA gets M (list of hash of chunks) from SP.
		owner, digestSet, tagDataSet := sp.PrepareVerificationData(f, &reqDataList[i].ChalData)

		// TPA verifies proof.
		result, err := tpa.VerifyAuditingProof(tagDataSet, digestSet, &reqDataList[i], owner)
		if err != nil { panic(err) }

		helper.PrintLog("Upload auditing result (file:%s, result:%t)", helper.Hex(f[:]), result)
		tpa.UploadAuditingResult(f, result)
	}

	helper.PrintLog("Finish verify auditing proof (entity:%s)", tpa.Name)

	sp.Dump(pathSP)
	tpa.Dump(pathTPA)
}

func main() {

	helpFlag = getopt.BoolLong("help", 'h', "display help")
	simFlag  = getopt.BoolLong("sim", 0, "simulation mode (disable blockchain)")
	optServer = getopt.StringLong("server", 0, "", "server's URL")
	optContractAddr = getopt.StringLong("contract", 0, "", "contract address")
	optSenderAddr = getopt.StringLong("sender-addr", 0, "", "sender's address")
	optSenderPrivKey = getopt.StringLong("sender-key", 0, "", "sender's private key")

	getopt.Parse()

	if *helpFlag {
		getopt.Usage()
		os.Exit(1)
	}

	args := getopt.Args()
	setup(args)

	switch command {
	case "setup":
		fmt.Println(args)
		runSetupPhase(args[1], args[2], args[3], args[4])
	case "enroll":
		if args[1] == "auditor" {
			runEnrollAuditor(args[2], args[3]) // TODO: optarg
		} else if args[1] == "user" {
			runEnrollUser(args[2], args[3], args[4]) // TODO: optarg
		}
	case "upload":
		runUploadPhase(args[1], data1)
	case "challenge":
		runUploadAuditingChal(args[1])
	case "proof":
		runUploadAuditingProof()
	case "audit":
		runVerifyAuditingProof(args[1])
	default:
		getopt.Usage()
		os.Exit(1)
	}

	if *simFlag {
		ledger.Dump("./cache/fake-ledger.json")
	}
}