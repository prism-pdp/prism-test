package entity

import (
	"fmt"
	"encoding/json"

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

	AuditingFileList [][32]byte

	client client.BaseClient
}


func GenUser(_name string, _addr string, _privKey string, _param *pdp.PairingParam, _simFlag bool) *User {
	u := new(User)

	u.Name = _name

	u.Addr = common.HexToAddress(_addr)
	u.PrivKey = _privKey

	pk, sk := pdp.GenPairingKey(_param)
	u.PublicKeyData = pk.Export()
	u.PrivateKeyData = sk.Export()

	if _simFlag {
		u.SetupSimClient(client.GetFakeLedger())
	} else {
		u.SetupEthClient()
	}

	return u
}

func (this *User) SetupSimClient(_ledger *client.FakeLedger) {
	this.client = client.NewSimClient(_ledger, this.Addr)
}

func (this *User) SetupEthClient() {
	this.client = client.NewEthClient(helper.Server, helper.ContractAddr, helper.SenderPrivKey, helper.SenderAddr)
}

func (this *User) PrepareUpload(_data []byte, _chunkNum uint32) pdp.TagSet {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)
	sk := this.PrivateKeyData.Import(param)

	helper.PrintLog("Start generate tags (chunks:%d)", int(_chunkNum))
	setChunk := pdp.GenChunkSet(_data, _chunkNum)
	tagSet, _ := pdp.GenTags(param, sk, setChunk)
	helper.PrintLog("Finish generate tags (chunks:%d)", int(_chunkNum))

	return tagSet
}

func (this *User) GenDedupProof(_chal *pdp.ChalData, _data []byte, _chunkNum uint32) pdp.ProofData {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	chal := _chal.Import(param)
	proof, _, _ := pdp.GenProof(param, chal, _chunkNum, _data)
	proofData := proof.Export()

	return proofData
}

func (this *User) GenAuditingChal(_hash [32]byte, _ratio float64) *pdp.ChalData {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	fileProp, err := this.client.SearchFile(_hash)
	if err != nil { panic(err) }
	if helper.IsEmptyFileProperty(&fileProp) { panic(fmt.Errorf("File property is not found")) }

	chal, err := pdp.NewChal(param, fileProp.SplitNum, _ratio)
	if err != nil { panic(err) }

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

func (this *User) AppendAuditingFile(_hash [32]byte) {
	this.AuditingFileList = append(this.AuditingFileList, _hash)
}

func (this *User) RemoveAuditingFile(_hash [32]byte) {
	this.AuditingFileList = append(this.AuditingFileList, _hash)
	newSlice := [][32]byte{}
    for _, v := range this.AuditingFileList {
        if v != _hash { // 条件に一致しない要素だけを新しいスライスに追加
            newSlice = append(newSlice, v)
        }
    }
	this.AuditingFileList = newSlice
}

func (this *User) GetAuditingFileList() [][32]byte {
	list := make([][32]byte, len(this.AuditingFileList))
	for i, v := range this.AuditingFileList {
		list[i] = v
	}
	return list
}

func (this *User) GetName() string {
	return this.Name
}

func (this *User) ToJson() (string, error) {
	b, err := json.MarshalIndent(this, "", "\t")
	return string(b), err
}

func (this *User) FromJson(_json []byte, _simFlag bool) {
	json.Unmarshal(_json, this)

	if _simFlag {
		this.SetupSimClient(client.GetFakeLedger())
	}
}

func (this *User) AfterLoad() {
	if *helper.SimFlag {
		this.SetupSimClient(client.GetFakeLedger())
	} else {
		this.SetupEthClient()
	}
}