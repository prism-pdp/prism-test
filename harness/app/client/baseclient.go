package client

import (
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/dpduado/dpduado-test/harness/types"
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
	EnrollAccount(_addr common.Address, _pubKey []byte) error // B
	AppendOwner(_hash [32]byte, _owner common.Address) error // F
	SetChal(_hash [32]byte, _chalBytes []byte) (bool, error) // G
	GetChalList() ([][32]byte, []pdp.ChalData, error) // H
	SetProof(_hash [32]byte, _proofBytes []byte) error // I
	GetAuditingReqList() ([][32]byte, []pdp.XZ21AuditingReq, error) // J
	SetAuditingResult(_hash [32]byte, _result bool) error // K
	GetAuditingLogs(_hash [32]byte) ([]pdp.XZ21AuditingLog, error)
}

type ClientOpts struct {
	Server string
	ContractAddr string

	AddrTable map[types.EntityType]common.Address
	PrivKeyTable map[types.EntityType]string

	Ledger *FakeLedger
}

func NewClientOpts() ClientOpts {
	var opts ClientOpts
	opts.AddrTable = make(map[types.EntityType]common.Address)
	opts.PrivKeyTable = make(map[types.EntityType]string)
	return opts
}

func NewClient(_mode string, _entity types.EntityType, _opts *ClientOpts) BaseClient {
	switch _mode {
	case "sim":
		return NewSimClient(_opts.Ledger, _opts.AddrTable[_entity])
	}
	return nil
}

func NewEthClient(_server string, _contractAddr string, _privKey string, _addr common.Address) BaseClient {
	var ethClient EthClient
	ethClient.Setup(_server, _contractAddr, _privKey, _addr)
	return &ethClient
}

func NewSimClient(_ledger *FakeLedger, _addr common.Address) BaseClient {
	var simClient SimClient
	simClient.Setup(_addr, _ledger)
	return &simClient

	return nil
}
