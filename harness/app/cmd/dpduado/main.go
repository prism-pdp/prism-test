package main

import (
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

const nameSM = "sm"
const nameSP = "sp"

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
	chunkNum = uint32(100) // TODO: change to option

	command = _opts[0]

	// server := toString(optServer)
	// contractAddr:= toString(optContractAddr)
	senderAddr = toString(optSenderAddr)
	senderPrivKey = toString(optSenderPrivKey)

	if *simFlag {
		// make fake ledger
		if command == "setup" {
			client.GenFakeLedger()
		} else {
			client.LoadFakeLedger()
		}
	}
}

func runSetupPhase(_smAddr string, _smPrivKey string, _spAddr string, _spPrivKey string) {
	helper.PrintLog("Start setup")

	// --------------------------
	// Prepare entities
	// --------------------------
	sm := entity.GenManager(nameSM, _smAddr, _smPrivKey, *simFlag)
	sp := entity.GenProvider(nameSP, _spAddr, _spPrivKey, *simFlag)

	// --------------------------
	// Main processing
	// --------------------------
	sm.RegisterParam()
	helper.PrintLog("Register param")

	// --------------------------
	// Save entities
	// --------------------------
	sm.Dump()
	sp.Dump()

	helper.PrintLog("Finish setup")
}

func runEnrollUser(_name string, _addr string, _privKey string) {
	helper.PrintLog("Start enroll user")

	// --------------------------
	// Prepare entities
	// --------------------------
	sm := entity.LoadManager(nameSM, *simFlag)
	su := entity.GenUser(_name, _addr, _privKey, sm.GetParam(), *simFlag)

	// --------------------------
	// Main processing
	// --------------------------
	sm.EnrollUser(su)
	helper.PrintLog("enroll service user (name:%s)", su.Name)

	// --------------------------
	// Save entities
	// --------------------------
	sm.Dump()
	su.Dump()

	helper.PrintLog("Finish enroll user")
}

func runEnrollAuditor(_name string, _addr string) {
	helper.PrintLog("Start enroll auditor")

	// --------------------------
	// Prepare entities
	// --------------------------
	sm := entity.LoadManager(nameSM, *simFlag)
	tpa := entity.GenAuditor(_name, _addr, *simFlag)

	// --------------------------
	// Main processing
	// --------------------------
	sm.EnrollAuditor(tpa)
	helper.PrintLog("enroll auditor (name:%s)", tpa.Name)

	// --------------------------
	// Save entities
	// --------------------------
	sm.Dump()
	tpa.Dump()

	helper.PrintLog("Finish enroll auditor")
}

func runUploadPhase(_name string, _path string) {
	helper.PrintLog("Start upload")

	// --------------------------
	// Prepare entities
	// --------------------------
	sp := entity.LoadProvider(nameSP, *simFlag)
	su := entity.LoadUser(_name, *simFlag)

	// --------------------------
	// Main processing
	// --------------------------
	// Read file
	data, err := helper.ReadFile(_path)
	if err != nil { panic(err) }

	// SU checks whether data is uploaded.
	isUploaded := su.IsUploaded(data)

	// Processing differs depending on whether the file has already been uploaded or not.
	digest := helper.CalcDigest(data)
	hex := helper.Hex(digest[:])
	if isUploaded {
		// SP generates a challenge for deduplication.
		chalData := sp.GenDedupChal(data, su.Addr)

		// (SP sends the challenge to SU.)

		// SU generates a proof to prove ownership of the data to be uploaded.
		proofData := su.GenDedupProof(chalData, data, chunkNum)

		// (SU sends the proof to SP.)

		// SP verifies the proof.
		success, err := sp.RegisterOwnerToFile(su, data, chalData, proofData)
		if err != nil { panic(err) }

		if success {
			helper.PrintLog("Register an owner to a file (owner:%s, file:%s)", su.Name, hex)
		} else {
			helper.PrintLog("Failure registering an owner to a file (owner:%s, file:%s)", su.Name, hex)
		}
	} else {
		// SU uploads the file.
		tag := su.PrepareUpload(data, chunkNum)

		// SP accepts the file.
		err := sp.UploadNewFile(data, &tag, su.Addr, &su.PublicKeyData)
		if err != nil { panic(err) }

		helper.PrintLog("Upload new file (owner:%s, file:%s)", su.Name, hex)
	}

	// --------------------------
	// Save entities
	// --------------------------
	sp.Dump()
	su.Dump()

	helper.PrintLog("Finish upload")
}

func runUploadAuditingChal(_name string) {
	helper.PrintLog("Start challenge")

	// --------------------------
	// Prepare entities
	// --------------------------
	su := entity.LoadUser(_name, *simFlag)

	// --------------------------
	// Main processing
	// --------------------------
	// SU gets the list of his/her files.
	fileList := su.GetFileList()
	// SU generates challenge and requests to audit each file
	for i, f := range fileList {
		helper.PrintLog("Upload auditing chal (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(fileList))
		chalData := su.GenAuditingChal(f)
		su.UploadAuditingChal(f, chalData)
	}

	// --------------------------
	// Save entities
	// --------------------------
	su.Dump()

	helper.PrintLog("Finish challenge")
}

func runUploadAuditingProof() {
	helper.PrintLog("Start proof")

	// --------------------------
	// Prepare entities
	// --------------------------
	sp := entity.LoadProvider(nameSP, *simFlag)

	// --------------------------
	// Main processing
	// --------------------------
	// SP gets challenge from blockchain.
	fileList, chalDataList := sp.DownloadAuditingChal()
	for i, h := range fileList {
		helper.PrintLog("Download auditing chal (file:%s, index:%d/%d)", helper.Hex(h[:]), i+1, len(fileList))
	}

	for i, f := range fileList {
		helper.PrintLog("Upload auditing proof (entity:%s, file:%s, index:%d/%d)", sp.Name, helper.Hex(f[:]), i+1, len(fileList))
		proofData := sp.GenAuditingProof(f, chalDataList[i])
		sp.UploadAuditingProof(f, proofData)
	}

	// --------------------------
	// Save entities
	// --------------------------
	sp.Dump()

	helper.PrintLog("Finish proof")
}

func runVerifyAuditingProof(_name string) {
	helper.PrintLog("Start auditing")

	// --------------------------
	// Prepare entities
	// --------------------------
	sp := entity.LoadProvider(nameSP, *simFlag)
	tpa := entity.LoadAuditor(_name, *simFlag)

	// --------------------------
	// Main processing
	// --------------------------
	// TPA gets challenge and proof from blockchain.
	fileList, reqDataList := tpa.GetAuditingReqList()
	for i, f := range fileList {
		helper.PrintLog("Download auditing req (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(fileList))
	}

	for i, f := range fileList {
		// TPA gets M (list of hash of chunks) from SP.
		owner, setDigest, tagDataSet := sp.PrepareVerificationData(f, reqDataList[i].ChalData)

		// TPA verifies proof.
		result, err := tpa.VerifyAuditingProof(f, tagDataSet, setDigest, &reqDataList[i], owner)
		if err != nil { panic(err) }

		helper.PrintLog("Upload auditing result (file:%s, result:%t)", helper.Hex(f[:]), result)
		tpa.UploadAuditingResult(f, result)
	}

	// --------------------------
	// Save entities
	// --------------------------
	sp.Dump()
	tpa.Dump()

	helper.PrintLog("Finish auditing")
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
		runSetupPhase(args[1], args[2], args[3], args[4])
	case "enroll":
		if args[1] == "auditor" {
			runEnrollAuditor(args[2], args[3]) // TODO: optarg
		} else if args[1] == "user" {
			runEnrollUser(args[2], args[3], args[4]) // TODO: optarg
		}
	case "upload":
		runUploadPhase(args[1], args[2])
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
		client.GetFakeLedger().Dump()
	}
}
