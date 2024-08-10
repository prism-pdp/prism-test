package entity

import (
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/session"
)

type Manager struct {
	Param pdp.PairingParam

	session session.Session
}

func GenManager(_server string, _contractAddr string, _privKey string, _session session.Session) Manager {
	var manager Manager
	manager.Param = pdp.GenPairingParam()
	manager.session = _session
	return manager
}

func LoadManager(_server string, _contractAddr string, _privKey string, _session session.Session) *Manager {
	manager := new(Manager)
	manager.session = _session

	xz21Para, err := manager.session.GetPara()
	if err != nil { panic(err) }

	manager.Param = pdp.GenParamFromXZ21Para(&xz21Para)

	return manager
}

func (this *Manager) RegisterPara() {
	xz21Para := this.Param.ToXZ21Para()
	this.session.RegisterPara(
		xz21Para.Params,
		xz21Para.G,
		xz21Para.U,
	)
}

func (this *Manager) EnrollUser(_addr common.Address, _pubKey []byte)  {
	this.session.EnrollAccount(_addr, _pubKey)
}