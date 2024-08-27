package entity

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/session"
)

type Auditor struct {
	session session.Session
}

func GenAuditor(_session session.Session) *Auditor {
	e := new(Auditor)
	e.session = _session
	return e
}

func LoadAuditor(_path string, _session session.Session) *Auditor {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := io.ReadAll(f)
	if err != nil { panic(err) }

	e := new(Auditor)
	json.Unmarshal(s, &e)

	e.session = _session

	return e
}

func (this *Auditor) DownloadAuditProof() ([][32]byte, []pdp.ChalData, []pdp.ProofData) {
	hashList, chalDataList, proofDataList := this.session.DownloadAuditChallenAndProof()
	return hashList, chalDataList, proofDataList
}

func (this *Auditor) VerifyAuditProof(_tagData *pdp.TagData, _hashChunks [][]byte, _chalData *pdp.ChalData, _proofData *pdp.ProofData, _owner common.Address) (bool, error) {
	xz21Params, err := this.session.GetPara()
	if err != nil { return false, err }

	params := pdp.GenParamFromXZ21Para(&xz21Params)

	tag := _tagData.Import(&params)
	chal := _chalData.Import(&params)
	proof := _proofData.Import(&params)

	pubKeyBytes, found := this.session.SearchPublicKey(_owner)
	if found == false { return false, fmt.Errorf("Owner is not found.") }

	pubKeyData := pdp.PublicKeyData{pubKeyBytes}
	pubKey := pubKeyData.Import(&params)

	result := pdp.VerifyProof(&params, &tag, _hashChunks, &chal, &proof, pubKey.Key)

	return result, nil
}

func (this *Auditor) UploadAuditResult(_hash [32]byte, _result bool) {
	err := this.session.UploadAuditResult(_hash, _result)
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