package session

import (
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

type Session interface {
	GetPara() (pdp.XZ21Para, error)
	RegisterPara(_params string, _g []byte, _u []byte)
	RegisterFileProperty(_hash [32]byte, _splitNum uint32, _owner common.Address)
	FetchFileList() [][32]byte
	SearchFile(_hash [32]byte) *pdp.XZ21FileProperty
	SearchPublicKey(_addr common.Address) ([]byte, bool)
	EnrollAccount(_addr common.Address, _pubKey []byte)
	AppendAccount(_hash [32]byte, _owner common.Address)
	UploadChallen(_hash [32]byte, _chalBytes []byte)
	DownloadChallen() ([][32]byte, []pdp.ChalData)
	UploadProof(_hash [32]byte, _proofBytes []byte)
	DownloadAuditChallenAndProof() ([][32]byte, []pdp.ChalData, []pdp.ProofData)
	UploadAuditResult(_hash [32]byte, _result bool) error
	FetchAuditingReqList() [][32]byte
}

func NewSession(_mode string, _ledger *FakeLedger, _addr common.Address) Session {
	switch _mode {
	case "sim":
		var simSession SimSession
		simSession.Setup(_addr, _ledger)
		return &simSession
	}

	return nil
}
