package entity

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/helper"
)

type File struct {
	Filename string `json:'filename'`
	TagFilename string `json:'tagfilename'`
	Owners []common.Address `json:'owners'`
}

type Provider struct {
	Name string
	Files map[string]*File `json:'files'`
	Addr common.Address
	PrivKey string

	client client.BaseClient
}

func (this *Provider) SearchFile(_hash [32]byte) *File {
	if v, ok := this.Files[helper.Hex(_hash[:])]; ok {
		return v
	}
	return nil
}

func GenProvider(_name string, _addr string, _privKey string, _simFlag bool) *Provider {
	p := new(Provider)

	p.Name = _name
	p.Files = make(map[string]*File)
	p.Addr = common.HexToAddress(_addr)
	p.PrivKey = _privKey

	if _simFlag {
		p.SetupSimClient(client.GetFakeLedger())
	} else {
		p.SetupEthClient()
	}

	return p
}

func (this *Provider) SetupSimClient(_ledger *client.FakeLedger) {
	this.client = client.NewSimClient(_ledger, this.Addr)
}

func (this *Provider) SetupEthClient() {
	this.client = client.NewEthClient(helper.Server, helper.ContractAddr, helper.SenderPrivKey, helper.SenderAddr)
}

func (this *Provider) NewFile(_addr common.Address, _hash [32]byte, _data []byte, _tagSet *pdp.TagSet, _pubKey *pdp.PublicKeyData) error {
	hex := helper.Hex(_hash[:])

	var file File
	file.Filename = fmt.Sprintf("%s.dat", hex)
	file.TagFilename = file.Filename + ".tag"
	file.Owners = append(file.Owners, _addr)
	this.Files[hex] = &file

	// Save file
	pathFile := helper.MakeDumpFilePath(this.Name, file.Filename)
	helper.WriteFile(pathFile, _data)

	// Save tag file
	setTagData := _tagSet.Export()
	b, err := json.MarshalIndent(setTagData, "", "\t")
	if err != nil { return err }
	pathTagFile := helper.MakeDumpFilePath(this.Name, file.TagFilename)
	helper.WriteFile(pathTagFile, b)

	err = this.client.RegisterFile(_hash, _tagSet.Size(), _addr)
	if err != nil { return err }

	return nil
}

func (this *Provider) IsUploaded(_data []byte) (bool, uint32) {
	hash := sha256.Sum256(_data)
	fileProp, err := this.client.SearchFile(hash)
	if err != nil { panic(err) }
	if helper.IsEmptyFileProperty(&fileProp) { return false, 0 }

	return true, fileProp.SplitNum
}

func (this *Provider) UploadNewFile(_data []byte, _tagSet *pdp.TagSet, _addrSU common.Address, _pubKeySU *pdp.PublicKeyData) error {
	hash := sha256.Sum256(_data)

	isUploaded, _ := this.IsUploaded(_data)

	if isUploaded {
		return fmt.Errorf("File is already uploaded. (hash:%s)", helper.Hex(hash[:]))
	} else {
		err := this.NewFile(_addrSU, hash, _data, _tagSet, _pubKeySU)
		if err != nil { return err }
	}

	return nil
}

func (this *Provider) RegisterOwnerToFile(_su *User, _data []byte, _chalData *pdp.ChalData, _proofData pdp.ProofData) (bool, error) {
	// check file
	hash := sha256.Sum256(_data)
	file := this.SearchFile(hash)
	if file == nil { return false, fmt.Errorf("File is not found.") }
	fileProp, err := this.client.SearchFile(hash)
	if err != nil { return false, err }
	if helper.IsEmptyFileProperty(&fileProp) { return false, fmt.Errorf("File property is not found.")}
	// prepare param
	xz21Param, err := this.client.GetParam()
	if err != nil { return false, err }
	param := pdp.GenParamFromXZ21Param(&xz21Param)

	// ===================================
	// Verify chal & proof
	// ===================================
	// prepare public key of the creator of the file
	account, err := this.client.GetAccount(fileProp.Creator)
	if err != nil { return false, fmt.Errorf("Account is not found.") }
	pkData := (pdp.PublicKeyData)(account.PubKey)
	pk := pkData.Import(param)
	// prepare chal, proof
	chal := _chalData.Import(param)
	proof := _proofData.Import(param)
	// prepare hash chunks
	// TODO: function VerifyProof内で必要なタグだけ復元するのがよい
	setTagData, err := this.ReadTagFile(file)
	if err != nil { return false, err }
	tagSet := setTagData.ImportSubset(param, fileProp.SplitNum, chal)
	data, err := this.ReadFile(file)
	if err != nil { return false, err }
	subsetChunk := pdp.GenChunkSubset(data, fileProp.SplitNum, chal)
	subsetDigest := subsetChunk.Hash()
	// verify chal & proof
	isVerified, err := pdp.VerifyProof(param, fileProp.SplitNum, tagSet, subsetDigest, chal, proof, pk)
	if err != nil { return false, err }
	if !isVerified { return false, nil }

	// ===================================
	// Verify chal & proof
	// ===================================
	file.Owners = append(file.Owners, _su.Addr)
	err = this.client.AppendOwner(hash, _su.Addr)
	if err != nil { return false, err }

	return true, nil
}

func (this *Provider) GenDedupChal(_data []byte, _addrSU common.Address) *pdp.ChalData {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	hash := sha256.Sum256(_data)

	file := this.SearchFile(hash)
	if file == nil { panic(fmt.Errorf("File is not found.")) }

	fileProp, err := this.client.SearchFile(hash)

	chal, err := pdp.NewChal(param, fileProp.SplitNum, 1.0)
	if err != nil { panic(err) }

	chalData := chal.Export()

	return chalData
}

func (this *Provider) DownloadAuditingChal(_fileList [][32]byte) ([]*pdp.ChalData) {
	chalDataList := make([]*pdp.ChalData, len(_fileList))

	for i, v := range _fileList {
		req, err := this.client.GetAuditingReq(v)
		if err != nil { panic(err) }

		if len(req.Proof) == 0 {
			chalData, err := pdp.DecodeToChalData(req.Chal)
			if err != nil { panic(err) }
			chalDataList[i] = chalData
		}
	}
	return chalDataList
}

func (this *Provider) GenAuditingProof(_hash [32]byte, _chal *pdp.ChalData) pdp.ProofData {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)

	file := this.SearchFile(_hash)
	if file == nil { panic(fmt.Errorf("Unknown file: %s", helper.Hex(_hash[:]))) }

	fileProp, err := this.client.SearchFile(_hash)
	if err != nil { panic(err) }

	chal := _chal.Import(param)
	if chal == nil { panic(fmt.Errorf("Invalid chal")) }

	data, err := this.ReadFile(file)
	if err != nil { panic(err) }

	target := chal.GetTargetBlockCount()
	helper.PrintLog("Start generating proof (chunks:%d, target:%d)", fileProp.SplitNum, target)
	proof, _, _ := pdp.GenProof(param, chal, fileProp.SplitNum, data)
	helper.PrintLog("Finish generating proof (chunks:%d, target:%d)", fileProp.SplitNum, target)

	proofData := proof.Export()

	return proofData
}

func (this *Provider) UploadAuditingProof(_hash [32]byte, _proofData pdp.ProofData) {
	proofBytes := _proofData.Base()
	err := this.client.SetProof(_hash, proofBytes)
	if err != nil { panic(err) }
}

func (this *Provider) PrepareVerificationData(_hash [32]byte, _chalData *pdp.ChalData) (common.Address, pdp.DigestSet, pdp.TagDataSet) {
	xz21Param, err := this.client.GetParam()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Param(&xz21Param)
	chal := _chalData.Import(param)

	file := this.SearchFile(_hash)
	if file == nil { panic(fmt.Errorf("Unknown file (hash:%s)", helper.Hex(_hash[:])))}

	fileProp, err := this.client.SearchFile(_hash)
	if err != nil { panic(err) }

	data, err := this.ReadFile(file)
	if err != nil { panic(err) }
	subsetDigest := pdp.GenChunkSubset(data, fileProp.SplitNum, chal).Hash()

	setTagData, err := this.ReadTagFile(file)
	if err != nil { panic(err) }
	subsetTagData := setTagData.DuplicateSubset(fileProp.SplitNum, chal)

	return file.Owners[0], subsetDigest, subsetTagData
}

func (this *Provider) GetName() string {
	return this.Name
}

func (this *Provider) GetFilePath(_f *File) string {
	return helper.MakeDumpFilePath(this.Name, _f.Filename)
}

func (this *Provider) ToJson() (string, error) {
	b, err := json.MarshalIndent(this, "", "\t")
	return string(b), err
}

func (this *Provider) FromJson(_json []byte, _simFlag bool) {
	json.Unmarshal(_json, this)

	if _simFlag {
		this.SetupSimClient(client.GetFakeLedger())
	}
}

func (this *Provider) AfterLoad() {
	if this.Files == nil {
		this.Files = make(map[string]*File)
	}
	if *helper.SimFlag {
		this.SetupSimClient(client.GetFakeLedger())
	} else {
		this.SetupEthClient()
	}
}

func (this *Provider) ReadFile(_f *File) ([]byte, error) {
	path := helper.MakeDumpFilePath(this.Name, _f.Filename)
	return helper.ReadFile(path)
}

func (this *Provider) ReadTagFile(_f *File) (*pdp.TagDataSet, error) {
	path := helper.MakeDumpFilePath(this.Name, _f.TagFilename)
	data, err := helper.ReadFile(path)
	if err != nil { return nil, err }

	var setTagData pdp.TagDataSet
	json.Unmarshal(data, &setTagData)

	return &setTagData, nil
}
