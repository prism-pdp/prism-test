package entity

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type File struct {
	Data []byte    `json:'data'`
	Owners []common.Address `json:'owners'`
}

type DedupState struct {
	AddrSU string `json:'addr'`
	Hash [32]byte `json:'hash'`
}

type Provider struct {
	Files map[string]*File `json:'files'`
	State map[uint32]*DedupState `json:'state'`

	ledger  *FakeLedger
	session *pdp.XZ21Session
}

func (this *Provider) searchFile(_hash [32]byte) *File {
	if v, ok := this.Files[helper.Hex(_hash[:])]; ok {
		return v
	}
	return nil
}

func GenProvider(_server string, _contractAddr string, _privKey string, _ledger *FakeLedger) Provider {
	var provider Provider

	provider.Files = make(map[string]*File)
	provider.State = make(map[uint32]*DedupState)

	provider.ledger = _ledger
	provider.session = helper.GenSession(_server, _contractAddr, _privKey)

	return provider
}

func LoadProvider(_path string, _server string, _contractAddr string, _privKey string, _ledger *FakeLedger) Provider {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := ioutil.ReadAll(f)
	if err != nil { panic(err) }

	var sp Provider
	json.Unmarshal(s, sp)

	if sp.Files == nil {
		sp.Files = make(map[string]*File)
	}
	if sp.State == nil {
		sp.State = make(map[uint32]*DedupState)
	}

	sp.ledger = _ledger
	sp.session = helper.GenSession(_server, _contractAddr, _privKey)

	return sp
}

func (this *Provider) Dump(_path string) {
	s, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)
	if err != nil { panic(err) }
}

func (this *Provider) NewFile(_addr common.Address, _hash [32]byte, _data []byte, _tags *pdp.Tags, _pubKey *pdp.PublicKeyData) {
	var file File
	file.Data = _data
	file.Owners = append(file.Owners, _addr)

	this.Files[helper.Hex(_hash[:])] = &file

	tagsData := _tags.Export()
	this.ledger.RegisterFile(_hash, &tagsData, _addr)
}

func (this *Provider) IsUploaded(_data []byte) bool {
	hash := sha256.Sum256(_data)
	fileProp := this.ledger.SearchFile(hash)
	if fileProp == nil { return false }

	return true
}

func (this *Provider) UploadNewFile(_data []byte, _tags *pdp.Tags, _addrSU common.Address, _pubKeySU *pdp.PublicKeyData) error {
	hash := sha256.Sum256(_data)

	isUploaded := this.IsUploaded(_data)

	if isUploaded {
		return fmt.Errorf("File is already uploaded. (hash:%s)", helper.Hex(hash[:]))
	} else {
		this.NewFile(_addrSU, hash, _data, _tags, _pubKeySU)
		this.session.RegisterFile(hash, _addrSU)
	}

	return nil
}

func (this *Provider) AppendOwner(_su *User, _data []byte) {
	hash := sha256.Sum256(_data)
	file := this.searchFile(hash)

	file.Owners = append(file.Owners, _su.Addr)
	this.ledger.AppendAccount(hash, _su.Addr)
}

func (this *Provider) GenDedupChallen(_data []byte, _addrSU common.Address) (pdp.ChalData, uint32) {
	params := helper.FetchPairingParam(this.session)

	hash := sha256.Sum256(_data)
	// file := this.searchFile(hash)
	// if file == nil { panic(fmt.Errorf("File is not found.")) }
	numTags := this.ledger.GetNumTags(hash)

	chal := pdp.GenChal(&params, numTags)
	chalData := chal.Export()

	//
	var s DedupState
	s.AddrSU = _addrSU.Hex()
	s.Hash = hash

	id := rand.Uint32()
	this.State[id] = &s

	return chalData, id
}

func (this *Provider) VerifyDedupProof(_id uint32, _chalData *pdp.ChalData, _proofData *pdp.ProofData) bool {
	params := helper.FetchPairingParam(this.session)

	state := this.State[_id]

	fileProp := this.ledger.SearchFile(state.Hash)
	if fileProp == nil { panic(fmt.Errorf("File property is not found.")) }

	file := this.searchFile(state.Hash)
	if file == nil { panic(fmt.Errorf("File is not found.")) }

	pkData := this.ledger.SearchPublicKey(fileProp.Creator)
	if pkData == nil { panic(fmt.Errorf("Account is not found.")) }

	// TODO: function VerifyProof内で必要なタグだけ復元するのがよい
	tags := fileProp.Tags.Import(&params)

	chunks, err := pdp.SplitData(file.Data, fileProp.GetNumTags())
	if err != nil { panic(err) }

	hashChunks := pdp.HashChunks(chunks)

	chal := _chalData.Import(&params)
	proof := _proofData.Import(&params)
	pk := pkData.Import(&params)

	isVerified := pdp.VerifyProof(&params, &tags, hashChunks, &chal, &proof, pk.Key)

	return isVerified
}