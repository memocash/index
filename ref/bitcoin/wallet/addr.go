package wallet

import (
	"bytes"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcutil"
	"github.com/jchavannes/btcutil/base58"
	"github.com/jchavannes/jgo/jerr"
)

const (
	AddrVersionP2PKH = 0
	AddrVersionP2SH  = 5
)

type Addr [25]byte

func (a Addr) String() string {
	return base58.Encode(a[:])
}

func (a Addr) IsP2SH() bool {
	return a[0] == AddrVersionP2SH
}

func (a Addr) IsP2PKH() bool {
	return a[0] == AddrVersionP2PKH
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
	p2pkhAddr, errP2pkh := GetP2pkhAddrFromUnlockScript(unlockScript)
	if errP2pkh == nil {
		return p2pkhAddr, nil
	}
	p2shAddr, errP2sh := GetP2shAddrFromUnlockScript(unlockScript)
	if errP2sh == nil {
		return p2shAddr, nil
	}
	return nil, jerr.Get("error getting address from unlock script", jerr.Combine(errP2pkh, errP2sh))
}

func GetP2pkhAddrFromUnlockScript(unlockScript []byte) (*Addr, error) {
	l := len(unlockScript)
	if l < 2 {
		return nil, jerr.Newf("error unlock script is not a standard address 0: none")
	} else if int(unlockScript[0]) < txscript.OP_DATA_64 || int(unlockScript[0]) > txscript.OP_DATA_73 {
		return nil, jerr.Newf("error unlock script is not a standard address 1: %d", unlockScript[0])
	}
	s := int(unlockScript[0])
	if l < s+35 {
		return nil, jerr.Newf("error unlock script is not a standard address 2: %d %d", l, s)
	} else if unlockScript[s+1] != txscript.OP_DATA_33 && unlockScript[s+1] != txscript.OP_DATA_65 {
		return nil, jerr.Newf("error unlock script is not a standard address 3: %d %d %d", l, s, unlockScript[s+1])
	}
	var addr = new(Addr)
	copy(addr[1:21], btcutil.Hash160(unlockScript[s+2:]))
	copy(addr[21:], chainhash.DoubleHashB(addr[0:21])[:4])
	return addr, nil
}

func GetP2shAddrFromUnlockScript(unlockScript []byte) (*Addr, error) {
	if len(unlockScript) < 2 {
		return nil, jerr.New("error unlock script is too short for p2sh")
	}
	opCodes := Decompile(unlockScript)
	if len(opCodes) == 0 {
		return nil, jerr.Newf("error decompiling unlock script for p2sh")
	}
	redeemScript := opCodes[len(opCodes)-1].Data
	chunks := Decompile(redeemScript)
	if len(chunks) == 0 {
		return nil, jerr.Newf("error decompiling redeem script for p2sh")
	}
	hash := btcutil.Hash160(redeemScript)
	var addr = new(Addr)
	addr[0] = AddrVersionP2SH
	copy(addr[1:21], hash)
	copy(addr[21:], chainhash.DoubleHashB(addr[0:21])[:4])
	return addr, nil
}

func GetAddrFromLockScript(lockScript []byte) (*Addr, error) {
	p2pkhAddr, errP2pkh := GetP2pkhAddrFromLockScript(lockScript)
	if errP2pkh == nil {
		return p2pkhAddr, nil
	}
	p2shAddr, errP2sh := GetP2shAddrFromLockScript(lockScript)
	if errP2sh == nil {
		return p2shAddr, nil
	}
	return nil, jerr.Get("error getting address from lock script", jerr.Combine(errP2pkh, errP2sh))
}

func GetP2pkhAddrFromLockScript(lockScript []byte) (*Addr, error) {
	if len(lockScript) != 25 ||
		!bytes.Equal(lockScript[0:3], []byte{txscript.OP_DUP, txscript.OP_HASH160, txscript.OP_DATA_20}) ||
		!bytes.Equal(lockScript[23:], []byte{txscript.OP_EQUALVERIFY, txscript.OP_CHECKSIG}) {
		return nil, jerr.New("error lock script is not a p2pkh address")
	}
	var addr = new(Addr)
	copy(addr[1:21], lockScript[3:23])
	copy(addr[21:], chainhash.DoubleHashB(addr[0:21])[:4])
	return addr, nil
}

func GetP2shAddrFromLockScript(lockScript []byte) (*Addr, error) {
	if len(lockScript) != 23 ||
		!bytes.Equal(lockScript[0:2], []byte{txscript.OP_HASH160, txscript.OP_DATA_20}) ||
		lockScript[22] != txscript.OP_EQUAL {
		return nil, jerr.New("error lock script is not a p2sh address")
	}
	var addr = new(Addr)
	addr[0] = AddrVersionP2SH
	copy(addr[1:21], lockScript[2:22])
	copy(addr[21:], chainhash.DoubleHashB(addr[0:21])[:4])
	return addr, nil
}

func GetAddrFromBytes(b []byte) *Addr {
	var addr = new(Addr)
	if len(b) == 25 {
		copy(addr[:], b)
	}
	return addr
}
