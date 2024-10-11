package entity

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/helper"
)

type File struct {
	Data []byte    `json:'data'`
	TagData pdp.TagData `json:'tag'`
	Owners []common.Address `json:'owners'`
}

type Provider struct {
	Name string
	Files map[string]*File `json:'files'`

	client client.BaseClient
}

func (this *Provider) SearchFile(_hash [32]byte) *File {
	if v, ok := this.Files[helper.Hex(_hash[:])]; ok {
		return v
	}
	return nil
}

func MakeProvider(_path string, _client client.BaseClient, _name string) *Provider {
	if (helper.IsFile(_path)) {
		return LoadProvider(_path, _client)
	} else {
		return GenProvider(_client, _name)
	}
}

func GenProvider(_client client.BaseClient, _name string) *Provider {
	provider := new(Provider)

	provider.Name = _name
	provider.Files = make(map[string]*File)

	provider.client = _client

	return provider
}

func LoadProvider(_path string, _client client.BaseClient) *Provider {
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

	sp.client = _client

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

	this.client.RegisterFile(_hash, _tag.Size, _addr)
}

func (this *Provider) IsUploaded(_data []byte) bool {
	hash := sha256.Sum256(_data)
	fileProp, err := this.client.SearchFile(hash)
	if err != nil { panic(err) }
	if helper.IsEmptyFileProperty(&fileProp) { return false }

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

func (this *Provider) RegisterOwnerToFile(_su *User, _data []byte, _chalData *pdp.ChalData, _proofData *pdp.ProofData) bool {
	// check file
	hash := sha256.Sum256(_data)
	file := this.SearchFile(hash)
	if file == nil { panic(fmt.Errorf("File is not found.")) }
	fileProp, err := this.client.SearchFile(hash)
	if err != nil { panic(err) }
	if helper.IsEmptyFileProperty(&fileProp) { panic(fmt.Errorf("File property is not found."))}
	// prepare params
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }
	params := pdp.GenParamFromXZ21Param(&xz21Param)

	// ===================================
	// Verify chal & proof
	// ===================================
	// prepare public key of the creator of the file
	account, err := this.client.GetAccount(fileProp.Creator)
	if err != nil { panic(fmt.Errorf("Account is not found.")) }
	pkData := pdp.PublicKeyData{account.PubKey}
	pk := pkData.Import(&params)
	// prepare chal, proof
	chal := _chalData.Import(&params)
	proof := _proofData.Import(&params)
	// prepare hash chunks
	// TODO: function VerifyProof内で必要なタグだけ復元するのがよい
	tag := file.TagData.ImportSubset(&params, &chal)
	chunks, err := pdp.SplitData(file.Data, tag.Size)
	if err != nil { panic(err) }
	digestSubset := pdp.HashSampledChunks(chunks, &chal)
	// verify chal & proof
	isVerified, err := pdp.VerifyProof(&params, &tag, digestSubset, &chal, &proof, pk.Key)
	if err != nil { panic(err) }
	if !isVerified { return false }

	// ===================================
	// Verify chal & proof
	// ===================================
	file.Owners = append(file.Owners, _su.Addr)
	err = this.client.AppendOwner(hash, _su.Addr)
	if err != nil { panic(err) }

	return true
}

func (this *Provider) GenDedupChal(_data []byte, _addrSU common.Address) (pdp.ChalData) {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	hash := sha256.Sum256(_data)
	file := this.SearchFile(hash)
	if file == nil { panic(fmt.Errorf("File is not found.")) }

	chal := pdp.NewChal(&param, file.TagData.Size)
	chalData := chal.Export()

	return chalData
}

func (this *Provider) DownloadAuditingChal() ([][32]byte, []pdp.ChalData) {
	hashList, reqList, err := this.client.GetAuditingReqList()
	if err != nil { panic(err) }
	chalDataList := make([]pdp.ChalData, 0)
	for _, v := range reqList {
		if len(v.Proof) == 0 {
			chalData, err := pdp.DecodeToChalData(v.Chal)
			if err != nil { panic(err) }
			chalDataList = append(chalDataList, chalData)
		}
	}
	return hashList, chalDataList
}

func (this *Provider) GenAuditingProof(_hash [32]byte, _chal *pdp.ChalData) pdp.ProofData {
	xz21Param, err := this.client.GetParam()
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

func (this *Provider) UploadAuditingProof(_hash [32]byte, _proofData *pdp.ProofData) {
	proofBytes, err := _proofData.Encode()
	if err != nil { panic(err) }
	err = this.client.SetProof(_hash, proofBytes)
	if err != nil { panic(err) }
}

func (this *Provider) PrepareVerificationData(_hash [32]byte, _chalData *pdp.ChalData) (common.Address, *pdp.DigestSet, *pdp.TagData) {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Param(&xz21Param)
	chal := _chalData.Import(&params)

	file := this.SearchFile(_hash)
	digestSubset, tagDataSubset, err := pdp.MakeSubset(file.Data, &file.TagData, &chal)
	if err != nil { panic(err) }

	return file.Owners[0], digestSubset, tagDataSubset
}