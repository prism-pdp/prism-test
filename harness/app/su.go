package main

import (
	pdp "github.com/dpduado/dpduado-go/xz21"

	"github.com/ethereum/go-ethereum/common"
)

type RamSU struct {
	addr common.Address
	key pdp.PairingKey
}

func GenRamSU(_param *pdp.PairingParam, _addr common.Address) *RamSU {
	ram := new(RamSU)
	ram.addr = _addr
	ram.key.Gen(_param)
	return ram
}