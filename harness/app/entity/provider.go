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
	"github.com/dpduado/dpduado-test/harness/session"
)

type File struct {
	Data []byte    `json:'data'`
	TagData pdp.TagData `json:'tag'`
	Owners []common.Address `json:'owners'`
}

type DedupState struct {
	AddrSU string `json:'addr'`
	Hash [32]byte `json:'hash'`
}

type Provider struct {
	Files map[string]*File `json:'files'`
	State map[uint32]*DedupState `json:'state'`

	session session.Session
}

func (this *Provider) SearchFile(_hash [32]byte) *File {
	if v, ok := this.Files[helper.Hex(_hash[:])]; ok {
		return v
	}
	return nil
}

func GenProvider(_server string, _contractAddr string, _privKey string, _session session.Session) *Provider {
	provider := new(Provider)

	provider.Files = make(map[string]*File)
	provider.State = make(map[uint32]*DedupState)

	provider.session = _session

	return provider
}

func LoadProvider(_path string, _server string, _contractAddr string, _privKey string, _session session.Session) *Provider {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := ioutil.ReadAll(f)
	if err != nil { panic(err) }

	sp := new(Provider)
	json.Unmarshal(s, sp)

	if sp.Files == nil {
		sp.Files = make(map[string]*File)
	}
	if sp.State == nil {
		sp.State = make(map[uint32]*DedupState)
	}

	sp.session = _session

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

func (this *Provider) NewFile(_addr common.Address, _hash [32]byte, _data []byte, _tag *pdp.Tag, _pubKey *pdp.PublicKeyData) {
	var file File
	file.Data = _data
	file.TagData = _tag.Export()
	file.Owners = append(file.Owners, _addr)

	this.Files[helper.Hex(_hash[:])] = &file

	this.session.RegisterFile(_hash, _tag.Size, _addr)
}

func (this *Provider) IsUploaded(_data []byte) bool {
	hash := sha256.Sum256(_data)
	fileProp, err := this.session.SearchFile(hash)
	if err != nil { panic(err) }
	if len(fileProp.Creator.Bytes()) == 0 { return false }

	return true
}

func (this *Provider) UploadNewFile(_data []byte, _tag *pdp.Tag, _addrSU common.Address, _pubKeySU *pdp.PublicKeyData) error {
	hash := sha256.Sum256(_data)

	isUploaded := this.IsUploaded(_data)

	if isUploaded {
		return fmt.Errorf("File is already uploaded. (hash:%s)", helper.Hex(hash[:]))
	} else {
		this.NewFile(_addrSU, hash, _data, _tag, _pubKeySU)
	}

	return nil
}

func (this *Provider) AppendOwner(_su *User, _data []byte) {
	hash := sha256.Sum256(_data)
	file := this.SearchFile(hash)

	file.Owners = append(file.Owners, _su.Addr)
	err := this.session.AppendOwner(hash, _su.Addr)
	if err != nil { panic(err) }
}

func (this *Provider) GetTagSize(_hash [32]byte) uint32 {
	file := this.SearchFile(_hash)
	if file == nil { panic(fmt.Errorf("File is not found.")) }

	return file.TagData.Size
}

func (this *Provider) GenDedupChallen(_data []byte, _addrSU common.Address) (pdp.ChalData, uint32) {
	xz21Param, err := this.session.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	hash := sha256.Sum256(_data)
	// file := this.searchFile(hash)
	// if file == nil { panic(fmt.Errorf("File is not found.")) }
	tagSize := this.GetTagSize(hash)

	chal := pdp.GenChal(&param, tagSize)
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
	xz21Param, err := this.session.GetParam()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Param(&xz21Param)

	state := this.State[_id]

	fileProp, err := this.session.SearchFile(state.Hash) // TODO: SearchFile -> SearchFileProperty
	if err != nil { panic(err) }
	if len(fileProp.Creator.Bytes()) == 0 { panic(fmt.Errorf("File property is not found.")) }

	account, err := this.session.GetAccount(fileProp.Creator)
	if err != nil { panic(fmt.Errorf("Account is not found.")) }
	pkData := pdp.PublicKeyData{account.PubKey}

	file := this.SearchFile(state.Hash)
	if file == nil { panic(fmt.Errorf("File is not found.")) }

	// TODO: function VerifyProof内で必要なタグだけ復元するのがよい
	tag := file.TagData.Import(&params)

	chunks, err := pdp.SplitData(file.Data, tag.Size)
	if err != nil { panic(err) }

	hashChunks := pdp.HashChunks(chunks)

	chal := _chalData.Import(&params)
	proof := _proofData.Import(&params)
	pk := pkData.Import(&params)

	isVerified := pdp.VerifyProof(&params, &tag, hashChunks, &chal, &proof, pk.Key)

	return isVerified
}

func (this *Provider) DownloadChallen() ([][32]byte, []pdp.ChalData) {
	hashList, chalDataList, err := this.session.GetChalList()
	if err != nil { panic(err) }
	return hashList, chalDataList
}

func (this *Provider) GenAuditProof(_hash [32]byte, _chal *pdp.ChalData) pdp.ProofData {
	xz21Param, err := this.session.GetParam()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Param(&xz21Param)

	f := this.SearchFile(_hash)
	if f == nil { panic(fmt.Errorf("Unknown file: %s", helper.Hex(_hash[:]))) }

	chunks, err := pdp.SplitData(f.Data, f.TagData.Size)
	if err != nil { panic(err) }

	chal := _chal.Import(&params)
	proof := pdp.GenProof(&params, &chal, chunks)
	proofData := proof.Export()

	return proofData
}

func (this *Provider) UploadProof(_hash [32]byte, _proofData *pdp.ProofData) {
	proofBytes, err := _proofData.Encode()
	if err != nil { panic(err) }
	err = this.session.SetProof(_hash, proofBytes)
	if err != nil { panic(err) }
}