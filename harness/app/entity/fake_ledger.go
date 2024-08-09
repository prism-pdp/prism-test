package entity

import (
	"encoding/json"
	"io/ioutil"
	"github.com/ethereum/go-ethereum/common"
	"os"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type FileProperty struct {
	Tags    pdp.TagsData `json:'tags'`
	Owners []common.Address `json:'owners'`
}

type FakeLedger struct {
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

func (this *FakeLedger) RegisterFile(_hash [32]byte, _tags *pdp.TagsData, _addr common.Address) {
	var p FileProperty
	p.Tags = *_tags
	p.Owners = append(p.Owners, _addr)
	this.FileProperties[helper.Hex(_hash[:])] = &p
}

func (this *FileProperty) GetCreatorAddr() common.Address {
	return this.Owners[0]
}

func (this *FakeLedger) RegisterAccount(_addr common.Address, _key *pdp.PublicKeyData) {
	this.Accounts[_addr] = *_key
}

func (this *FakeLedger) AppendAccount(_hash [32]byte, _addr common.Address) {
	if v, ok := this.FileProperties[helper.Hex(_hash[:])]; ok {
		v.Owners = append(v.Owners, _addr)
	}
}

func (this *FileProperty) GetNumTags() uint32 {
	return uint32(len(this.Tags.Tags))
}

func (this *FakeLedger) GetNumTags(_hash [32]byte) uint32 {
	if v, ok := this.FileProperties[helper.Hex(_hash[:])]; ok {
		return v.GetNumTags()
	}
	return 0
}

func (this *FakeLedger) SearchFile(_hash [32]byte) *FileProperty {
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
