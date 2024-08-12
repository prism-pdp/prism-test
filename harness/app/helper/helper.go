package helper

import (
	"encoding/hex"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"

	pdp "github.com/dpduado/dpduado-go/xz21"
)

func GenSession(_server string, _contractAddr string, _privKey string) *pdp.XZ21Session {
	cl, err := ethclient.Dial(_server)
	if err != nil { panic(err) }

	contract, err := pdp.NewXZ21(common.HexToAddress(_contractAddr), cl)
	if err != nil { panic(err) }

	privKey, err := crypto.HexToECDSA(_privKey)
	if err != nil { panic(err) }

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, big.NewInt(31337))
	if err != nil { panic(err) }

	session := pdp.XZ21Session{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: bind.TransactOpts{
			From: auth.From,
			Signer: auth.Signer,
		},
	}

	return &session
}

func Hex(_data []byte) string {
	return "0x" + hex.EncodeToString(_data)
}

func DecodeHex(_s string) ([]byte, error) {
	return hex.DecodeString(_s[2:])
}

func GetCreatorAddr(_prop *pdp.XZ21FileProperty) common.Address {
	return _prop.Owners[0]
}

func ToXZ21FileProperty(_from *pdp.XZ21FileProperty) pdp.XZ21FileProperty {
	var fileProp pdp.XZ21FileProperty
	fileProp.Owners = _from.Owners
	return fileProp
}