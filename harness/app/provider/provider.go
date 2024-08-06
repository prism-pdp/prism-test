package provider

import (
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
	"github.com/dpduado/dpduado-test/harness/user"
)

type Property struct {
	TagsData pdp.TagsData `json:'tags'`
	KeyData pdp.PublicKeyData `json:'key'`
}

func (this *Property) numTags() uint32 {
	return uint32(len(this.TagsData.Tags))
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

type DedupState struct {
	AddrSU string `json:'addr'`
	Hash [32]byte `json:'hash'`
}

type Storage struct {
	Files []*File `json:'files'`
}

type Provider struct {
	Storage Storage `json:'storage'`
	State map[uint32]DedupState `json:'state'`

	session *pdp.XZ21Session
}

func (this *Provider) searchFile(_hash [32]byte) *File {
	for _, v := range this.Storage.Files {
		if v.Hash == _hash {
			return v
		}
	}
	return nil
}

func GenProvider(_server string, _contractAddr string, _privKey string) Provider {
	var provider Provider

	provider.State = make(map[uint32]DedupState, 1)
	provider.session = helper.GenSession(_server, _contractAddr, _privKey)

	return provider
}

func LoadProvider(_path string, _server string, _contractAddr string, _privKey string) Provider {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := ioutil.ReadAll(f)
	if err != nil { panic(err) }

	var sp Provider
	json.Unmarshal(s, sp)

	if sp.State == nil {
		sp.State = make(map[uint32]DedupState )
	}

	sp.session = helper.GenSession(_server, _contractAddr, _privKey)

	return sp
}

func (this *Provider) Dump(_path string) {
	s, err := json.Marshal(this)
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)
	if err != nil { panic(err) }
}

func (this *Provider) NewFile(_addr string, _hash [32]byte, _data []byte, _tags *pdp.Tags, _pubKey *pdp.PublicKeyData) {
	var owner Owner
	owner.Addr = _addr

	var prop Property
	prop.TagsData = _tags.Export()
	prop.KeyData = *_pubKey

	var file File
	file.Hash = _hash
	file.Data = _data
	file.Prop = &prop
	file.Owners = append(file.Owners, &owner)

	this.Storage.Files = append(this.Storage.Files, &file)
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

func (this *Provider) SaveStorage(_path string) {
	tmp, err := json.MarshalIndent(this.Storage, "", "\t")
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	defer f.Close()
	if err != nil { panic(err) }

	_, err = f.WriteString(string(tmp))
	if err != nil { panic(err) }
}

func (this *Provider) IsUploaded(_data []byte) bool {
	hash := sha256.Sum256(_data)
	isUploaded, err := this.session.SearchFile(hash)
	if err != nil { panic(err) }

	return isUploaded
}

func (this *Provider) UploadNewFile(_data []byte, _tags *pdp.Tags, _addrSU common.Address, _pubKeySU *pdp.PublicKeyData) error {
	hash := sha256.Sum256(_data)

	isUploaded := this.IsUploaded(_data)

	if isUploaded {
		return fmt.Errorf("File is already uploaded. (hash:%s)", hex.EncodeToString(hash[:]))
	} else {
		this.NewFile(_addrSU.Hex(), hash, _data, _tags, _pubKeySU)
		this.session.RegisterFile(hash, _addrSU)
	}

	return nil
}

func (this *Provider) AppendOwner(_su *user.User, _data []byte) {
	hash := sha256.Sum256(_data)
	file := this.searchFile(hash)

	var owner Owner
	owner.Addr = _su.Addr.Hex()

	file.Owners = append(file.Owners, &owner)
}

func (this *Provider) GenDedupChallen(_data []byte, _addrSU common.Address) (pdp.ChalData, uint32) {
	params := helper.FetchPairingParam(this.session)

	hash := sha256.Sum256(_data)
	file := this.searchFile(hash)
	if file == nil { panic(fmt.Errorf("File is not found.")) }

	chal := pdp.GenChal(&params, file.Prop.numTags())
	chalData := chal.Export()

	//
	var s DedupState
	s.AddrSU = _addrSU.Hex()
	s.Hash = hash

	id := rand.Uint32()
	this.State[id] = s

	return chalData, id
}

func (this *Provider) VerifyDedupProof(_id uint32, _chalData *pdp.ChalData, _proofData *pdp.ProofData) bool {
	params := helper.FetchPairingParam(this.session)

	state := this.State[_id]
	file := this.searchFile(state.Hash)
	if file == nil { panic(fmt.Errorf("File is not found.")) }

	pk := file.Prop.KeyData.Import(&params)

	// TODO: function VerifyProof内で必要なタグだけ復元するのがよい
	tags := file.Prop.TagsData.Import(&params)

	chunks, err := pdp.SplitData(file.Data, file.Prop.numTags())
	if err != nil { panic(err) }

	hashChunks := pdp.HashChunks(chunks)

	chal := _chalData.Import(&params)
	proof := _proofData.Import(&params)
	isVerified := pdp.VerifyProof(&params, &tags, hashChunks, &chal, &proof, pk.Key)

	return isVerified
}