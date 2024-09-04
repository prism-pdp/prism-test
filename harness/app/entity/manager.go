package entity

import (
	"encoding/json"
	"io"
	"os"
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/session"
)

type Manager struct {
	Param pdp.PairingParam // TODO: Use Params struct (same with fake ledger)

	session session.Session
}

func GenManager( _session session.Session) *Manager {
	manager := new(Manager)
	manager.Param = pdp.GenPairingParam()
	manager.session = _session
	return manager
}

func LoadManager(_path string, _session session.Session) *Manager {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := io.ReadAll(f)
	if err != nil { panic(err) }

	sm := new(Manager)
	json.Unmarshal(s, &sm)

	sm.session = _session

	return sm
}

func (this *Manager) RegisterParam() {
	xz21Param := this.Param.ToXZ21Param()
	this.session.RegisterParam(
		xz21Param.P,
		xz21Param.G,
		xz21Param.U,
	)
}

func (this *Manager) EnrollUser(_addr common.Address, _pubKey []byte)  {
	this.session.EnrollAccount(_addr, _pubKey)
}

func (this *Manager) Dump(_path string) {
	s, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)

	if err != nil { panic(err) }
}