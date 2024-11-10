package client

import (
	"context"
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
		helper.PrintLog("Failed to call RegisterParam contract (caller:%s)", this.Addr)
		return nil
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog("Failed to complete RegisterParam contract (caller:%s)", this.Addr)
		return err
	}

	helper.PrintLog("Completed smart contract (name:RegisterParam, caller:%s, gasUsed:%d, gasPrice:%d)", this.Addr, receipt.GasUsed, receipt.EffectiveGasPrice)

	return err
}

func (this *EthClient) RegisterFile(_hash [32]byte, _splitNum uint32, _owner common.Address) error {
	tx, err := this.Session.RegisterFile(_hash, _splitNum, _owner)
	if err != nil {
		helper.PrintLog("Failed to call RegisterFile contract (caller:%s, owner:%s, file:%s)", this.Addr, _owner, helper.Hex(_hash[:]))
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog("Failed to complete RegisterFile contract (caller:%s, owner:%s, file:%s)", this.Addr, _owner, helper.Hex(_hash[:]))
		return err
	}

	helper.PrintLog("Completed smart contract (name:RegisterFile, caller:%s, gasUsed:%d, gasPrice:%d, owner:%s, file:%s)", this.Addr, receipt.GasUsed, receipt.EffectiveGasPrice, _owner, helper.Hex(_hash[:]))

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
	account, err := this.Session.GetUserAccount(_addr)
	return account, err
}

func (this *EthClient) EnrollAuditor(_addr common.Address) error {
	return this.enroll(0, _addr, []byte{})
}

func (this *EthClient) EnrollUser(_addr common.Address, _pubKey pdp.PublicKeyData) error {
	return this.enroll(1, _addr, _pubKey)
}

func (this *EthClient) enroll(_type int, _addr common.Address, _pubKey pdp.PublicKeyData) error {
	t := big.NewInt(int64(_type))
	tx, err := this.Session.EnrollAccount(t, _addr, _pubKey.Base())
	if err != nil {
		helper.PrintLog("Failed to call EnrollAccount contract (caller:%s, type:%d, addr:%s, key:%s)", this.Addr, _type, _addr, helper.Hex(_pubKey.Base()))
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog("Failed to complete EnrollAccount contract (caller:%s, type:%d, addr:%s, key:%s)", this.Addr, _type, _addr, helper.Hex(_pubKey.Base()))
		return err
	}
	
	helper.PrintLog("Completed smart contract (name:EnrollAccount, caller:%s, gasUsed:%d, gasPrice:%d, type:%d, addr:%s, key:%s)", this.Addr, receipt.GasUsed, receipt.EffectiveGasPrice, _type, _addr, helper.Hex(_pubKey.Base()))

	return err
}

func (this *EthClient) AppendOwner(_hash [32]byte, _owner common.Address) error {
	tx, err := this.Session.AppendOwner(_hash, _owner)
	if err != nil {
		helper.PrintLog("Failed to call AppendOwner contract (caller:%s, owner:%s, file:%s)", this.Addr, _owner, helper.Hex(_hash[:]))
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog("Failed to complete AppendOwner contract (caller:%s, owner:%s, file:%s)", this.Addr, _owner, helper.Hex(_hash[:]))
		return err
	}

	helper.PrintLog("Completed smart contract (name:AppendOwner, caller:%s, gasUsed:%d, gasPrice:%d, owner:%s, file:%s)", this.Addr, receipt.GasUsed, receipt.EffectiveGasPrice, _owner, helper.Hex(_hash[:]))

	return err
}

func (this *EthClient) SetChal(_hash [32]byte, _chalBytes []byte) error {
	tx, err := this.Session.SetChal(_hash, _chalBytes)
	if err != nil {
		if helper.ErrSetChal.Comp(err) {
			helper.PrintLog("Skiped SetChal contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:]))
			return nil
		} else {
			helper.PrintLog("Failed to call SetChal contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:]))
			return err
		}
	}

	// Receipt -- https://github.com/ethereum/go-ethereum/blob/master/core/types/receipt.go#L52
	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog("Failed to complete SetChal contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:]))
		return err
	}

	helper.PrintLog("Completed smart contract (name:SetChal, caller:%s, gasUsed:%d, gasPrice:%d, file:%s)", this.Addr, receipt.GasUsed, receipt.EffectiveGasPrice, helper.Hex(_hash[:]))

	return nil
}

func (this *EthClient) SetProof(_hash [32]byte, _proofBytes []byte) error {
	tx, err := this.Session.SetProof(_hash, _proofBytes)
	if err != nil {
		helper.PrintLog("Failed to call SetProof contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:]))
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog("Failed to complete SetProof contract (caller:%s, file:%s)", this.Addr, helper.Hex(_hash[:]))
		return err
	}

	helper.PrintLog("Completed smart contract (name:SetProof, caller:%s, gasUsed:%d, gasPrice:%d, file:%s)", this.Addr, receipt.GasUsed, receipt.EffectiveGasPrice, helper.Hex(_hash[:]))

	return err
}

func (this *EthClient) GetAuditingReqList() ([][32]byte, []pdp.XZ21AuditingReq, error) {
	fileList, reqList, err := this.Session.GetAuditingReqList()
	return fileList, reqList, err
}

func (this *EthClient) SetAuditingResult(_hash [32]byte, _result bool) error {
	tx, err := this.Session.SetAuditingResult(_hash, _result)
	if err != nil {
		helper.PrintLog("Failed to call SetAuditingResult contract (caller:%s, file:%s, result:%t)", this.Addr, helper.Hex(_hash[:]), _result)
		return err
	}

	receipt, err := bind.WaitMined(context.Background(), this.Client, tx)
	if err != nil {
		helper.PrintLog("Failed to complete SetAuditingResult contract (caller:%s, file:%s, result:%t)", this.Addr, helper.Hex(_hash[:]), _result)
		return err
	}

	helper.PrintLog("Completed smart contract (name:SetAuditingResult, caller:%s, gasUsed:%d, gasPrice:%d, file:%s, result:%t)", this.Addr, receipt.GasUsed, receipt.EffectiveGasPrice, helper.Hex(_hash[:]), _result)

	return err
}

func (this *EthClient) GetAuditingLogs(_hash [32]byte) ([]pdp.XZ21AuditingLog, error) {
	logs, err := this.Session.GetAuditingLogs(_hash)
	return logs, err
}