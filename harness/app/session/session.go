package session

import (
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

type Session interface {
	GetPara() (pdp.XZ21Para, error)
	RegisterPara(_params string, _g []byte, _u []byte)
	RegisterFile(_hash [32]byte, _owner common.Address)
	SearchFile(_hash [32]byte) *pdp.XZ21FileProperty
	SearchPublicKey(_addr common.Address) ([]byte, bool)
	EnrollAccount(_addr common.Address, _pubKey []byte)
	AppendAccount(_hash [32]byte, _owner common.Address)
	Dump(_path string)
	Load(_path string)
}

func NewSession(_mode string) Session {
	switch _mode {
	case "sim":
		var simSession SimSession
		simSession.Setup()
		return &simSession
	}

	return nil
}

func LoadSession(_mode string, _path string) Session {
	switch _mode {
	case "sim":
		var simSession SimSession
		simSession.Load(_path)
		return &simSession
	}

	return nil
}