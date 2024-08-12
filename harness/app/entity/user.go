package entity

import (
	"fmt"
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/session"
)

type User struct {
	Addr common.Address `json:'addr'`
	PublicKeyData pdp.PublicKeyData `json:'publicKey'`
	PrivateKeyData pdp.PrivateKeyData `json:'privateKey'`

	session session.Session
}

func GenUser(_server string, _contractAddr string, _ethAddr common.Address, _ethKey string, _param *pdp.PairingParam, _session session.Session) *User {
	user := new(User)

	user.Addr = _ethAddr

	pk, sk := pdp.GenPairingKey(_param)
	user.PublicKeyData = pk.Export()
	user.PrivateKeyData = sk.Export()

	user.session = _session

	return user
}

func LoadUser(_path string, _server string, _contractAddr string, _ethKey string, _session session.Session) *User {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := ioutil.ReadAll(f)
	if err != nil { panic(err) }

	su := new(User)
	json.Unmarshal(s, &su)

	su.session = _session

	return su
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
	found := this.session.SearchFile(hash)
	return !(found == nil)
}

func (this *User) PrepareUpload(_data []byte, _chunkNum uint32) pdp.Tag {
	xz21Para, err := this.session.GetPara()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Para(&xz21Para)

	chunks, err := pdp.SplitData(_data, _chunkNum)
	if err != nil { panic(err) }

	sk := this.PrivateKeyData.Import(&params)
	tag, _ := pdp.GenTag(&params, sk.Key, chunks)
	return tag
}

func (this *User) GenDedupProof(_chal *pdp.ChalData, _data []byte, _chunkNum uint32) pdp.ProofData {
	xz21Params, err := this.session.GetPara()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Para(&xz21Params)

	chunks, err := pdp.SplitData(_data, _chunkNum)
	if err != nil { panic(err) }

	chal := _chal.Import(&params)
	proof := pdp.GenProof(&params, &chal, chunks)
	proofData := proof.Export()

	return proofData
}

func (this *User) GenAuditChallen(_hash [32]byte) pdp.ChalData {
	xz21Params, err := this.session.GetPara()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Para(&xz21Params)

	fileProp := this.session.SearchFile(_hash)
	if fileProp == nil { panic(fmt.Errorf("File property is not found.")) }

	chal := pdp.GenChal(&params, fileProp.SplitNum)
	chalData := chal.Export()

	return chalData
}

func (this *User) UploadChallen(_hash [32]byte, _chalData *pdp.ChalData) {
	chalBytes, err := _chalData.Encode()
	if err != nil { panic(err) }
	this.session.UploadChallen(_hash, chalBytes)
}

func (this *User) FetchFileList() [][32]byte {
	return this.session.FetchFileList()
}