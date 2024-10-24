package client

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type SimClient struct {
	Addr common.Address `json:'addr'`
	Ledger *FakeLedger  `json:'ledger'`
}

func (this *SimClient) Setup(_addr common.Address, _ledger *FakeLedger) {
	this.Addr = _addr
	this.Ledger = _ledger
}

func (this *SimClient) GetAddr() common.Address {
	return this.Addr
}

func (this *SimClient) GetParam() (pdp.XZ21Param, error) {
	var xz21Param pdp.XZ21Param
	xz21Param.P = this.Ledger.Param.P
	xz21Param.G = this.Ledger.Param.G
	xz21Param.U = this.Ledger.Param.U
	return xz21Param, nil
}

func (this *SimClient) RegisterParam(_params string, _g []byte, _u []byte) error {
	this.Ledger.Param.P = _params
	this.Ledger.Param.G = _g
	this.Ledger.Param.U = _u
	return nil
}

func (this *SimClient) RegisterFile(_hash [32]byte, _splitNum uint32, _owner common.Address) error {
	this.Ledger.RegisterFile(_hash, _splitNum, _owner)
	return nil
}

func (this *SimClient) GetFileList(_addr common.Address) ([][32]byte, error) {
	var fileList [][32]byte
	for hashHex, prop := range this.Ledger.FileProperties {
		if prop.Creator == _addr {
			hash, err := helper.DecodeHex(hashHex)
			if err != nil { panic(err) }
			fileList = append(fileList, [32]byte(hash))
		}
	}
	return fileList, nil
}

func (this *SimClient) SearchFile(_hash [32]byte) (pdp.XZ21FileProperty, error) {
	return this.Ledger.SearchFile(_hash)
}

func (this *SimClient) GetAccount(_addr common.Address) (pdp.XZ21Account, error) {
	return *this.Ledger.Accounts[_addr], nil
}

func (this *SimClient) EnrollAccount(_type int, _addr common.Address, _pubKey []byte) error {
	this.Ledger.EnrollAccount(_type, _addr, _pubKey)
	return nil
}

func (this *SimClient) AppendOwner(_hash [32]byte, _addr common.Address) error {
	return this.Ledger.AppendOwner(_hash, _addr)
}

func (this *SimClient) SetChal(_hash [32]byte, _chalBytes []byte) error {
	hashHex := helper.Hex(_hash[:])
	if _, ok := this.Ledger.Reqs[hashHex]; ok {
		return nil
	}

	var req pdp.XZ21AuditingReq
	req.Chal = _chalBytes
	this.Ledger.Reqs[hashHex] = &req

	return nil
}

func (this *SimClient) SetProof(_hash [32]byte, _proofBytes []byte) error {
	this.Ledger.Reqs[helper.Hex(_hash[:])].Proof = _proofBytes
	return nil
}

func (this *SimClient) SetAuditingResult(_hash [32]byte, _result bool) error {
	hashHex := helper.Hex(_hash[:])

	req, ok := this.Ledger.Reqs[hashHex]
	if ok == false {
		return fmt.Errorf("Invalid request")
	}

	if len(req.Chal) <= 0 && len(req.Proof) <= 0 {
		return fmt.Errorf("Invalid status")
	}

	var log pdp.XZ21AuditingLog
	log.Req = *req
	log.Result = _result
	this.Ledger.Logs[hashHex] = append(this.Ledger.Logs[hashHex], &log)

	delete(this.Ledger.Reqs, hashHex)

	return nil
}

func (this *SimClient) GetAuditingReqList() ([][32]byte, []pdp.XZ21AuditingReq, error) {
	var fileList [][32]byte
	var reqList []pdp.XZ21AuditingReq
	for hashHex, req := range this.Ledger.Reqs {
		hash, err := helper.DecodeHex(hashHex)
		if err != nil { panic(err) }
		fileList = append(fileList, [32]byte(hash))
		reqList = append(reqList, *req)
	}
	return fileList, reqList, nil
}

func (this *SimClient) GetAuditingLogs(_hash [32]byte) ([]pdp.XZ21AuditingLog, error) {
	hashHex := helper.Hex(_hash[:])
	var logs []pdp.XZ21AuditingLog
	for _, v := range this.Ledger.Logs[hashHex] {
		logs = append(logs, *v)
	}
	return logs, nil
}