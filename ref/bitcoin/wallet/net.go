package wallet

import (
	"github.com/jchavannes/bchutil"
	"github.com/jchavannes/btcd/chaincfg"
	"github.com/jchavannes/btcd/txscript"
)

var _mainNetParams *chaincfg.Params

func GetMainNetParams() *chaincfg.Params {
	if _mainNetParams == nil {
		cpyParams := chaincfg.MainNetParams
		_mainNetParams = &cpyParams
		_mainNetParams.Net = bchutil.MainnetMagic
	}
	return _mainNetParams
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
