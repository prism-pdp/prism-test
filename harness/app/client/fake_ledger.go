package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"github.com/ethereum/go-ethereum/common"
	"os"
	// "slices"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type FakeLedger struct {
	Param pdp.XZ21Param `json:'param'`
	FileProperties map[string]*pdp.XZ21FileProperty `json:'fileProperties'`
	UserAccountTable map[common.Address]*pdp.XZ21Account `json:'accounts'`
	AuditorAddrList []common.Address `json:'addrListTPA'`
	Reqs map[string]*pdp.XZ21AuditingReq `json:'auditReqs'`
	Logs map[string][]*pdp.XZ21AuditingLog `json:'auditLogs'`
}

func GenFakeLedger() FakeLedger {
	var ledger FakeLedger
	ledger.FileProperties = make(map[string]*pdp.XZ21FileProperty)
	ledger.UserAccountTable = make(map[common.Address]*pdp.XZ21Account)
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

	this.UserAccountTable[_addr].FileList = append(this.UserAccountTable[_addr].FileList, _hash)
}

func (this *FakeLedger) EnrollAuditor(_addr common.Address) error {
	return this.enroll(0, _addr, []byte{})
}

func (this *FakeLedger) EnrollUser(_addr common.Address, _key []byte) error {
	return this.enroll(1, _addr, _key)
}

func (this *FakeLedger) enroll(_type int, _addr common.Address, _key []byte) error {
	if _type == 0 {
		this.AuditorAddrList = append(this.AuditorAddrList, _addr)
		// slices.Sort(this.AuditorAddrList) TODO
		// this.AuditorAddrList = slices.Compact(this.AuditorAddrList) TODO
	} else if _type == 1 {
		var a pdp.XZ21Account
		a.PubKey = _key
		this.UserAccountTable[_addr] = &a
	}

	return nil
}

func (this *FakeLedger) AppendOwner(_hash [32]byte, _addr common.Address) error {
	if _, ok := this.FileProperties[helper.Hex(_hash[:])]; !ok {
		return fmt.Errorf("Unknown file")
	}
	if _, ok := this.UserAccountTable[_addr]; !ok {
		return fmt.Errorf("Unknown account")
	}

	this.UserAccountTable[_addr].FileList = append(this.UserAccountTable[_addr].FileList, _hash)

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
