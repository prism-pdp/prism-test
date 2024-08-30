package session

import (
	"github.com/ethereum/go-ethereum/common"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

type Session interface {
	GetPara() (pdp.XZ21Para, error) // E
	RegisterPara(_params string, _g []byte, _u []byte) // A
	RegisterFileProperty(_hash [32]byte, _splitNum uint32, _owner common.Address) // D
	FetchFileList() [][32]byte
	SearchFile(_hash [32]byte) *pdp.XZ21FileProperty // C
	SearchPublicKey(_addr common.Address) ([]byte, bool)
	EnrollAccount(_addr common.Address, _pubKey []byte) // B
	AppendAccount(_hash [32]byte, _owner common.Address) // F
	UploadChallen(_hash [32]byte, _chalBytes []byte) // G
	DownloadChallen() ([][32]byte, []pdp.ChalData) // H
	UploadProof(_hash [32]byte, _proofBytes []byte) // I
	DownloadAuditChallenAndProof() ([][32]byte, []pdp.ChalData, []pdp.ProofData) // J
	UploadAuditResult(_hash [32]byte, _result bool) error // K
	FetchAuditingReqList() [][32]byte
}

func NewSimSession(_mode string, _ledger *FakeLedger, _addr common.Address) Session {
	switch _mode {
	case "sim":
		var simSession SimSession
		simSession.Setup(_addr, _ledger)
		return &simSession
	}

	return nil
}
