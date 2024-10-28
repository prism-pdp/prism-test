package entity

import (
	"encoding/json"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/client"
	"github.com/dpduado/dpduado-test/harness/helper"
)

type Manager struct {
	Name string
	ParamXZ21 pdp.XZ21Param
	Addr common.Address
	PrivKey string

	param pdp.PairingParam

	client client.BaseClient
}


func GenManager(_name string, _addr string, _privKey string, _simFlag bool) *Manager {
	sm := new(Manager)
	sm.Name = _name
	sm.Addr = common.HexToAddress(_addr)
	sm.PrivKey = _privKey
	sm.param = pdp.GenPairingParam()
	sm.ParamXZ21 = sm.param.ToXZ21Param()

	if _simFlag {
		sm.SetupSimClient(client.GetFakeLedger())
	}

	return sm
}

func LoadManager(_path string, _simFlag bool) *Manager {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	s, err := io.ReadAll(f)
	if err != nil { panic(err) }

	sm := new(Manager)
	json.Unmarshal(s, &sm)

	sm.param = pdp.GenParamFromXZ21Param(&sm.ParamXZ21)

	if _simFlag {
		sm.SetupSimClient(client.GetFakeLedger())
	}

	return sm
}

func (this *Manager) SetupSimClient(_ledger *client.FakeLedger) {
	this.client = client.NewSimClient(_ledger, this.Addr)
}

func (this *Manager) RegisterParam() {
	xz21Param := this.param.ToXZ21Param()
	this.client.RegisterParam(
		xz21Param.P,
		xz21Param.G,
		xz21Param.U,
	)
}

func (this *Manager) EnrollUser(_su *User)  {
	this.client.EnrollUser(_su.Addr, _su.PublicKeyData.Key)
}

func (this *Manager) EnrollAuditor(_tpa *Auditor)  {
	this.client.EnrollAuditor(_tpa.Addr)
}

func (this *Manager) Dump(_pathDir string) {
	s, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	path := helper.MakeDumpPath(_pathDir, this.Name)
	f, err := os.Create(path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)

	if err != nil { panic(err) }
}

func (this *Manager) GetParam() *pdp.PairingParam {
	return &this.param
}
