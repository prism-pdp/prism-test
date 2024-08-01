package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/Nik-U/pbc"

	pdp "github.com/dpduado/dpduado-go/xz21"
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

func reqUpload(_ctx *Context, _data []byte, _chunkSize uint32, _ram *RamSU) (bool, *pdp.Metadata, error) {
	hash := sha256.Sum256(_data)
	isFound, err := _ctx.Session.SearchFile(hash)

	if err != nil { return false, nil, err }

	if isFound {
		fmt.Printf("File is already stored. (hash:%s)\n", hex.EncodeToString(hash[:]))
		return false, nil, nil
	} else {
		chunk, err := pdp.SplitData(data, _chunkSize)

		if err != nil { return false, nil, err }

		meta := pdp.GenMetadata(&param, _ram.key.PrivateKey, chunk)
		return true, meta, nil
	}
}

func genDedupProof(_ctx *Context, _chal *pdp.Chal, _data []byte, _chunkSize uint32) *pbc.Element {
	xz21Para, err := _ctx.Session.GetPara()
	para := pdp.GenParamFromXZ21Para(&xz21Para)

	chunk, err := pdp.SplitData(data, _chunkSize)
	if err != nil { panic(err) }

	proof := pdp.GenProof(para, _chal, chunk)

	return proof
}