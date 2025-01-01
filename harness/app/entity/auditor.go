package entity

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/prism-pdp/prism-go/xz21"

	"github.com/prism-pdp/prism-test/harness/client"
	"github.com/prism-pdp/prism-test/harness/helper"
)

type AuditingSummary struct {
	File []string
	Result []bool
}

type Auditor struct {
	Name string
	Addr common.Address
	Summary []*AuditingSummary

	client client.BaseClient
}

func GenAuditor(_name string, _addr string, _simFlag bool) *Auditor {
	a := new(Auditor)
	a.Name = _name
	a.Addr = common.HexToAddress(_addr)

	if _simFlag {
		a.SetupSimClient(client.GetFakeLedger())
	} else {
		a.SetupEthClient()
	}

	return a
}

func NewAuditingSummary(_len int) *AuditingSummary{
	obj := new(AuditingSummary)
	obj.File = make([]string, _len)
	obj.Result = make([]bool, _len)
	return obj
}

func (this *Auditor) SetupSimClient(_ledger *client.FakeLedger) {
	this.client = client.NewSimClient(_ledger, this.Addr)
}

func (this *Auditor) SetupEthClient() {
	this.client = client.NewEthClient(helper.Server, helper.ContractAddr, helper.SenderPrivKey, helper.SenderAddr)
}

func (this *Auditor) GetWaitingResultList(_fileList [][32]byte) ([]*pdp.AuditingLogData) {
	logDataList := make([]*pdp.AuditingLogData, len(_fileList))

	for i, v := range _fileList {
		log, err := this.client.GetLatestAuditingLog(v)
		if err != nil { panic(err) }

		if log.Stage == pdp.WaitingResult {
			logData := new(pdp.AuditingLogData)
			logData.LoadFromXZ21(log)
			logDataList[i] = logData
		}
	}

	return logDataList
}

func (this *Auditor) VerifyAuditingProof(_hash [32]byte, _setTagData pdp.TagDataSet, _setDigest pdp.DigestSet, _auditingLogData *pdp.AuditingLogData, _owner common.Address) (bool, error) {
	xz21Param, err := this.client.GetParam()
	if err != nil { return false, err }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	fileProp, err := this.client.SearchFile(_hash)
	if err != nil { helper.Panic(err) }

	auditingLog := _auditingLogData.Import(param)
	subsetTag := _setTagData.ImportSubset(param, fileProp.SplitNum, auditingLog.Chal)

	account, err := this.client.GetAccount(_owner)
	if err != nil { helper.Panic(err) }

	var pubKeyData pdp.PublicKeyData
	pubKeyData.Load(account.PubKey)
	pubKey := pubKeyData.Import(param)

	helper.PrintLog("Start verifying proof (splitNum:%d, blockCount:%d)", fileProp.SplitNum, auditingLog.Chal.GetTargetBlockCount())
	result, err := pdp.VerifyProof(param, fileProp.SplitNum, subsetTag, _setDigest, auditingLog.Chal, auditingLog.Proof, pubKey)
	if err != nil { return false, err }
	helper.PrintLog("Finish verifying proof (splitNum:%d, blockCount:%d)", fileProp.SplitNum, auditingLog.Chal.GetTargetBlockCount())

	return result, nil
}

func (this *Auditor) UploadAuditingResult(_hash [32]byte, _result bool) {
	err := this.client.SetAuditingResult(_hash, _result)
	if err != nil { panic(err) }
}

func (this *Auditor) SaveSummary(_fileList [][32]byte, _resultList []bool) {
	len := len(_fileList)
	s := NewAuditingSummary(len)

	for i := 0; i < len; i++ {
		s.File[i] = helper.Hex(_fileList[i][:])
		s.Result[i] = _resultList[i]
	}

	this.Summary = append(this.Summary, s)
}

func (this *Auditor) ListCorruptedFiles() []string {
	var list []string

	len := len(this.Summary)
	if len == 0 {
		return list
	}

	s := this.Summary[len-1]
	for i, v := range s.Result {
		if !v {
			list = append(list, s.File[i])
		}
	}

	return list
}

func (this *Auditor) GetName() string {
	return this.Name
}

func (this *Auditor) ToJson() (string, error) {
	b, err := json.MarshalIndent(this, "", "\t")
	return string(b), err
}

func (this *Auditor) FromJson(_json []byte, _simFlag bool) {
	json.Unmarshal(_json, this)

	if _simFlag {
		this.SetupSimClient(client.GetFakeLedger())
	}
}

func (this *Auditor) AfterLoad() {
	if *helper.SimFlag {
		this.SetupSimClient(client.GetFakeLedger())
	} else {
		this.SetupEthClient()
	}
}