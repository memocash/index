package memo

import (
	"bytes"
	"encoding/binary"
	"github.com/gcash/bchd/chaincfg/chainhash"
	"github.com/gcash/bchd/txscript"
	"github.com/gcash/bchd/wire"
	"github.com/jchavannes/jgo/jerr"
	"github.com/jchavannes/jgo/jlog"
	"github.com/jchavannes/jgo/jutil"
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
	jlog.Logf("h.BVersion: %x\n", h.BVersion)
	jlog.Logf("h.HashPrevOuts: %x\n", h.HashPrevOuts)
	jlog.Logf("h.HashSequence: %x\n", h.HashSequence)
	jlog.Logf("h.OutPointHash: %x\n", h.OutPointHash)
	jlog.Logf("h.OutPointIndex: %x\n", h.OutPointIndex)
	jlog.Logf("h.InputSubScript: %x\n", h.InputSubScript)
	jlog.Logf("h.BAmount: %x\n", h.BAmount)
	jlog.Logf("h.BSequence: %x\n", h.BSequence)
	jlog.Logf("h.HashOutputs: %x\n", h.HashOutputs)
	jlog.Logf("h.BLockTime: %x\n", h.BLockTime)
	jlog.Logf("h.BHashType: %x\n", h.BHashType)
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
		return jerr.Newf("idx %d but %d TxIn", idx, len(h.Tx.TxIn))
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
	vm, err := txscript.NewEngine(pkScript, &h.Tx, idx, flags, nil, nil, amt)
	if err != nil {
		return jerr.Get("error new pk script engine", err)
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
