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

type FileProperty struct {
	Owners []common.Address `json:'owners'`
}

type FakeLedger struct {
	Params Params `json:'params'`
	FileProperties map[string]*FileProperty `json:'fileProperties'`
	Accounts map[common.Address]pdp.PublicKeyData `json:'accounts'`
}

func GenFakeLedger() FakeLedger {
	var ledger FakeLedger
	ledger.FileProperties = make(map[string]*FileProperty)
	ledger.Accounts = make(map[common.Address]pdp.PublicKeyData)
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

func (this *FakeLedger) RegisterFile(_hash [32]byte, _addr common.Address) {
	var p FileProperty
	p.Owners = append(p.Owners, _addr)
	this.FileProperties[helper.Hex(_hash[:])] = &p
}

func (this *FileProperty) GetCreatorAddr() common.Address {
	return this.Owners[0]
}

func (this *FileProperty) ToXZ21File() pdp.XZ21File {
	var to pdp.XZ21File
	to.Owners = this.Owners
	return to
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

func (this *FakeLedger) SearchFile(_hash [32]byte) *pdp.XZ21File {
	if v, ok := this.FileProperties[helper.Hex(_hash[:])]; ok {
		tmp := v.ToXZ21File()
		return &tmp
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
