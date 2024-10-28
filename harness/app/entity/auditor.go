package entity

import (
	"encoding/json"
	"io"
	"os"
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

func LoadAuditor(_path string, _simFlag bool) *Auditor {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := io.ReadAll(f)
	if err != nil { panic(err) }

	a := new(Auditor)
	json.Unmarshal(s, &a)

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

func (this *Auditor) VerifyAuditingProof(_tagDataSet *pdp.TagDataSet, _digestSet *pdp.DigestSet, _auditingReqData *pdp.AuditingReqData, _owner common.Address) (bool, error) {
	xz21Param, err := this.client.GetParam()
	if err != nil { return false, err }

	params := pdp.GenParamFromXZ21Param(&xz21Param)

	auditingReq := _auditingReqData.Import(&params)
	tag := _tagDataSet.ImportSubset(&params, &auditingReq.Chal)

	account, err := this.client.GetAccount(_owner)
	if err != nil { panic(err) }

	pubKeyData := pdp.PublicKeyData{account.PubKey}
	pubKey := pubKeyData.Import(&params)

	result, err := pdp.VerifyProof(&params, &tag, _digestSet, &auditingReq.Chal, &auditingReq.Proof, pubKey.Key)
	if err != nil { return false, err }

	return result, nil
}

func (this *Auditor) UploadAuditingResult(_hash [32]byte, _result bool) {
	err := this.client.SetAuditingResult(_hash, _result)
	if err != nil { panic(err) }
}

func (this *Auditor) Dump(_pathDir string) {
	s, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	path := helper.MakeDumpPath(_pathDir, this.Name)
	f, err := os.Create(path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)

	if err != nil { panic(err) }
}
