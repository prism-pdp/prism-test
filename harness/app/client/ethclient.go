package client

import (
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type EthClient struct {
	Addr common.Address `json:'addr'`
	Session pdp.XZ21Session
}

func (this *EthClient) Setup(_server string, _contractAddr string, _privKey string, _addr common.Address) {
	this.Addr = _addr
	this.Session = helper.GenXZ21Session(_server, _contractAddr, _privKey)
}

func (this *EthClient) GetAddr() common.Address {
	return this.Addr
}

func (this *EthClient) GetParam() (pdp.XZ21Param, error) {
	xz21Param, err := this.Session.GetParam()
	return xz21Param, err
}

func (this *EthClient) RegisterParam(_params string, _g []byte, _u []byte) error {
	_, err := this.Session.RegisterParam(_params, _g, _u)
	return err
}

func (this *EthClient) RegisterFile(_hash [32]byte, _splitNum uint32, _owner common.Address) error {
	_, err := this.Session.RegisterFile(_hash, _splitNum, _owner)
	return err
}

func (this *EthClient) GetFileList(_addr common.Address) ([][32]byte, error) {
	fileList, err := this.Session.GetFileList(_addr)
	return fileList, err
}

func (this *EthClient) SearchFile(_hash [32]byte) (pdp.XZ21FileProperty, error){
	fileProp, err := this.Session.SearchFile(_hash)
	return fileProp, err
}

// Use GetAccount
//func (this *EthClient) SearchPublicKey(_addr common.Address) ([]byte, bool, error) {
//}
func (this *EthClient) GetAccount(_addr common.Address) (pdp.XZ21Account, error) {
	account, err := this.Session.GetAccount(_addr)
	return account, err
}

func (this *EthClient) EnrollAccount(_addr common.Address, _pubKey []byte) error {
	_, err := this.Session.EnrollAccount(_addr, _pubKey)
	return err
}

func (this *EthClient) AppendOwner(_hash [32]byte, _owner common.Address) error {
	_, err := this.Session.AppendOwner(_hash, _owner)
	return err
}

func (this *EthClient) SetChal(_hash [32]byte, _chalBytes []byte) (bool, error) {
	_, err := this.Session.SetChal(_hash, _chalBytes)
	// TODO: retrieve return value from Transaction
	return true, err
}

func (this *EthClient) GetChalList() ([][32]byte, []pdp.ChalData, error) {
	fileList, chalList, err := this.Session.GetChalList()
	if err != nil {
		return make([][32]byte, 0), make([]pdp.ChalData, 0), err
	}
	chalDataList := make([]pdp.ChalData, 0)
	for _, chal := range chalList {
		chalData, err := pdp.DecodeToChalData(chal)
		if err != nil { panic(err) }
		chalDataList = append(chalDataList, chalData)
	}

	return fileList, chalDataList, nil
}

func (this *EthClient) SetProof(_hash [32]byte, _proofBytes []byte) error {
	_, err := this.Session.SetProof(_hash, _proofBytes)
	return err
}

func (this *EthClient) GetAuditingReqList() ([][32]byte, []pdp.XZ21AuditingReq, error) {
	fileList, reqList, err := this.Session.GetAuditingReqList()
	return fileList, reqList, err
}

func (this *EthClient) SetAuditingResult(_hash [32]byte, _result bool) error {
	_, err := this.Session.SetAuditingResult(_hash, _result)
	return err
}

func (this *EthClient) GetAuditingLogs(_hash [32]byte) ([]pdp.XZ21AuditingLog, error) {
	logs, err := this.Session.GetAuditingLogs(_hash)
	return logs, err
}