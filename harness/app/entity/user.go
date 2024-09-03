package entity

import (
	"fmt"
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
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
	fileProp, err := this.session.SearchFile(hash)
	if err != nil { panic(err) }

	// if len(fileProp.Creator.Bytes()) == 0 { return false } // TODO: make function
	if helper.IsEmptyFileProperty(&fileProp) { return false }
	return true
}

func (this *User) PrepareUpload(_data []byte, _chunkNum uint32) pdp.Tag {
	xz21Param, err := this.session.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	chunks, err := pdp.SplitData(_data, _chunkNum)
	if err != nil { panic(err) }

	sk := this.PrivateKeyData.Import(&param)
	tag, _ := pdp.GenTag(&param, sk.Key, chunks)
	return tag
}

func (this *User) GenDedupProof(_chal *pdp.ChalData, _data []byte, _chunkNum uint32) pdp.ProofData {
	xz21Param, err := this.session.GetParam()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Param(&xz21Param)

	chunks, err := pdp.SplitData(_data, _chunkNum)
	if err != nil { panic(err) }

	chal := _chal.Import(&params)
	proof := pdp.GenProof(&params, &chal, chunks)
	proofData := proof.Export()

	return proofData
}

func (this *User) GenAuditChallen(_hash [32]byte) pdp.ChalData {
	xz21Param, err := this.session.GetParam()
	if err != nil { panic(err) }

	params := pdp.GenParamFromXZ21Param(&xz21Param)

	fileProp, err := this.session.SearchFile(_hash)
	if err != nil { panic(err) }
	if len(fileProp.Creator.Bytes()) == 0 { panic(fmt.Errorf("File property is not found")) }

	chal := pdp.GenChal(&params, fileProp.SplitNum)
	chalData := chal.Export()

	return chalData
}

// Return true when upload is success.
// Return false when the file is under auditing.
func (this *User) UploadChallen(_hash [32]byte, _chalData *pdp.ChalData) bool {
	chalBytes, err := _chalData.Encode()
	if err != nil { panic(err) }
	success, err := this.session.SetChal(_hash, chalBytes)
	if err != nil { panic(err) }

	return success
}

func (this *User) FetchFileList() [][32]byte {
	fileList, err := this.session.GetFileList(this.Addr)
	if err != nil { panic(err) }
	return fileList
}