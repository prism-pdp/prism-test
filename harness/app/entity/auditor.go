package entity

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/helper"
)

type Auditor struct {
	Name string
	Addr common.Address

	client client.BaseClient
}

func GenAuditor(_name string, _addr string, _simFlag bool) *Auditor {
	a := new(Auditor)
	a.Name = _name
	a.Addr = common.HexToAddress(_addr)

	if _simFlag {
		a.SetupSimClient(client.GetFakeLedger())
	}

	return a
}


func (this *Auditor) SetupSimClient(_ledger *client.FakeLedger) {
	this.client = client.NewSimClient(_ledger, this.Addr)
}

func (this *Auditor) GetAuditingReqList() ([][32]byte, []pdp.AuditingReqData) {
	hashList, xz21ReqList, err := this.client.GetAuditingReqList()
	if err != nil { panic(err) }

	var reqDataList []pdp.AuditingReqData
	for _, v := range xz21ReqList {
		var reqData pdp.AuditingReqData
		reqData.LoadFromXZ21(&v)
		reqDataList = append(reqDataList, reqData)
	}
	return hashList, reqDataList
}

func (this *Auditor) VerifyAuditingProof(_hash [32]byte, _setTagData pdp.TagDataSet, _setDigest pdp.DigestSet, _auditingReqData *pdp.AuditingReqData, _owner common.Address) (bool, error) {
	xz21Param, err := this.client.GetParam()
	if err != nil { return false, err }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	fileProp, err := this.client.SearchFile(_hash)
	if err != nil { panic(err) }

	auditingReq := _auditingReqData.Import(param)
	subsetTag := _setTagData.ImportSubset(param, fileProp.SplitNum, auditingReq.Chal)

	account, err := this.client.GetAccount(_owner)
	if err != nil { panic(err) }

	var pubKeyData pdp.PublicKeyData
	pubKeyData.Load(account.PubKey)
	pubKey := pubKeyData.Import(param)

	helper.PrintLog("Verify proof (file:%s, splitNum:%d, blockCount:%d)", helper.Hex(_hash[:]), fileProp.SplitNum, auditingReq.Chal.GetTargetBlockCount())
	result, err := pdp.VerifyProof(param, fileProp.SplitNum, subsetTag, _setDigest, auditingReq.Chal, auditingReq.Proof, pubKey)
	if err != nil { return false, err }

	return result, nil
}

func (this *Auditor) UploadAuditingResult(_hash [32]byte, _result bool) {
	err := this.client.SetAuditingResult(_hash, _result)
	if err != nil { panic(err) }
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
	// Do nothing
}