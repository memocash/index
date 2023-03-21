package wallet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/jchavannes/bchutil"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcutil"
	"github.com/jchavannes/jgo/jerr"
	"strings"
)

const (
	BitcoinPrefix  = "bitcoin"
	CashAddrPrefix = "bitcoincash"
	SlpAddrPrefix  = "simpleledger"

	UnknownAddressTypeErrorMessage = "unknown address type"
)

func GetAddress(pubKey []byte) Address {
	if len(pubKey) == 0 {
		return Address{}
	}
	addr, err := btcutil.NewAddressPubKey(pubKey, GetMainNetParams())
	if err != nil {
		//fmt.Println(jerr.Get("error getting address", err))
		return Address{}
	}
	address, err := btcutil.DecodeAddress(addr.EncodeAddress(), GetMainNetParams())
	if err != nil {
		//fmt.Printf("error decoding address: %v\n", err)
		return Address{}
	}
	return Address{
		address: address,
	}
}

func GetAddressFromString(addressString string) Address {
	if addressString == "" {
		return Address{}
	}
	address, err := GetAddressFromStringErr(addressString)
	if err != nil {
		jerr.Get("error getting address from string", err).PrintWithStack()
		return Address{}
	}
	return *address
}

func GetAddressFromStringErr(addressString string) (*Address, error) {
	address, err := btcutil.DecodeAddress(addressString, GetMainNetParams())
	if err != nil {
		if len(addressString) > 0 {
			address, err = bchutil.DecodeAddress(addressString, GetMainNetParams())
			if err != nil && !strings.Contains(addressString, ":") {
				address, err = bchutil.DecodeAddress(SlpAddrPrefix+":"+addressString, GetMainNetParams())
			}
		}
		if err != nil {
			return nil, jerr.Getf(err, "error decoding address: %s", addressString)
		}
		if strings.HasPrefix(addressString, "p") || strings.HasPrefix(addressString, "simpleledger:p") ||
			strings.HasPrefix(addressString, "bitcoincash:p") || strings.HasPrefix(addressString, "bitcoin:p") {
			address, err = btcutil.NewAddressScriptHashFromHash(address.ScriptAddress(), GetMainNetParams())
			if err != nil {
				return nil, jerr.Getf(err, "error getting p2sh address: %s", addressString)
			}
		} else {
			address, err = btcutil.NewAddressPubKeyHash(address.ScriptAddress(), GetMainNetParams())
			if err != nil {
				return nil, jerr.Getf(err, "error getting btc address from bch address: %s", addressString)
			}
		}
	}
	return &Address{
		address: address,
	}, nil
}

func GetAddressFromPkHash(pkHash []byte) Address {
	addr, err := btcutil.NewAddressPubKeyHash(pkHash, GetMainNetParams())
	if err != nil {
		//fmt.Println(jerr.Get("error getting address", err))
		return Address{}
	}
	address, err := btcutil.DecodeAddress(addr.EncodeAddress(), GetMainNetParams())
	if err != nil {
		//fmt.Printf("error decoding address: %v\n", err)
		return Address{}
	}
	return Address{
		address: address,
	}
}

func GetAddressFromPkHashNew(pkHash []byte) (Address, error) {
	addr, err := btcutil.NewAddressPubKeyHash(pkHash, GetMainNetParams())
	if err != nil {
		return Address{}, fmt.Errorf("error getting address; %w", err)
	}
	address, err := btcutil.DecodeAddress(addr.EncodeAddress(), GetMainNetParams())
	if err != nil {
		return Address{}, fmt.Errorf("error decoding address; %w", err)
	}
	return Address{
		address: address,
	}, nil
}

func GetAddressesForPkHashes(pkHashes [][]byte) ([]Address, error) {
	var addresses []Address
	for _, pkHash := range pkHashes {
		if len(pkHash) == 0 {
			continue
		}
		address, err := GetAddressFromPkHashNew(pkHash)
		if err != nil {
			return nil, jerr.Getf(err, "error getting address from pkHash (%x)", pkHash)
		}
		addresses = append(addresses, address)
	}
	return addresses, nil
}

const (
	TooManyAddressesErrorMsg = "error too many addresses in pk script"
	NoAddressesErrorMsg      = "error unable to find any addresses"
)

var tooManyAddressesError = jerr.New(TooManyAddressesErrorMsg)
var noAddressesError = jerr.New(NoAddressesErrorMsg)

func IsTooManyAddressesError(err error) bool {
	return jerr.HasError(err, TooManyAddressesErrorMsg)
}

func IsNoAddressesError(err error) bool {
	return jerr.HasError(err, NoAddressesErrorMsg)
}

func IsAddressQuantityError(err error) bool {
	return IsTooManyAddressesError(err) || IsNoAddressesError(err)
}

func GetAddressFromPkScript(pkScript []byte) (*Address, error) {
	_, addresses, _, err := txscript.ExtractPkScriptAddrs(pkScript, GetMainNetParams())
	if err != nil {
		return nil, jerr.Get("error extracting addresses from pk script", err)
	}
	if len(addresses) > 1 {
		return nil, jerr.Getf(tooManyAddressesError, "unexpected number of addresses (%d)", len(addresses))
	} else if len(addresses) == 0 {
		return nil, jerr.Getf(noAddressesError, "error no addresses found")
	}
	return &Address{
		address: addresses[0],
	}, nil
}

func GetAddressStringFromPkScript(pkScript []byte) string {
	scriptClass, addresses, _, err := txscript.ExtractPkScriptAddrs(pkScript, GetMainNetParams())
	if err != nil {
		return "error: " + scriptClass.String()
	}
	if len(addresses) > 1 {
		return "multiple: " + scriptClass.String()
	} else if len(addresses) == 0 {
		return "unknown: " + scriptClass.String()
	}
	return addresses[0].String()
}

func GetAddressFromRedeemScript(redeemScript []byte) (*Address, error) {
	address, err := btcutil.NewAddressScriptHash(redeemScript, GetMainNetParams())
	if err != nil {
		return nil, jerr.Get("error getting address script hash from redeem script", err)
	}
	return &Address{
		address: address,
	}, nil
}

func GetAddressFromScriptHash(scriptHash []byte) Address {
	address, _ := GetAddressFromScriptHashNew(scriptHash)
	if address == nil {
		return Address{}
	}
	return *address
}

func GetAddressFromScriptHashNew(scriptHash []byte) (*Address, error) {
	address, err := btcutil.NewAddressScriptHashFromHash(scriptHash, GetMainNetParams())
	if err != nil {
		return nil, jerr.Get("error getting address script hash from hash", err)
	}
	return &Address{
		address: address,
	}, nil
}

func GetAddressFromSignatureScript(unlockScript []byte) (*Address, error) {
	unlockString, err := txscript.DisasmString(unlockScript)
	if err != nil {
		return nil, jerr.Get("error disasm unlock script", err)
	}
	split := strings.Split(unlockString, " ")
	if len(split) == 2 {
		pubKey, err := hex.DecodeString(split[1])
		if err != nil {
			return nil, jerr.Get("error decoding pub key", err)
		}
		address := GetAddress(pubKey)
		return &address, nil
	} else if len(split) > 0 {
		redeemScript, err := hex.DecodeString(split[len(split)-1])
		if err != nil {
			return nil, jerr.Get("error decoding script hash", err)
		}
		address, err := GetAddressFromRedeemScript(redeemScript)
		if err != nil {
			return nil, jerr.Get("error getting address from redeem script", err)
		}
		return address, nil
	}
	return nil, jerr.Newf("error unexpected number of items in unlock script (%d)", len(split))
}

func GetAddressListPkHashes(addresses []Address) [][]byte {
	var pkHashes [][]byte
	for _, address := range addresses {
		pkHashes = append(pkHashes, address.GetPkHash())
	}
	return pkHashes
}

type Address struct {
	address btcutil.Address
}

func (a Address) GetEncoded() string {
	if a.address == nil {
		return ""
	}
	return a.address.EncodeAddress()
}

func (a Address) GetCashAddrString() string {
	if a.address == nil {
		return ""
	}
	if a.IsP2SH() {
		cashAddr, err := bchutil.NewCashAddressScriptHashFromHash(a.GetPkHash(), GetMainNetParams())
		if err == nil {
			return cashAddr.String()
		}
	} else {
		cashAddr, err := bchutil.NewCashAddressPubKeyHash(a.GetPkHash(), GetMainNetParams())
		if err == nil {
			return cashAddr.String()
		}
	}
	return ""
}

func (a Address) GetSlpAddrString() string {
	var addr string
	if a.address != nil {
		if a.IsP2SH() {
			slpAddr, err := bchutil.NewSlpAddressScriptHashFromHash(a.GetPkHash(), GetMainNetParams())
			if err == nil {
				addr = slpAddr.String()
			}
		} else {
			slpAddr, err := bchutil.NewSlpAddressPubKeyHash(a.GetPkHash(), GetMainNetParams())
			if err == nil {
				addr = slpAddr.String()
			}
		}
		addr = strings.TrimPrefix(addr, SlpAddrPrefix+":")
	}
	return addr
}

func (a Address) GetAddress() btcutil.Address {
	return a.address
}

func (a Address) IsSet() bool {
	return a.address != nil
}

func (a Address) GetPkHash() []byte {
	return a.ScriptAddress()
}

func (a Address) GetAddr() Addr {
	b := append([]byte{0x00}, a.GetPkHash()...)
	var r [25]byte
	copy(r[:], append(b, chainhash.DoubleHashB(b)[:4]...))
	return r
}

func (a Address) ScriptAddress() []byte {
	if a.address == nil {
		return []byte{}
	}
	return a.address.ScriptAddress()
}

func (a Address) IsP2SH() bool {
	if a.address == nil {
		return false
	}
	switch a.address.(type) {
	case *btcutil.AddressScriptHash:
		return true
	}
	return false
}

func (a Address) IsP2PKH() bool {
	if a.address == nil {
		return false
	}
	switch a.address.(type) {
	case *btcutil.AddressPubKeyHash:
		return true
	}
	return false
}

func (a Address) IsP2PK() bool {
	if a.address == nil {
		return false
	}
	switch a.address.(type) {
	case *btcutil.AddressPubKey:
		return true
	}
	return false
}

func (a Address) IsSame(b Address) bool {
	if !a.IsSet() || !b.IsSet() {
		return false
	}
	return bytes.Equal(a.GetPkHash(), b.GetPkHash())
}
