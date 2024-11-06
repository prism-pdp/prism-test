package main

import (
	"os"
	"strconv"

	"github.com/pborman/getopt/v2"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/entity"
	"github.com/dpduado/dpduado-test/harness/helper"
)

var command string

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
	command = _opts[0]

	// server := toString(optServer)
	// contractAddr:= toString(optContractAddr)
	helper.SenderAddr = toString(helper.OptSenderAddr)
	helper.SenderPrivKey = toString(helper.OptSenderPrivKey)

	if *helper.SimFlag {
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
	sm := entity.GenManager(nameSM, _smAddr, _smPrivKey, *helper.SimFlag)
	sp := entity.GenProvider(nameSP, _spAddr, _spPrivKey, *helper.SimFlag)

	// --------------------------
	// Main processing
	// --------------------------
	sm.RegisterParam()
	helper.PrintLog("Register param")

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(sm)
	helper.DumpEntity(sp)

	helper.PrintLog("Finish setup")
}

func runEnrollUser(_name string, _addr string, _privKey string) {
	helper.PrintLog("Start enroll user")

	// --------------------------
	// Prepare entities
	// --------------------------
	var sm entity.Manager
	helper.LoadEntity(nameSM, &sm)
	su := entity.GenUser(_name, _addr, _privKey, sm.GetParam(), *helper.SimFlag)

	// --------------------------
	// Main processing
	// --------------------------
	sm.EnrollUser(su)
	helper.PrintLog("enroll service user (name:%s)", su.Name)

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(&sm)
	helper.DumpEntity(su)

	helper.PrintLog("Finish enroll user")
}

func runEnrollAuditor(_name string, _addr string) {
	helper.PrintLog("Start enroll auditor")

	// --------------------------
	// Prepare entities
	// --------------------------
	var sm entity.Manager
	helper.LoadEntity(nameSM, &sm)
	tpa := entity.GenAuditor(_name, _addr, *helper.SimFlag)

	// --------------------------
	// Main processing
	// --------------------------
	sm.EnrollAuditor(tpa)
	helper.PrintLog("enroll auditor (name:%s)", tpa.Name)

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(&sm)
	helper.DumpEntity(tpa)

	helper.PrintLog("Finish enroll auditor")
}

func runUploadPhase(_name string, _path string, _chunkNum string) {
	helper.PrintLog("Start upload")

	// --------------------------
	// Prepare entities
	// --------------------------
	var sp entity.Provider
	var su entity.User
	helper.LoadEntity(nameSP, &sp)
	helper.LoadEntity(_name, &su)

	// --------------------------
	// Main processing
	// --------------------------
	// Read file
	data, err := helper.ReadFile(_path)
	if err != nil { panic(err) }

	helper.PrintLog("Read file (filesize:%d, path:%s)", len(data), _path)

	// SU inquires with SP whether the data is uploaded.
	isUploaded, chunkNum := sp.IsUploaded(data)

	// Processing differs depending on whether the file has already been uploaded or not.
	digest := helper.CalcDigest(data)
	hex := helper.Hex(digest[:])
	if isUploaded {
		// SP generates a challenge for deduplication.
		chalData := sp.GenDedupChal(data, su.Addr)

		// (SP sends the challenge to SU.)

		// SU generates a proof to prove ownership of the data to be uploaded.
		// Use chunkNum which has been already determined
		proofData := su.GenDedupProof(chalData, data, chunkNum)

		// (SU sends the proof to SP.)

		// SP verifies the proof.
		success, err := sp.RegisterOwnerToFile(&su, data, chalData, proofData)
		if err != nil { panic(err) }

		if success {
			helper.PrintLog("Register an owner to a file (owner:%s, file:%s)", su.Name, hex)
		} else {
			helper.PrintLog("Failure registering an owner to a file (owner:%s, file:%s)", su.Name, hex)
		}
	} else {
		tmp, err := strconv.ParseUint(_chunkNum, 10, 32)
		if err != nil { panic(err) }
		chunkNum = uint32(tmp) // Overwrite chunkNum with user defined value

		// SU uploads the file.
		tag := su.PrepareUpload(data, chunkNum)

		// SP accepts the file.
		err = sp.UploadNewFile(data, &tag, su.Addr, &su.PublicKeyData)
		if err != nil { panic(err) }

		helper.PrintLog("Upload new file (owner:%s, file:%s)", su.Name, hex)
	}

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(&sp)
	helper.DumpEntity(&su)

	helper.PrintLog("Finish upload")
}

func runUploadAuditingChal(_name string, _ratio string) {
	helper.PrintLog("Start challenge")

	// --------------------------
	// Prepare entities
	// --------------------------
	var su entity.User
	helper.LoadEntity(_name, &su)

	// --------------------------
	// Main processing
	// --------------------------
	// SU gets the list of his/her files.
	ratio, err := strconv.ParseFloat(_ratio, 64)
	if err != nil { panic(err) }

	fileList := su.GetFileList()
	// SU generates challenge and requests to audit each file
	for i, f := range fileList {
		helper.PrintLog("Upload auditing chal (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(fileList))
		chalData := su.GenAuditingChal(f, ratio)
		su.UploadAuditingChal(f, chalData)
	}

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(&su)

	helper.PrintLog("Finish challenge")
}

func runUploadAuditingProof() {
	helper.PrintLog("Start proof")

	// --------------------------
	// Prepare entities
	// --------------------------
	var sp entity.Provider
	helper.LoadEntity(nameSP, &sp)

	// --------------------------
	// Main processing
	// --------------------------
	// SP gets challenge from blockchain.
	fileList, chalDataList := sp.DownloadAuditingChal()
	for i, h := range fileList {
		helper.PrintLog("Download auditing chal (file:%s, index:%d/%d)", helper.Hex(h[:]), i+1, len(fileList))
	}

	for i, f := range fileList {
		proofData := sp.GenAuditingProof(f, chalDataList[i])
		helper.PrintLog("Upload auditing proof (entity:%s, file:%s, index:%d/%d)", sp.Name, helper.Hex(f[:]), i+1, len(fileList))
		sp.UploadAuditingProof(f, proofData)
	}

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(&sp)

	helper.PrintLog("Finish proof")
}

func runVerifyAuditingProof(_name string) {
	helper.PrintLog("Start auditing")

	// --------------------------
	// Prepare entities
	// --------------------------
	var sp entity.Provider
	var tpa entity.Auditor
	helper.LoadEntity(nameSP, &sp)
	helper.LoadEntity(_name, &tpa)

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
	helper.DumpEntity(&sp)
	helper.DumpEntity(&tpa)

	helper.PrintLog("Finish auditing")
}

func main() {

	helper.SetupOpt()

	getopt.Parse()

	if *helper.HelpFlag {
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
			runEnrollAuditor(args[2], args[3])
		} else if args[1] == "user" {
			runEnrollUser(args[2], args[3], args[4])
		}
	case "upload":
		runUploadPhase(args[1], args[2], args[3])
	case "challenge":
		runUploadAuditingChal(args[1], args[2])
	case "proof":
		runUploadAuditingProof()
	case "audit":
		runVerifyAuditingProof(args[1])
	default:
		getopt.Usage()
		os.Exit(1)
	}

	if *helper.SimFlag {
		client.GetFakeLedger().Dump()
	}
}
