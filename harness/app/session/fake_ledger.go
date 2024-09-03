package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/ethereum/go-ethereum/common"
	"os"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type Params struct {
	P string `json:'params'`
	G []byte `json:'G'`
	U []byte `json:'U'`
}

type FakeLedger struct {
	Params Params `json:'params'`
	FileProperties map[string]*pdp.XZ21FileProperty `json:'fileProperties'`
	Accounts map[common.Address]*pdp.XZ21Account `json:'accounts'`
	Reqs map[string]*pdp.XZ21AuditingReq `json:'auditReqs'`
	Logs map[string][]*pdp.XZ21AuditingLog `json:'auditLogs'`
}

func GenFakeLedger() FakeLedger {
	var ledger FakeLedger
	ledger.FileProperties = make(map[string]*pdp.XZ21FileProperty)
	ledger.Accounts = make(map[common.Address]*pdp.XZ21Account)
	ledger.Reqs = make(map[string]*pdp.XZ21AuditingReq)
	ledger.Logs = make(map[string][]*pdp.XZ21AuditingLog)
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

func (this *FakeLedger) RegisterFile(_hash [32]byte, _splitNum uint32, _addr common.Address) {
	var p pdp.XZ21FileProperty
	p.SplitNum = _splitNum
	p.Creator = _addr
	this.FileProperties[helper.Hex(_hash[:])] = &p

	this.Accounts[_addr].FileList = append(this.Accounts[_addr].FileList, _hash)
}

func (this *FakeLedger) EnrollAccount(_addr common.Address, _key []byte) error {
	var a pdp.XZ21Account
	a.PubKey = _key
	this.Accounts[_addr] = &a

	return nil
}

func (this *FakeLedger) AppendOwner(_hash [32]byte, _addr common.Address) error {
	if _, ok := this.FileProperties[helper.Hex(_hash[:])]; !ok {
		return fmt.Errorf("Unknown file")
	}
	if _, ok := this.Accounts[_addr]; !ok {
		return fmt.Errorf("Unknown account")
	}

	this.Accounts[_addr].FileList = append(this.Accounts[_addr].FileList, _hash)

	return nil
}

func (this *FakeLedger) SearchFile(_hash [32]byte) (pdp.XZ21FileProperty, error) {
	if v, ok := this.FileProperties[helper.Hex(_hash[:])]; ok {
		return *v, nil
	}
	return pdp.XZ21FileProperty{}, nil
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
