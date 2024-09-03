package session

import (
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

type Session interface {
	GetParam() (pdp.XZ21Param, error) // E
	RegisterParam(_param string, _g []byte, _u []byte) error // A
	RegisterFile(_hash [32]byte, _splitNum uint32, _owner common.Address) error // D
	GetFileList(_addr common.Address) ([][32]byte, error)
	SearchFile(_hash [32]byte) (pdp.XZ21FileProperty, error) // C
	GetAccount(_addr common.Address) (pdp.XZ21Account, error)
	EnrollAccount(_addr common.Address, _pubKey []byte) error // B
	AppendOwner(_hash [32]byte, _owner common.Address) error // F
	SetChal(_hash [32]byte, _chalBytes []byte) (bool, error) // G
	GetChalList() ([][32]byte, []pdp.ChalData, error) // H
	SetProof(_hash [32]byte, _proofBytes []byte) error // I
	GetAuditingReqList() ([][32]byte, []pdp.XZ21AuditingReq, error) // J
	SetAuditingResult(_hash [32]byte, _result bool) error // K
	GetAuditingLogs(_hash [32]byte) ([]pdp.XZ21AuditingLog, error)
}

func NewSession(_server string, _contractAddr string, _privKey string, _addr common.Address) Session {
	var ethClient EthClient
	ethClient.Setup(_server, _contractAddr, _privKey, _addr)
	return &ethClient
}

func NewSimSession(_ledger *FakeLedger, _addr common.Address) Session {
	var simSession SimSession
	simSession.Setup(_addr, _ledger)
	return &simSession

	return nil
}
