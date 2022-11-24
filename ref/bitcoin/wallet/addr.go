package wallet

import (
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcutil"
	"github.com/jchavannes/btcutil/base58"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jutil"
)

type Addr [25]byte

func (a Addr) String() string {
	return base58.Encode(a[:])
}

func GetAddrFromString(addrString string) (*Addr, error) {
	var addr = new(Addr)
	d := base58.Decode(addrString)
	if len(d) != 25 {
		return nil, jerr.New("error decoding base58 address, invalid address length found")
	}
	copy(addr[:], d)
	return addr, nil
}

func GetAddrFromUnlockScript(unlockScript []byte) (*Addr, error) {
	l := len(unlockScript)
	if l < 2 || !jutil.InIntSlice(int(unlockScript[0]),
		[]int{txscript.OP_DATA_64, txscript.OP_DATA_65, txscript.OP_DATA_71, txscript.OP_DATA_72}) {
		return nil, jerr.New("error unlock script is not a standard address")
	}
	s := int(unlockScript[0])
	if l < s+35 || unlockScript[s+1] != txscript.OP_DATA_33 {
		return nil, jerr.New("error unlock script is not a standard address")
	}
	var addr = new(Addr)
	copy(addr[:], btcutil.Hash160(unlockScript[s+2:]))
	return addr, nil
}

func GetAddrFromLockScript(lockScript []byte) (*Addr, error) {
	if len(lockScript) != 25 || lockScript[0] != txscript.OP_DUP || lockScript[1] != txscript.OP_HASH160 ||
		lockScript[2] != txscript.OP_DATA_20 || lockScript[23] != txscript.OP_EQUALVERIFY ||
		lockScript[24] != txscript.OP_CHECKSIG {
		return nil, jerr.New("error lock script is not a standard address")
	}
	var addr = new(Addr)
	copy(addr[:], lockScript[3:23])
	return addr, nil
}
