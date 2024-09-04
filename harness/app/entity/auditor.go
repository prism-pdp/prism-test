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
	client client.BaseClient
}

func MakeAuditor(_path string, _client client.BaseClient) *Auditor {
	if (helper.IsFile(_path)) {
		return LoadAuditor(_path, _client)
	} else {
		return GenAuditor(_client)
	}
}

func GenAuditor(_client client.BaseClient) *Auditor {
	e := new(Auditor)
	e.client = _client
	return e
}

func LoadAuditor(_path string, _client client.BaseClient) *Auditor {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := io.ReadAll(f)
	if err != nil { panic(err) }

	e := new(Auditor)
	json.Unmarshal(s, &e)

	e.client = _client

	return e
}

func (this *Auditor) GetAuditingReqList() ([][32]byte, []pdp.AuditingReq) {
	hashList, xz21ReqList, err := this.client.GetAuditingReqList()
	if err != nil { panic(err) }

	var reqList []pdp.AuditingReq
	for _, v := range xz21ReqList {
		var req pdp.AuditingReq
		req.Import(&v)
		reqList = append(reqList, req)
	}
	return hashList, reqList
}

func (this *Auditor) VerifyAuditProof(_tagData *pdp.TagData, _hashChunks [][]byte, _chalData *pdp.ChalData, _proofData *pdp.ProofData, _owner common.Address) (bool, error) {
	xz21Param, err := this.client.GetParam()
	if err != nil { return false, err }

	params := pdp.GenParamFromXZ21Param(&xz21Param)

	tag := _tagData.Import(&params)
	chal := _chalData.Import(&params)
	proof := _proofData.Import(&params)

	account, err := this.client.GetAccount(_owner)
	if err != nil { panic(err) }

	pubKeyData := pdp.PublicKeyData{account.PubKey}
	pubKey := pubKeyData.Import(&params)

	result := pdp.VerifyProof(&params, &tag, _hashChunks, &chal, &proof, pubKey.Key)

	return result, nil
}

func (this *Auditor) UploadAuditResult(_hash [32]byte, _result bool) {
	err := this.client.SetAuditingResult(_hash, _result)
	if err != nil { panic(err) }
}

func (this *Auditor) Dump(_path string) {
	s, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)

	if err != nil { panic(err) }
}