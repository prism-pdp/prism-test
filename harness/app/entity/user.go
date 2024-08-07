package entity

import (
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type User struct {
	Addr common.Address `json:'addr'`
	PublicKeyData pdp.PublicKeyData `json:'publicKey'`
	PrivateKeyData pdp.PrivateKeyData `json:'privateKey'`

	session *pdp.XZ21Session
}

func GenUser(_server string, _contractAddr string, _ethAddr common.Address, _ethKey string, _param *pdp.PairingParam) User {
	var user User

	user.Addr = _ethAddr

	pk, sk := pdp.GenPairingKey(_param)
	user.PublicKeyData = pk.Export()
	user.PrivateKeyData = sk.Export()

	user.session = helper.GenSession(_server, _contractAddr, _ethKey)

	return user
}

func LoadUser(_path string, _server string, _contractAddr string, _ethKey string) User {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := ioutil.ReadAll(f)
	if err != nil { panic(err) }

	var su User
	json.Unmarshal(s, &su)

	su.session = helper.GenSession(_server, _contractAddr, _ethKey)

	return su
}

func (this *User) Dump(_path string) {
	s, err := json.Marshal(this)
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)
	if err != nil { panic(err) }
}

func (this *User) IsUploaded(_data []byte) bool {
	hash := sha256.Sum256(_data)
	isUploaded, err := this.session.SearchFile(hash)
	if err != nil { panic(err) }

	return isUploaded
}

func (this *User) PrepareUpload(_data []byte, _chunkNum uint32) (pdp.Tags, uint32) {
	param := helper.FetchPairingParam(this.session)

	chunks, err := pdp.SplitData(_data, _chunkNum)
	if err != nil { panic(err) }

	sk := this.PrivateKeyData.Import(&param)
	tags, _, numTags := pdp.GenTags(&param, sk.Key, chunks)
	return tags, numTags
}

func (this *User) GenDedupProof(_chal *pdp.ChalData, _data []byte, _chunkNum uint32) pdp.ProofData {
	xz21Params, err := this.session.GetPara()
	params := pdp.GenParamFromXZ21Para(&xz21Params)

	chunks, err := pdp.SplitData(_data, _chunkNum)
	if err != nil { panic(err) }

	chal := _chal.Import(&params)
	proof := pdp.GenProof(&params, &chal, chunks)
	proofData := proof.Export()

	return proofData
}