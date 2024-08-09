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

	ledger *FakeLedger
	session *pdp.XZ21Session
}

func GenUser(_server string, _contractAddr string, _ethAddr common.Address, _ethKey string, _param *pdp.PairingParam, _ledger *FakeLedger) User {
	var user User

	user.Addr = _ethAddr

	pk, sk := pdp.GenPairingKey(_param)
	user.PublicKeyData = pk.Export()
	user.PrivateKeyData = sk.Export()

	user.ledger = _ledger
	user.session = helper.GenSession(_server, _contractAddr, _ethKey)

	return user
}

func LoadUser(_path string, _server string, _contractAddr string, _ethKey string, _ledger *FakeLedger) User {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := ioutil.ReadAll(f)
	if err != nil { panic(err) }

	var su User
	json.Unmarshal(s, &su)

	su.ledger = _ledger
	su.session = helper.GenSession(_server, _contractAddr, _ethKey)

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
	fileProp := this.ledger.SearchFile(hash)
	if fileProp == nil { return false }

	return true
}

func (this *User) PrepareUpload(_data []byte, _chunkNum uint32) pdp.Tag {
	param := helper.FetchPairingParam(this.session)

	chunks, err := pdp.SplitData(_data, _chunkNum)
	if err != nil { panic(err) }

	sk := this.PrivateKeyData.Import(&param)
	tag, _ := pdp.GenTag(&param, sk.Key, chunks)
	return tag
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

// func (this *Provider) GenAuditChallen(_data []byte) (pdp.ChalData, uint32) {
// 	params := helper.FetchPairingParam(this.session)

// 	hash := sha256.Sum256(_data)
// 	// file := this.searchFile(hash)
// 	fileProp := this.ledger.SearchFile(hash)
// 	if fileProp == nil { panic(fmt.Errorf("File property is not found.")) }

// 	chal := pdp.GenChal(&params, fileProp.GetNumTags())
// 	chalData := chal.Export()

// 	return chalData
// }