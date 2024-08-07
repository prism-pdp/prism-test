package entity

import (
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/helper"
)

type Manager struct {
	Param pdp.PairingParam
	session *pdp.XZ21Session
}

func GenManager(_server string, _contractAddr string, _privKey string) Manager {
	var manager Manager
	manager.Param = pdp.GenPairingParam()
	manager.session = helper.GenSession(_server, _contractAddr, _privKey)
	return manager
}

func LoadManager(_server string, _contractAddr string, _privKey string) *Manager {
	manager := new(Manager)
	manager.session = helper.GenSession(_server, _contractAddr, _privKey)
	manager.Param = helper.FetchPairingParam(manager.session)
	return manager
}

func (this *Manager) RegisterPara() error {
	xz21Para := this.Param.ToXZ21Para()
	_, err := this.session.RegisterPara(
		xz21Para.Params,
		xz21Para.G,
		xz21Para.U,
	)

	return err
}

func (this *Manager) EnrollUser(_addr common.Address, _pubKey []byte) error {
	_, err := this.session.EnrollAccount(_addr, _pubKey)
	return err
}