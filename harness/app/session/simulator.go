package session

import (
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/slices"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type SimSession struct {
	Addr common.Address `json:'addr'`
	Ledger *FakeLedger  `json:'ledger'`
}

func (this *SimSession) Setup(_addr common.Address, _ledger *FakeLedger) {
	this.Addr = _addr
	this.Ledger = _ledger
}

func (this *SimSession) GetPara() (pdp.XZ21Para, error) {
	var xz21Para pdp.XZ21Para
	xz21Para.Params = this.Ledger.Params.Params
	xz21Para.G = this.Ledger.Params.G
	xz21Para.U = this.Ledger.Params.U
	return xz21Para, nil
}

func (this *SimSession) RegisterPara(_params string, _g []byte, _u []byte) {
	this.Ledger.Params.Params = _params
	this.Ledger.Params.G = _g
	this.Ledger.Params.U = _u
}

func (this *SimSession) RegisterFileProperty(_hash [32]byte, _splitNum uint32, _owner common.Address) {
	this.Ledger.RegisterFileProperty(_hash, _splitNum, _owner)
}

func (this *SimSession) FetchFileList() [][32]byte {
	var fileList [][32]byte
	for hashHex, prop := range this.Ledger.FileProperties {
		if slices.Contains(prop.Owners, this.Addr) {
			hash, err := helper.DecodeHex(hashHex)
			if err != nil { panic(err) }
			fileList = append(fileList, [32]byte(hash))
		}
	}
	return fileList
}

func (this *SimSession) SearchFile(_hash [32]byte) *pdp.XZ21FileProperty {
	return this.Ledger.SearchFile(_hash)
}

func (this *SimSession) SearchPublicKey(_addr common.Address) ([]byte, bool) {
	if v, ok := this.Ledger.Accounts[_addr]; ok {
		return v.Key, true
	}
	return []byte{0}, false
}

func (this *SimSession) EnrollAccount(_addr common.Address, _pubKey []byte) {
	this.Ledger.EnrollAccount(_addr, _pubKey)
}

func (this *SimSession) AppendAccount(_hash [32]byte, _addr common.Address) {
	this.Ledger.AppendAccount(_hash, _addr)
}

func (this *SimSession) UploadChallen(_hash [32]byte, _chalBytes []byte) {
	var req AuditReq

	req.ChalData = _chalBytes
	hashHex := helper.Hex(_hash[:])
	this.Ledger.Reqs[hashHex] = &req
}

func (this *SimSession) DownloadChallen() ([][32]byte, []pdp.ChalData) {
	hashList := make([][32]byte, 0)
	chalDataList := make([]pdp.ChalData, 0)
	for k, v := range this.Ledger.Reqs {
		if len(v.ProofData) == 0 {
			h, err := helper.DecodeHex(k)
			if err != nil { panic(err) }
			hashList = append(hashList, [32]byte(h))

			chalData, err := pdp.DecodeToChalData(v.ChalData)
			if err != nil { panic(err) }
			chalDataList = append(chalDataList, chalData)
		}
	}
	return hashList, chalDataList
}

func (this *SimSession) UploadProof(_hash [32]byte, _proofBytes []byte) {
	this.Ledger.Reqs[helper.Hex(_hash[:])].ProofData = _proofBytes
}

func (this *SimSession) DownloadAuditChallenAndProof() ([][32]byte, []pdp.ChalData, []pdp.ProofData) {
	hashList := make([][32]byte, 0)
	chalDataList := make([]pdp.ChalData, 0)
	proofDataList := make([]pdp.ProofData, 0)
	for k, v := range this.Ledger.Reqs {
		if len(v.ProofData) > 0 {
			h, err := helper.DecodeHex(k)
			if err != nil { panic(err) }
			hashList = append(hashList, [32]byte(h))

			chalData, err := pdp.DecodeToChalData(v.ChalData)
			if err != nil { panic(err) }
			chalDataList = append(chalDataList, chalData)

			proofData, err := pdp.DecodeToProofData(v.ProofData)
			if err != nil { panic(err) }
			proofDataList = append(proofDataList, proofData)
		}
	}
	return hashList, chalDataList, proofDataList
}

func (this *SimSession) FetchAuditingReqList() [][32]byte {
	var fileList [][32]byte
	for hashHex, _ := range this.Ledger.Reqs {
		hash, err := helper.DecodeHex(hashHex)
		if err != nil { panic(err) }
		fileList = append(fileList, [32]byte(hash))
	}
	return fileList
}