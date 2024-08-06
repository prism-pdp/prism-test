package helper

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

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

func FetchPairingParam(_session *pdp.XZ21Session) pdp.PairingParam {
	xz21Para, err := _session.GetPara()
	if err != nil { panic(err) }

	param := pdp.GenParamFromXZ21Para(&xz21Para)

	return param
}