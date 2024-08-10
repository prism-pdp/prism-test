package session

import (
	"encoding/json"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

type SimSession struct {
	Ledger FakeLedger `json:'ledger'`
}

func (this *SimSession) Setup() {
	this.Ledger = GenFakeLedger()
}

func (this *SimSession) GetPara() (pdp.XZ21Para, error) {
	var xz21Para pdp.XZ21Para
	xz21Para.Params = this.Ledger.Params.Params
	xz21Para.G = this.Ledger.Params.G
	xz21Para.U = this.Ledger.Params.U
	return xz21Para, nil
}

func (this *SimSession) RegisterPara(_params string, _g []byte, _u []byte) {
	this.Ledger.Params.Params = _params
	this.Ledger.Params.G = _g
	this.Ledger.Params.U = _u
}

func (this *SimSession) RegisterFile(_hash [32]byte, _owner common.Address) {
	this.Ledger.RegisterFile(_hash, _owner)
}

func (this *SimSession) SearchFile(_hash [32]byte) *pdp.XZ21File {
	return this.Ledger.SearchFile(_hash)
}

func (this *SimSession) SearchPublicKey(_addr common.Address) ([]byte, bool) {
	if v, ok := this.Ledger.Accounts[_addr]; ok {
		return v.Key, true
	}
	return []byte{0}, false
}

func (this *SimSession) EnrollAccount(_addr common.Address, _pubKey []byte) {
	this.Ledger.EnrollAccount(_addr, _pubKey)
}

func (this *SimSession) AppendAccount(_hash [32]byte, _addr common.Address) {
	this.Ledger.AppendAccount(_hash, _addr)
}

func (this *SimSession) Dump(_path string) {
	s, err := json.MarshalIndent(this, "", "\t")
	if err != nil { panic(err) }

	f, err := os.Create(_path)
	if err != nil { panic(err) }
	defer f.Close()

	_, err = f.Write(s)
	if err != nil { panic(err) }
}

func (this *SimSession) Load(_path string) {
	f, err := os.Open(_path)
	if err != nil { panic(err) }
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil { panic(err) }

	json.Unmarshal(b, this)
}