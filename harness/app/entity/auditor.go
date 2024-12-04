package entity

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/helper"
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

func (this *Auditor) GetAuditingReqList(_fileList [][32]byte) ([]*pdp.AuditingReqData) {
	reqDataList := make([]*pdp.AuditingReqData, len(_fileList))

	for i, v := range _fileList {
		req, err := this.client.GetAuditingReq(v)
		if err != nil { panic(err) }

		if len(req.Chal) > 0 && len(req.Proof) > 0 {
			reqData := new(pdp.AuditingReqData)
			reqData.LoadFromXZ21(req)
			reqDataList[i] = reqData
		}
	}

	return reqDataList
}

func (this *Auditor) VerifyAuditingProof(_hash [32]byte, _setTagData pdp.TagDataSet, _setDigest pdp.DigestSet, _auditingReqData *pdp.AuditingReqData, _owner common.Address) (bool, error) {
	xz21Param, err := this.client.GetParam()
	if err != nil { return false, err }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	fileProp, err := this.client.SearchFile(_hash)
	if err != nil { helper.Panic(err) }

	auditingReq := _auditingReqData.Import(param)
	subsetTag := _setTagData.ImportSubset(param, fileProp.SplitNum, auditingReq.Chal)

	account, err := this.client.GetAccount(_owner)
	if err != nil { helper.Panic(err) }

	var pubKeyData pdp.PublicKeyData
	pubKeyData.Load(account.PubKey)
	pubKey := pubKeyData.Import(param)

	helper.PrintLog("Start verifying proof (splitNum:%d, blockCount:%d)", fileProp.SplitNum, auditingReq.Chal.GetTargetBlockCount())
	result, err := pdp.VerifyProof(param, fileProp.SplitNum, subsetTag, _setDigest, auditingReq.Chal, auditingReq.Proof, pubKey)
	if err != nil { return false, err }
	helper.PrintLog("Finish verifying proof (splitNum:%d, blockCount:%d)", fileProp.SplitNum, auditingReq.Chal.GetTargetBlockCount())

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