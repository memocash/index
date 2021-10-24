package wallet

import (
	chainCfgOld "github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/jchavannes/bchutil"
	"github.com/jchavannes/btcd/chaincfg"
	"github.com/jchavannes/btcd/txscript"
)

var _mainNetParams *chaincfg.Params
var _mainNetParamsOld *chainCfgOld.Params

func GetMainNetParams() *chaincfg.Params {
	if _mainNetParams == nil {
		cpyParams := chaincfg.MainNetParams
		_mainNetParams = &cpyParams
		_mainNetParams.Net = bchutil.MainnetMagic
	}
	return _mainNetParams
}

func GetMainNetParamsOld() *chainCfgOld.Params {
	if _mainNetParamsOld == nil {
		cpyParams := chainCfgOld.MainNetParams
		_mainNetParamsOld = &cpyParams
		_mainNetParamsOld.Net = wire.BitcoinNet(bchutil.MainnetMagic)
	}
	return _mainNetParamsOld
}

const SigHashForkID txscript.SigHashType = 0x40

const (
	ChainNameABC = "abc"
	ChainNameSV  = "sv"

	ChainNameFullBCH  = "bch"
	ChainNameFullBSV  = "bsv"
	ChainNameFullBCHA = "bcha"
)

func GetChainNameFromFull(full string) string {
	switch full {
	case ChainNameFullBSV:
		return ChainNameSV
	case ChainNameFullBCHA:
		return ChainNameABC
	default:
		return ""
	}
}

func GetChainNameFullFromFull(full string) string {
	switch full {
	case ChainNameFullBSV:
		return ChainNameFullBSV
	case ChainNameFullBCH:
		return ChainNameFullBCH
	case ChainNameFullBCHA:
		return ChainNameFullBCHA
	default:
		return ""
	}
}
