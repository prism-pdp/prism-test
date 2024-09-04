package entity

import (
	"encoding/json"
	"io"
	"os"
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/client"
)

type Manager struct {
	Param pdp.PairingParam // TODO: Use Params struct (same with fake ledger)

	client client.BaseClient
}

func GenManager( _client client.BaseClient) *Manager {
	manager := new(Manager)
	manager.Param = pdp.GenPairingParam()
	manager.client = _client
	return manager
}

func LoadManager(_path string, _client client.BaseClient) *Manager {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := io.ReadAll(f)
	if err != nil { panic(err) }

	sm := new(Manager)
	json.Unmarshal(s, &sm)

	sm.client = _client

	return sm
}

func (this *Manager) RegisterParam() {
	xz21Param := this.Param.ToXZ21Param()
	this.client.RegisterParam(
		xz21Param.P,
		xz21Param.G,
		xz21Param.U,
	)
}

func (this *Manager) EnrollUser(_addr common.Address, _pubKey []byte)  {
	this.client.EnrollAccount(_addr, _pubKey)
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