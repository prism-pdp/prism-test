package main

import (
	"encoding/json"
	"os"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

type Property struct {
	Size uint32     `json:'size'`
	Hash [][]byte   `json:'hash'`
}

type Owner struct {
	ID int      `json:'id'`
	Tags [][]byte `json:'meta'`
}

type File struct {
	Hash [32]byte  `json:'hash'`
	Data []byte    `json:'data'`
	Prop *Property `json:'property'`
	Owners []Owner `json:'owners'`
}

type RamSP struct {
	Storage []File `json:'storage'`
}

func GenRamSP() *RamSP {
	ram := new(RamSP)
	return ram
}

func (this *RamSP) NewFile(_entity int, _hash [32]byte, _data []byte, _meta *pdp.Metadata) {
	var prop Property
	prop.Size = _meta.Size
	prop.Hash = _meta.Hash

	var owner Owner
	owner.ID  = _entity
	for i := range _meta.Tags {
		owner.Tags = append(owner.Tags, _meta.Tags[i].Bytes())
	}

	var file File
	file.Hash = _hash
	file.Data = _data
	file.Prop = &prop
	file.Owners = append(file.Owners, owner)

	this.Storage = append(this.Storage, file)
}

func (this *RamSP) AppendOwner(_entity int, _hash [32]byte, _meta *pdp.Metadata) {
	var owner Owner
	owner.ID  = _entity
	for i := range _meta.Tags {
		owner.Tags = append(owner.Tags, _meta.Tags[i].Bytes())
	}

	for _, v := range this.Storage {
		if v.Hash == _hash {
			v.Owners = append(v.Owners, owner)
			break
		}
	}
}

func (this *RamSP) Save(_path string) {
	tmp, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	defer f.Close()
	if err != nil { panic(err) }

	_, err = f.WriteString(string(tmp))
	if err != nil { panic(err) }
}