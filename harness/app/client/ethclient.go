package client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type EthClient struct {
	Addr common.Address `json:'addr'`
	Client *ethclient.Client
	Session pdp.XZ21Session
}

func (this *EthClient) Setup(_server string, _contractAddr string, _privKey string, _addr common.Address) {
	this.Addr = _addr
	this.Client, this.Session = helper.GenXZ21Session(_server, _contractAddr, _privKey)
}

func (this *EthClient) GetAddr() common.Address {
	return this.Addr
}

func (this *EthClient) GetParam() (pdp.XZ21Param, error) {
	xz21Param, err := this.Session.GetParam()
	return xz21Param, err
}

func (this *EthClient) RegisterParam(_params string, _g []byte, _u []byte) error {
	tx, err := this.Session.RegisterParam(_params, _g, _u)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to call RegisterParam contract (caller:%s)", this.Addr))
		return nil
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to complete RegisterParam contract (caller:%s)", this.Addr))
		return err
	}

	helper.PrintLog(fmt.Sprintf("Completed RegisterParam contract (caller:%s, gasUsed:%d)", this.Addr, receipt.GasUsed))

	return err
}

func (this *EthClient) RegisterFile(_hash [32]byte, _splitNum uint32, _owner common.Address) error {
	tx, err := this.Session.RegisterFile(_hash, _splitNum, _owner)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to call RegisterFile contract (caller:%s, owner:%s, file:%s)", this.Addr, _owner, helper.Hex(_hash[:])))
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to complete RegisterFile contract (caller:%s, owner:%s, file:%s)", this.Addr, _owner, helper.Hex(_hash[:])))
		return err
	}

	helper.PrintLog(fmt.Sprintf("Completed RegisterFile contract (caller:%s, owner:%s, file:%s, gasUsed:%d)", this.Addr, _owner, helper.Hex(_hash[:]), receipt.GasUsed))

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

func (this *EthClient) GetAccount(_addr common.Address) (pdp.XZ21Account, error) {
	account, err := this.Session.GetAccount(_addr)
	return account, err
}

func (this *EthClient) EnrollAuditor(_addr common.Address) error {
	return this.enroll(0, _addr, []byte{})
}

func (this *EthClient) EnrollUser(_addr common.Address, _pubKey []byte) error {
	return this.enroll(1, _addr, _pubKey)
}

func (this *EthClient) enroll(_type int, _addr common.Address, _pubKey []byte) error {
	t := big.NewInt(int64(0))
	tx, err := this.Session.EnrollAccount(t, _addr, _pubKey)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to call EnrollAccount contract (caller:%s, addr:%s, key:%s)", this.Addr, _addr, helper.Hex(_pubKey[:])))
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to complete EnrollAccount contract (caller:%s, addr:%s, key:%s)", this.Addr, _addr, helper.Hex(_pubKey[:])))
		return err
	}
	
	helper.PrintLog(fmt.Sprintf("Completed EnrollAccount contract (caller:%s, addr:%s, key:%s, gasUsed:%d)", this.Addr, _addr, helper.Hex(_pubKey[:]), receipt.GasUsed))

	return err
}

func (this *EthClient) AppendOwner(_hash [32]byte, _owner common.Address) error {
	tx, err := this.Session.AppendOwner(_hash, _owner)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to call AppendOwner contract (caller:%s, owner:%s, file:%s)", this.Addr, _owner, helper.Hex(_hash[:])))
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to complete AppendOwner contract (caller:%s, owner:%s, file:%s)", this.Addr, _owner, helper.Hex(_hash[:])))
		return err
	}

	helper.PrintLog(fmt.Sprintf("Completed AppendOwner contract (caller:%s, owner:%s, file:%s, gasUsed:%d)", this.Addr, _owner, helper.Hex(_hash[:]), receipt.GasUsed))

	return err
}

func (this *EthClient) SetChal(_hash [32]byte, _chalBytes []byte) error {
	tx, err := this.Session.SetChal(_hash, _chalBytes)
	if err != nil {
		if helper.ErrSetChal.Comp(err) {
			helper.PrintLog(fmt.Sprintf("Skiped SetChal contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:])))
			return nil
		} else {
			helper.PrintLog(fmt.Sprintf("Failed to call SetChal contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:])))
			return err
		}
	}

	// Receipt -- https://github.com/ethereum/go-ethereum/blob/master/core/types/receipt.go#L52
	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to complete SetChal contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:])))
		return err
	}

	helper.PrintLog(fmt.Sprintf("Completed SetChal contract (caller:%s, file:%s, gasUsed:%d)", this.Addr, helper.Hex(_hash[:]), receipt.GasUsed))

	return nil
}

func (this *EthClient) SetProof(_hash [32]byte, _proofBytes []byte) error {
	tx, err := this.Session.SetProof(_hash, _proofBytes)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to call SetProof contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:])))
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to complete SetProof contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:])))
		return err
	}

	helper.PrintLog(fmt.Sprintf("Completed SetProof contract (caller:%s, file:%s, gasUsed:%d)", this.Addr, helper.Hex(_hash[:]), receipt.GasUsed))

	return err
}

func (this *EthClient) GetAuditingReqList() ([][32]byte, []pdp.XZ21AuditingReq, error) {
	fileList, reqList, err := this.Session.GetAuditingReqList()
	return fileList, reqList, err
}

func (this *EthClient) SetAuditingResult(_hash [32]byte, _result bool) error {
	tx, err := this.Session.SetAuditingResult(_hash, _result)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to call SetAuditingResult contract (caller:%s, file:%s, result:%t)", this.Addr, helper.Hex(_hash[:]), _result))
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog(fmt.Sprintf("Failed to complete SetAuditingResult contract (caller:%s, file:%s, result:%t)", this.Addr, helper.Hex(_hash[:]), _result))
		return err
	}

	helper.PrintLog(fmt.Sprintf("Completed SetAuditingResult contract (caller:%s, file:%s, result:%t, gasUsed:%d)", this.Addr, helper.Hex(_hash[:]), _result, receipt.GasUsed))

	return err
}

func (this *EthClient) GetAuditingLogs(_hash [32]byte) ([]pdp.XZ21AuditingLog, error) {
	logs, err := this.Session.GetAuditingLogs(_hash)
	return logs, err
}