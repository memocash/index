package memo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/txscript"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jutil"
	"log"
)

type SigHash struct {
	BVersion       [4]byte
	HashPrevOuts   []byte
	HashSequence   []byte
	OutPointHash   []byte
	OutPointIndex  [4]byte
	InputSubScript []byte
	BAmount        [8]byte
	BSequence      [4]byte
	HashOutputs    []byte
	BLockTime      [4]byte
	BHashType      [4]byte
}

func (h SigHash) Get() []byte {
	return chainhash.DoubleHashB(h.GetCombined())
}

func (h SigHash) GetCombined() []byte {
	return jutil.CombineBytes(
		h.GetPrefix(),
		h.GetSuffix(),
	)
}

func (h SigHash) OutputEach() {
	log.Printf("h.BVersion: %x\n", h.BVersion)
	log.Printf("h.HashPrevOuts: %x\n", h.HashPrevOuts)
	log.Printf("h.HashSequence: %x\n", h.HashSequence)
	log.Printf("h.OutPointHash: %x\n", h.OutPointHash)
	log.Printf("h.OutPointIndex: %x\n", h.OutPointIndex)
	log.Printf("h.InputSubScript: %x\n", h.InputSubScript)
	log.Printf("h.BAmount: %x\n", h.BAmount)
	log.Printf("h.BSequence: %x\n", h.BSequence)
	log.Printf("h.HashOutputs: %x\n", h.HashOutputs)
	log.Printf("h.BLockTime: %x\n", h.BLockTime)
	log.Printf("h.BHashType: %x\n", h.BHashType)
}

func (h SigHash) GetPrefix() []byte {
	return jutil.CombineBytes(
		h.BVersion[:],
		h.HashPrevOuts,
		h.HashSequence,
		h.OutPointHash,
		h.OutPointIndex[:],
		h.InputSubScript,
		h.BAmount[:],
		h.BSequence[:],
	)
}

func (h SigHash) GetSuffix() []byte {
	return jutil.CombineBytes(
		h.HashOutputs,
		h.BLockTime[:],
		h.BHashType[:],
	)
}

type ScriptHasher struct {
	Tx          wire.MsgTx
	txSigHashes txscript.TxSigHashes
	SigHashes   []*SigHash
}

func (h *ScriptHasher) Add(pkScript []byte, hashType txscript.SigHashType, idx int, amt int64) error {
	if idx >= len(h.Tx.TxIn) {
		return fmt.Errorf("idx %d but %d TxIn", idx, len(h.Tx.TxIn))
	}
	var sigHash = &SigHash{}
	var zeroHash chainhash.Hash

	binary.LittleEndian.PutUint32(sigHash.BVersion[:], uint32(h.Tx.Version))

	if hashType&txscript.SigHashAnyOneCanPay == 0 {
		sigHash.HashPrevOuts = h.txSigHashes.HashPrevOuts[:]
	} else {
		sigHash.HashPrevOuts = zeroHash[:]
	}

	if hashType&txscript.SigHashAnyOneCanPay == 0 &&
		hashType&txscript.SigHashMask != txscript.SigHashSingle &&
		hashType&txscript.SigHashMask != txscript.SigHashNone {
		sigHash.HashSequence = h.txSigHashes.HashSequence[:]
	} else {
		sigHash.HashSequence = zeroHash[:]
	}

	sigHash.OutPointHash = h.Tx.TxIn[idx].PreviousOutPoint.Hash[:]
	binary.LittleEndian.PutUint32(sigHash.OutPointIndex[:], h.Tx.TxIn[idx].PreviousOutPoint.Index)

	flags := txscript.StandardVerifyFlags
	vm, err := txscript.NewEngine(pkScript, &h.Tx, idx, flags, nil, amt)
	if err != nil {
		return fmt.Errorf("error new pk script engine; %w", err)
	}
	subScript := vm.SubScript()
	var w bytes.Buffer
	wire.WriteVarBytes(&w, 0, subScript)
	sigHash.InputSubScript = w.Bytes()

	binary.LittleEndian.PutUint64(sigHash.BAmount[:], uint64(amt))
	binary.LittleEndian.PutUint32(sigHash.BSequence[:], h.Tx.TxIn[idx].Sequence)

	if hashType&txscript.SigHashMask != txscript.SigHashSingle &&
		hashType&txscript.SigHashMask != txscript.SigHashNone {
		sigHash.HashOutputs = h.txSigHashes.HashOutputs[:]
	} else if hashType&txscript.SigHashMask == txscript.SigHashSingle && idx < len(h.Tx.TxOut) {
		var b bytes.Buffer
		wire.WriteTxOut(&b, 0, 0, h.Tx.TxOut[idx])
		sigHash.HashOutputs = chainhash.DoubleHashB(b.Bytes())
	} else {
		sigHash.HashOutputs = zeroHash[:]
	}

	binary.LittleEndian.PutUint32(sigHash.BLockTime[:], h.Tx.LockTime)
	binary.LittleEndian.PutUint32(sigHash.BHashType[:], uint32(hashType))
	h.SigHashes = append(h.SigHashes, sigHash)
	return nil
}

func NewScriptHasher(tx *wire.MsgTx) *ScriptHasher {
	sigHashes := txscript.NewTxSigHashes(tx)
	return &ScriptHasher{
		Tx:          *tx,
		txSigHashes: *sigHashes,
	}
}
