package entity

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/client"
)

type Manager struct {
	Name string
	ParamXZ21 *pdp.XZ21Param
	Addr common.Address
	PrivKey string

	param *pdp.PairingParam

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
	this.client.EnrollUser(_su.Addr, _su.PublicKeyData)
}

func (this *Manager) EnrollAuditor(_tpa *Auditor)  {
	this.client.EnrollAuditor(_tpa.Addr)
}

func (this *Manager) GetParam() *pdp.PairingParam {
	return this.param
}

func (this *Manager) GetName() string {
	return this.Name
}

func (this *Manager) ToJson() (string, error) {
	b, err := json.MarshalIndent(this, "", "\t")
	return string(b), err
}

func (this *Manager) FromJson(_json []byte, _simFlag bool) {
	json.Unmarshal(_json, this)

	if _simFlag {
		this.SetupSimClient(client.GetFakeLedger())
	}
}

func (this *Manager) AfterLoad() {
	this.param = pdp.GenParamFromXZ21Param(this.ParamXZ21)
}
