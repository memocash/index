package wallet

import (
	"bytes"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
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
	copy(addr[1:21], btcutil.Hash160(unlockScript[s+2:]))
	copy(addr[21:], chainhash.DoubleHashB(addr[0:21])[:4])
	return addr, nil
}

func GetAddrFromLockScript(lockScript []byte) (*Addr, error) {
	if len(lockScript) != 25 ||
		!bytes.Equal(lockScript[0:3], []byte{txscript.OP_DUP, txscript.OP_HASH160, txscript.OP_DATA_20}) ||
		!bytes.Equal(lockScript[23:], []byte{txscript.OP_EQUALVERIFY, txscript.OP_CHECKSIG}) {
		return nil, jerr.New("error lock script is not a standard address")
	}
	var addr = new(Addr)
	copy(addr[1:21], lockScript[3:23])
	copy(addr[21:], chainhash.DoubleHashB(addr[0:21])[:4])
	return addr, nil
}
