package client

import (
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

type BaseClient interface {
	GetAddr() common.Address
	// interface of blockchain
	GetParam() (pdp.XZ21Param, error) // E
	RegisterParam(_param string, _g []byte, _u []byte) error // A
	RegisterFile(_hash [32]byte, _splitNum uint32, _owner common.Address) error // D
	GetFileList(_addr common.Address) ([][32]byte, error)
	SearchFile(_hash [32]byte) (pdp.XZ21FileProperty, error) // C
	GetAccount(_addr common.Address) (pdp.XZ21Account, error)
	EnrollAuditor(_addr common.Address) error // B
	EnrollUser(_addr common.Address, _pubKey pdp.PublicKeyData) error // B
	AppendOwner(_hash [32]byte, _owner common.Address) error // F
	SetChal(_hash [32]byte, _chalBytes []byte) error // G
	SetProof(_hash [32]byte, _proofBytes []byte) error // I
	GetLatestAuditingLog(_hash [32]byte) (*pdp.XZ21AuditingLog, error) // J
	SetAuditingResult(_hash [32]byte, _result bool) error // K
	GetAuditingLogs(_hash [32]byte) ([]pdp.XZ21AuditingLog, error)
}

type ClientOpts struct {
	Server string
	ContractAddr string

	Addr common.Address
	PrivKey string

	Ledger *FakeLedger
}

func NewClient(_simFlag bool, _server string, _contractAddr string, _senderAddr string, _senderPrivKey string, _ledger *FakeLedger) BaseClient {
	addr := common.HexToAddress(_senderAddr)

	if _simFlag {
		return NewSimClient(_ledger, addr)
	} else {
		return NewEthClient(_server, _contractAddr, _senderPrivKey, addr)
	}
	return nil
}

func NewEthClient(_server string, _contractAddr string, _senderPrivKey string, _senderAddr common.Address) BaseClient {
	var ethClient EthClient
	ethClient.Setup(_server, _contractAddr, _senderPrivKey, _senderAddr)
	return &ethClient
}

func NewSimClient(_ledger *FakeLedger, _senderAddr common.Address) BaseClient {
	var simClient SimClient
	simClient.Setup(_senderAddr, _ledger)
	return &simClient

	return nil
}
