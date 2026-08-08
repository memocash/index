// Package slp implements strict SLP message parsing and validation per the
// SLP token-type-1 and NFT1 specifications. Parsing here is consensus-strict
// (exact field sizes, push-only scripts); the lenient transcription layer in
// node/obj/op_return/save is intentionally separate.
package slp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
)

const (
	MaxSendOutputs = 19

	opReturn    = 0x6a
	opPushData1 = 0x4c
	opPushData2 = 0x4d
	opPushData4 = 0x4e
)

type Msg struct {
	TokenType uint16
	TxType    memo.SlpType
	Genesis   *GenesisMsg
	Mint      *MintMsg
	Send      *SendMsg
}

type GenesisMsg struct {
	Ticker    []byte
	Name      []byte
	DocUrl    []byte
	DocHash   []byte // empty or 32 bytes
	Decimals  uint8
	BatonVout int // 0 = no baton
	Quantity  uint64
}

type MintMsg struct {
	TokenHash [32]byte // internal (little-endian) byte order, matching TxHash convention
	BatonVout int      // 0 = no baton
	Quantity  uint64
}

type SendMsg struct {
	TokenHash  [32]byte // internal (little-endian) byte order
	Quantities []uint64 // 1 to 19 declared output quantities, in vout order starting at 1
}

// Parse strictly parses an SLP OP_RETURN output script per the spec's
// consensus parsing rules. Any returned error means the script is not a valid
// SLP message (and an SLP-prefixed tx bearing it at vout 0 is invalid SLP).
func Parse(pkScript []byte) (*Msg, error) {
	fields, err := parsePushOnly(pkScript)
	if err != nil {
		return nil, err
	}
	if len(fields) < 3 {
		return nil, fmt.Errorf("slp parse: %d fields, need at least 3", len(fields))
	}
	if !bytes.Equal(fields[0], memo.PrefixSlp) {
		return nil, fmt.Errorf("slp parse: bad lokad id: %x", fields[0])
	}
	if len(fields[1]) != 1 && len(fields[1]) != 2 {
		return nil, fmt.Errorf("slp parse: token type field must be 1 or 2 bytes, got %d", len(fields[1]))
	}
	var msg = new(Msg)
	for _, b := range fields[1] {
		msg.TokenType = msg.TokenType<<8 | uint16(b)
	}
	msg.TxType = memo.SlpType(fields[2])
	var rest = fields[3:]
	switch msg.TxType {
	case memo.SlpTxTypeGenesis:
		msg.Genesis, err = parseGenesis(msg.TokenType, rest)
	case memo.SlpTxTypeMint:
		msg.Mint, err = parseMint(msg.TokenType, rest)
	case memo.SlpTxTypeSend:
		msg.Send, err = parseSend(rest)
	case memo.SlpTxTypeCommit:
		// Recognized but not validated (out of scope)
	default:
		return nil, fmt.Errorf("slp parse: unknown tx type: %q", fields[2])
	}
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func parseGenesis(tokenType uint16, fields [][]byte) (*GenesisMsg, error) {
	if len(fields) != 7 {
		return nil, fmt.Errorf("slp parse: genesis field count %d, expected 7", len(fields))
	}
	var genesis = &GenesisMsg{
		Ticker:  fields[0],
		Name:    fields[1],
		DocUrl:  fields[2],
		DocHash: fields[3],
	}
	if len(genesis.DocHash) != 0 && len(genesis.DocHash) != 32 {
		return nil, fmt.Errorf("slp parse: genesis doc hash must be 0 or 32 bytes, got %d", len(genesis.DocHash))
	}
	if len(fields[4]) != 1 {
		return nil, fmt.Errorf("slp parse: genesis decimals must be 1 byte, got %d", len(fields[4]))
	}
	genesis.Decimals = fields[4][0]
	if genesis.Decimals > 9 {
		return nil, fmt.Errorf("slp parse: genesis decimals %d out of range", genesis.Decimals)
	}
	batonVout, err := parseBatonVout(fields[5])
	if err != nil {
		return nil, err
	}
	genesis.BatonVout = batonVout
	quantity, err := parseQuantity(fields[6])
	if err != nil {
		return nil, err
	}
	genesis.Quantity = quantity
	if tokenType == memo.SlpNftChildTokenType {
		// NFT1 child genesis constraints are consensus parsing rules
		if genesis.Decimals != 0 {
			return nil, fmt.Errorf("slp parse: nft child genesis decimals %d must be 0", genesis.Decimals)
		}
		if genesis.BatonVout != 0 {
			return nil, fmt.Errorf("slp parse: nft child genesis must not have a mint baton")
		}
		if genesis.Quantity != 1 {
			return nil, fmt.Errorf("slp parse: nft child genesis quantity %d must be 1", genesis.Quantity)
		}
	}
	return genesis, nil
}

func parseMint(tokenType uint16, fields [][]byte) (*MintMsg, error) {
	if tokenType == memo.SlpNftChildTokenType {
		return nil, fmt.Errorf("slp parse: nft children cannot be minted")
	}
	if len(fields) != 3 {
		return nil, fmt.Errorf("slp parse: mint field count %d, expected 3", len(fields))
	}
	var mint = new(MintMsg)
	if err := parseTokenHash(fields[0], &mint.TokenHash); err != nil {
		return nil, err
	}
	batonVout, err := parseBatonVout(fields[1])
	if err != nil {
		return nil, err
	}
	mint.BatonVout = batonVout
	quantity, err := parseQuantity(fields[2])
	if err != nil {
		return nil, err
	}
	mint.Quantity = quantity
	return mint, nil
}

func parseSend(fields [][]byte) (*SendMsg, error) {
	if len(fields) < 2 {
		return nil, fmt.Errorf("slp parse: send field count %d, expected at least 2", len(fields))
	}
	if len(fields) > 1+MaxSendOutputs {
		return nil, fmt.Errorf("slp parse: send has %d quantities, max %d", len(fields)-1, MaxSendOutputs)
	}
	var send = new(SendMsg)
	if err := parseTokenHash(fields[0], &send.TokenHash); err != nil {
		return nil, err
	}
	for _, field := range fields[1:] {
		quantity, err := parseQuantity(field)
		if err != nil {
			return nil, err
		}
		send.Quantities = append(send.Quantities, quantity)
	}
	return send, nil
}

func parseTokenHash(field []byte, tokenHash *[32]byte) error {
	if len(field) != 32 {
		return fmt.Errorf("slp parse: token id must be 32 bytes, got %d", len(field))
	}
	// Message carries the token id in display (big-endian) order; store internal order
	for i := 0; i < 32; i++ {
		tokenHash[i] = field[31-i]
	}
	return nil
}

func parseBatonVout(field []byte) (int, error) {
	switch len(field) {
	case 0:
		return 0, nil
	case 1:
		if field[0] < 2 {
			return 0, fmt.Errorf("slp parse: baton vout %d must be at least 2", field[0])
		}
		return int(field[0]), nil
	}
	return 0, fmt.Errorf("slp parse: baton vout field must be 0 or 1 bytes, got %d", len(field))
}

func parseQuantity(field []byte) (uint64, error) {
	if len(field) != 8 {
		return 0, fmt.Errorf("slp parse: quantity must be 8 bytes, got %d", len(field))
	}
	return binary.BigEndian.Uint64(field), nil
}

// HasSlpLokad reports whether the script is an OP_RETURN whose first push is
// the SLP lokad id, regardless of push encoding (the spec permits non-minimal
// pushes). Used to recognize potential SLP outputs without full parsing.
func HasSlpLokad(pkScript []byte) bool {
	if len(pkScript) == 0 || pkScript[0] != opReturn {
		return false
	}
	first, _, err := readPush(pkScript, 1)
	return err == nil && bytes.Equal(first, memo.PrefixSlp)
}

// parsePushOnly parses an OP_RETURN script into its pushed fields, enforcing
// the SLP rule that only data push opcodes 0x01-0x4e are permitted (OP_0,
// OP_1NEGATE, and OP_1-OP_16 are disallowed; non-minimal pushes are allowed).
func parsePushOnly(pkScript []byte) ([][]byte, error) {
	if len(pkScript) == 0 || pkScript[0] != opReturn {
		return nil, fmt.Errorf("slp parse: script does not start with OP_RETURN")
	}
	var fields [][]byte
	var pos = 1
	for pos < len(pkScript) {
		field, newPos, err := readPush(pkScript, pos)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
		pos = newPos
	}
	return fields, nil
}

func readPush(pkScript []byte, pos int) ([]byte, int, error) {
	var op = pkScript[pos]
	var size int
	var lenBytes int
	switch {
	case op >= 0x01 && op < opPushData1:
		size = int(op)
	case op == opPushData1:
		lenBytes = 1
	case op == opPushData2:
		lenBytes = 2
	case op == opPushData4:
		lenBytes = 4
	default:
		return nil, 0, fmt.Errorf("slp parse: disallowed opcode 0x%02x at position %d", op, pos)
	}
	pos++
	if lenBytes > 0 {
		if pos+lenBytes > len(pkScript) {
			return nil, 0, fmt.Errorf("slp parse: truncated push length at position %d", pos)
		}
		for i := 0; i < lenBytes; i++ {
			size |= int(pkScript[pos+i]) << (8 * i)
		}
		pos += lenBytes
	}
	if pos+size > len(pkScript) {
		return nil, 0, fmt.Errorf("slp parse: truncated push data at position %d", pos)
	}
	return pkScript[pos : pos+size], pos + size, nil
}
