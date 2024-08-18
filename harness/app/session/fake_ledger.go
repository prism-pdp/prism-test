package session

import (
	"encoding/json"
	"io/ioutil"
	"github.com/ethereum/go-ethereum/common"
	"os"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type Params struct {
	Params string `json:'params'`
	G []byte `json:'G'`
	U []byte `json:'U'`
}

type AuditLog struct {
	ChalData []byte `json:'chal'`
	ProofData []byte `json:'proof'`
	Result bool `json:'result'`
}

type AuditReq struct {
	ChalData []byte `json:'chal'` // TODO: ChalBytes
	ProofData []byte `json:'proof'` // TODO: ProofBytes
}

type FakeLedger struct {
	Params Params `json:'params'`
	FileProperties map[string]*pdp.XZ21FileProperty `json:'fileProperties'`
	Accounts map[common.Address]pdp.PublicKeyData `json:'accounts'`
	Reqs map[string]*AuditReq `json:'auditReqs'`
	Logs map[string][]*AuditLog `json:'auditLogs'`
}

func GenFakeLedger() FakeLedger {
	var ledger FakeLedger
	ledger.FileProperties = make(map[string]*pdp.XZ21FileProperty)
	ledger.Accounts = make(map[common.Address]pdp.PublicKeyData)
	ledger.Reqs = make(map[string]*AuditReq)
	ledger.Logs = make(map[string][]*AuditLog)
	return ledger
}

func LoadFakeLedger(_path string) FakeLedger {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := ioutil.ReadAll(f)
	if err != nil { panic(err) }

	var ledger FakeLedger
	json.Unmarshal(s, &ledger)

	return ledger
}

func (this *FakeLedger) RegisterFileProperty(_hash [32]byte, _splitNum uint32, _addr common.Address) {
	var p pdp.XZ21FileProperty
	p.SplitNum = _splitNum
	p.Owners = append(p.Owners, _addr)
	this.FileProperties[helper.Hex(_hash[:])] = &p
}

func (this *FakeLedger) EnrollAccount(_addr common.Address, _key []byte) {
	var pkData pdp.PublicKeyData
	pkData.Key = _key
	this.Accounts[_addr] = pkData
}

func (this *FakeLedger) AppendAccount(_hash [32]byte, _addr common.Address) {
	if v, ok := this.FileProperties[helper.Hex(_hash[:])]; ok {
		v.Owners = append(v.Owners, _addr)
	}
}

func (this *FakeLedger) SearchFile(_hash [32]byte) *pdp.XZ21FileProperty {
	if v, ok := this.FileProperties[helper.Hex(_hash[:])]; ok {
		return v
	}
	return nil
}

func (this *FakeLedger) SearchPublicKey(_addr common.Address) *pdp.PublicKeyData {
	if v, ok := this.Accounts[_addr]; ok {
		return &v
	}
	return nil
}

func (this *FakeLedger) Dump(_path string) {
	s, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)
	if err != nil { panic(err) }
}
