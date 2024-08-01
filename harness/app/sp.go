package main

import (
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/Nik-U/pbc"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

type Property struct {
	Size uint32   `json:'size'`
	Tags [][]byte `json:'meta'`
	Key  []byte   `json:'key'`
}

type Owner struct {
	Addr string   `json:'addr'`
}

type File struct {
	Hash [32]byte  `json:'hash'`
	Data []byte    `json:'data'`
	Prop *Property  `json:'property'`
	Owners []*Owner `json:'owners'`
}

type RamSP struct {
	Storage []*File `json:'storage'`
}

func (this *RamSP) searchFile(_hash [32]byte) *File {
	for _, v := range this.Storage {
		if v.Hash == _hash {
			return v
		}
	}
	return nil
}

func GenRamSP() *RamSP {
	ram := new(RamSP)
	return ram
}

func (this *RamSP) NewFile(_addr string, _hash [32]byte, _data []byte, _meta *pdp.Metadata, _pubKey *pbc.Element) {
	var owner Owner
	owner.Addr = _addr

	var prop Property
	prop.Size = uint32(len(_meta.Tags))
	prop.Key = _pubKey.Bytes()
	for i := range _meta.Tags {
		prop.Tags = append(prop.Tags, _meta.Tags[i].Bytes())
	}

	var file File
	file.Hash = _hash
	file.Data = _data
	file.Prop = &prop
	file.Owners = append(file.Owners, &owner)

	this.Storage = append(this.Storage, &file)
}

// func (this *RamSP) appendOwner(_ctx *Context, _ram *RamSP, _addr string, _hash [32]byte, _meta *pdp.Metadata) {
// 	var owner Owner
// 	owner.Addr  = _addr
// 	for i := range _meta.Tags {
// 		owner.Tags = append(owner.Tags, _meta.Tags[i].Bytes())
// 	}

// 	for _, v := range this.Storage {
// 		if v.Hash == _hash {
// 			v.Owners = append(v.Owners, owner)
// 			break
// 		}
// 	}
// }

func (this *RamSP) Save(_path string) {
	tmp, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	defer f.Close()
	if err != nil { panic(err) }

	_, err = f.WriteString(string(tmp))
	if err != nil { panic(err) }
}

func acceptNewFile(_ctx *Context, _data []byte, _meta *pdp.Metadata, _addrSU common.Address, _pubKeySU *pbc.Element, _ram *RamSP) error {
	hash := sha256.Sum256(_data)

	isFound, err := _ctx.Session.SearchFile(hash)
	if err != nil { panic(err) }
	if isFound {
		return fmt.Errorf("File is already stored. (hash:%s)", hex.EncodeToString(hash[:]))
	} else {
		_ram.NewFile(_addrSU.Hex(), hash, data, _meta, _pubKeySU)
		_ctx.Session.RegisterFile(hash, _addrSU)
	}

	return nil
}

func genDedupChallen(_ctx *Context, _data []byte, _ram *RamSP) *pdp.Chal {
	xz21Para, err := _ctx.Session.GetPara()
	if err != nil { panic(err) }

	para := pdp.GenParamFromXZ21Para(&xz21Para)

	hash := sha256.Sum256(data)
	file := _ram.searchFile(hash)
	if file == nil { panic(fmt.Errorf("File is not found.")) }

	chal := pdp.GenChal(para, file.Prop.Size)
	return chal
}

func verifyDedupProof(_ctx *Context, _ram *RamSP, _chal *pdp.Chal, _proof *pbc.Element) {
	xz21Para, err := _ctx.Session.GetPara()
	if err != nil { panic(err) }

	para := pdp.GenParamFromXZ21Para(&xz21Para)

	pdp.VerifyProof(para, )
}