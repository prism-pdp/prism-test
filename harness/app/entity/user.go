package entity

import (
	"fmt"
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/helper"
)

type User struct {
	Name string
	Addr common.Address `json:'addr'`
	PrivKey string
	PublicKeyData pdp.PublicKeyData `json:'publicKey'`
	PrivateKeyData pdp.PrivateKeyData `json:'privateKey'`

	client client.BaseClient
}


func GenUser(_addr string, _privKey string, _param *pdp.PairingParam, _name string) *User {
	user := new(User)

	user.Name = _name

	user.Addr = common.HexToAddress(_addr)
	user.PrivKey = _privKey

	pk, sk := pdp.GenPairingKey(_param)
	user.PublicKeyData = pk.Export()
	user.PrivateKeyData = sk.Export()

	return user
}

func LoadUser(_path string) *User {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := ioutil.ReadAll(f)
	if err != nil { panic(err) }

	su := new(User)
	json.Unmarshal(s, &su)

	return su
}

func (this *User) SetupSimClient(_ledger *client.FakeLedger) {
	this.client = client.NewSimClient(_ledger, this.Addr)
}

func (this *User) Dump(_path string) {
	s, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)
	if err != nil { panic(err) }
}

func (this *User) IsUploaded(_data []byte) bool {
	hash := sha256.Sum256(_data)
	fileProp, err := this.client.SearchFile(hash)
	if err != nil { panic(err) }

	if helper.IsEmptyFileProperty(&fileProp) { return false }
	return true
}

func (this *User) PrepareUpload(_data []byte, _chunkNum uint32) pdp.TagSet {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	chunks, err := pdp.SplitData(_data, _chunkNum)
	if err != nil { panic(err) }

	sk := this.PrivateKeyData.Import(&param)
	tagSet, _ := pdp.GenTags(&param, sk.Key, chunks)
	return tagSet
}

func (this *User) GenDedupProof(_chal *pdp.ChalData, _data []byte, _chunkNum uint32) pdp.ProofData {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Param(&xz21Param)

	chunks, err := pdp.SplitData(_data, _chunkNum)
	if err != nil { panic(err) }

	chal := _chal.Import(&params)
	proof := pdp.GenProof(&params, &chal, chunks)
	proofData := proof.Export()

	return proofData
}

func (this *User) GenAuditingChal(_hash [32]byte) pdp.ChalData {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Param(&xz21Param)

	fileProp, err := this.client.SearchFile(_hash)
	if err != nil { panic(err) }
	if helper.IsEmptyFileProperty(&fileProp) { panic(fmt.Errorf("File property is not found")) }

	chal := pdp.NewChal(&params, fileProp.SplitNum)
	chalData := chal.Export()

	return chalData
}

func (this *User) UploadAuditingChal(_hash [32]byte, _chalData *pdp.ChalData) {
	chalBytes, err := _chalData.Encode()
	if err != nil { panic(err) }
	err = this.client.SetChal(_hash, chalBytes)
}

func (this *User) GetFileList() [][32]byte {
	fileList, err := this.client.GetFileList(this.Addr)
	if err != nil { panic(err) }
	return fileList
}
