package main

import (
	"fmt"
	"os"
	"strconv"
	"github.com/ethereum/go-ethereum/common"
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

	helper.Server = toString(helper.OptServer)
	helper.ContractAddr = toString(helper.OptContractAddr)
	helper.SenderAddr = common.HexToAddress(toString(helper.OptSenderAddr))
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
	if err != nil { helper.Panic(err) }

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
		if err != nil { helper.Panic(err) }

		if success {
			helper.PrintLog("Register an owner to a file (owner:%s, file:%s)", su.Name, hex)
		} else {
			helper.PrintLog("Failure registering an owner to a file (owner:%s, file:%s)", su.Name, hex)
		}
	} else {
		tmp, err := strconv.ParseUint(_chunkNum, 10, 32)
		if err != nil { helper.Panic(err) }
		chunkNum = uint32(tmp) // Overwrite chunkNum with user defined value

		// SU uploads the file.
		tag := su.PrepareUpload(data, chunkNum)

		// SP accepts the file.
		err = sp.UploadNewFile(data, &tag, su.Addr, &su.PublicKeyData)
		if err != nil { helper.Panic(err) }

		helper.PrintLog("Upload new file (owner:%s, file:%s)", su.Name, hex)
	}

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(&sp)
	helper.DumpEntity(&su)

	helper.PrintLog("Finish upload")
}

func runUploadAuditingChal(_name string, _ratioData string, _ratioFile string) {
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
	ratioData, err := strconv.ParseFloat(_ratioData, 64)
	if err != nil { helper.Panic(err) }

	ratioFile, err := strconv.ParseFloat(_ratioFile, 64)
	if err != nil { helper.Panic(err) }

	fileList := su.GetFileList()
	// SU generates challenge and requests to audit each file
	for i, f := range fileList {
		if helper.DrawLots(ratioFile) {
			helper.PrintLog("Upload auditing chal (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(fileList))
			chalData := su.GenAuditingChal(f, ratioData)
			su.UploadAuditingChal(f, chalData)
			su.AppendAuditingFile(f)
		}
	}

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(&su)

	helper.PrintLog("Finish challenge")
}

func runUploadAuditingProof(_nameSU string) {
	helper.PrintLog("Start proof")

	// --------------------------
	// Prepare entities
	// --------------------------
	var sp entity.Provider
	var su entity.User
	helper.LoadEntity(nameSP, &sp)
	helper.LoadEntity(_nameSU, &su)

	// --------------------------
	// Main processing
	// --------------------------
	// SP gets the file list for auditing from SU.
	auditingFileList := su.GetAuditingFileList()

	// SP gets challenge from blockchain.
	chalDataList := sp.DownloadAuditingChal(auditingFileList)

	for i, f := range auditingFileList {
		proofData := sp.GenAuditingProof(f, chalDataList[i])
		if len([]byte(proofData)) > 0 {
			helper.PrintLog("Upload auditing proof (entity:%s, file:%s, index:%d/%d)", sp.Name, helper.Hex(f[:]), i+1, len(auditingFileList))
			sp.UploadAuditingProof(f, proofData)
		} else {
			err := fmt.Errorf("[Failure] Generate auditing proof (entity:%s, file:%s, index:%d/%d)", sp.Name, helper.Hex(f[:]), i+1, len(auditingFileList))
			helper.Panic(err)
		}
	}

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(&sp)

	helper.PrintLog("Finish proof")
}

func runVerifyAuditingProof(_nameTPA string, _nameSU string) {
	helper.PrintLog("Start auditing")

	// --------------------------
	// Prepare entities
	// --------------------------
	var sp entity.Provider
	var tpa entity.Auditor
	var su entity.User
	helper.LoadEntity(nameSP, &sp)
	helper.LoadEntity(_nameTPA, &tpa)
	helper.LoadEntity(_nameSU, &su)

	// --------------------------
	// Main processing
	// --------------------------
	// TPA gets the file list for auditing from SU.
	auditingFileList := su.GetAuditingFileList()
	//
	reqDataList := tpa.GetAuditingReqList(auditingFileList)
	for i, f := range auditingFileList {
		helper.PrintLog("Download auditing req (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(auditingFileList))
	}

	var resultList []bool
	for i, f := range auditingFileList {
		helper.PrintLog("Verify auditing proof (file:%s, index:%d/%d)", helper.Hex(f[:]), i+1, len(auditingFileList))

		// TPA gets M (list of hash of chunks) from SP.
		owner, setDigest, tagDataSet := sp.PrepareVerificationData(f, reqDataList[i].ChalData)

		// TPA verifies proof.
		result, err := tpa.VerifyAuditingProof(f, tagDataSet, setDigest, reqDataList[i], owner)
		if err != nil { helper.Panic(err) }

		helper.PrintLog("Upload auditing result (file:%s, index:%d/%d, result:%t)", helper.Hex(f[:]), i+1, len(auditingFileList), result)
		tpa.UploadAuditingResult(f, result)

		resultList = append(resultList, result)

		if helper.OptDetectedList != nil && result == false {
			if foundFile := sp.SearchFile(f); foundFile != nil {
				path := sp.GetFilePath(foundFile)
				s := fmt.Sprintf("%s\n", path)
				helper.AppendFile(*helper.OptDetectedList, []byte(s))
			}
		}
	}

	tpa.SaveSummary(auditingFileList, resultList)

	// Update auditing req list of SU
	for _, v := range auditingFileList {
		su.RemoveAuditingFile(v)
	}

	// --------------------------
	// Save entities
	// --------------------------
	helper.DumpEntity(&sp)
	helper.DumpEntity(&tpa)
	helper.DumpEntity(&su)

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
		runUploadAuditingChal(args[1], args[2], args[3])
	case "proof":
		runUploadAuditingProof(args[1])
	case "audit":
		runVerifyAuditingProof(args[1], args[2])
	default:
		getopt.Usage()
		os.Exit(1)
	}

	if *helper.SimFlag {
		client.GetFakeLedger().Dump()
	}
}
