package wallet

import (
	"fmt"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcutil"
)

// Addr32 is a P2SH32 address: the HASH256 of a redeem script, as committed
// to by an OP_HASH256 <32 bytes> OP_EQUAL lock script (activated on BCH in
// May 2023). Legacy base58 has no form for it; it is written as a CashAddr
// with the 256-bit size bits set. Addr stays the type for the 20-byte
// P2PKH and P2SH20 kinds, mirroring bchutil's separate AddressScriptHash32.
type Addr32 [32]byte

const (
	// AddrVersionP2SH32 tags a P2SH32 address in LockAddr key bytes. It is
	// not a base58 version (base58 has no P2SH32 form) and is distinct from
	// the legacy P2PKH (0) and P2SH (5) versions so key bytes stay
	// self-describing.
	AddrVersionP2SH32 = 32

	p2sh32LockScriptLength = 35
)

func (a Addr32) String() string {
	return encodeCashAddr(CashAddrPrefix, true, a[:])
}

// GetAddr32FromScriptHash builds a P2SH32 address from a 32-byte script hash.
func GetAddr32FromScriptHash(scriptHash []byte) (*Addr32, error) {
	if len(scriptHash) != 32 {
		return nil, fmt.Errorf("error p2sh32 script hash is not 32 bytes: %d", len(scriptHash))
	}
	var addr = new(Addr32)
	copy(addr[:], scriptHash)
	return addr, nil
}

// GetP2sh32AddrFromLockScript resolves an OP_HASH256 <32> OP_EQUAL lock script.
func GetP2sh32AddrFromLockScript(lockScript []byte) (*Addr32, error) {
	if len(lockScript) != p2sh32LockScriptLength ||
		lockScript[0] != txscript.OP_HASH256 || lockScript[1] != txscript.OP_DATA_32 ||
		lockScript[34] != txscript.OP_EQUAL {
		return nil, fmt.Errorf("error lock script is not a p2sh32 address")
	}
	return GetAddr32FromScriptHash(lockScript[2:34])
}

// GetP2shRedeemScriptFromUnlockScript returns the redeem script a P2SH-shaped
// unlock script carries as its final push. A P2SH20 spend and a P2SH32 spend
// of the same script have identical unlock scripts; which hash the output
// demanded is only known from the output, so callers that want an address
// derive both (see GetP2shAddrsFromRedeemScript).
func GetP2shRedeemScriptFromUnlockScript(unlockScript []byte) ([]byte, error) {
	if len(unlockScript) < 2 {
		return nil, fmt.Errorf("error unlock script is too short for p2sh")
	}
	opCodes := Decompile(unlockScript)
	if len(opCodes) == 0 {
		return nil, fmt.Errorf("error decompiling unlock script for p2sh")
	}
	redeemScript := opCodes[len(opCodes)-1].Data
	if len(Decompile(redeemScript)) == 0 {
		return nil, fmt.Errorf("error decompiling redeem script for p2sh")
	}
	return redeemScript, nil
}

// GetP2shAddrsFromRedeemScript returns both addresses a redeem script can be
// locked under: the P2SH20 (HASH160) form and the P2SH32 (HASH256) form.
func GetP2shAddrsFromRedeemScript(redeemScript []byte) (Addr, Addr32) {
	var addr32 Addr32
	copy(addr32[:], chainhash.DoubleHashB(redeemScript))
	return *GetAddrFromScriptHash(btcutil.Hash160(redeemScript)), addr32
}
