package slp_validate

import (
	"context"
	"encoding/binary"
	"testing"

	"github.com/jchavannes/btcd/chaincfg/chainhash"
	"github.com/jchavannes/btcd/wire"
	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/db/item/chain"
	"github.com/memocash/index/db/item/db"
	item_slp "github.com/memocash/index/db/item/slp"
	"github.com/memocash/index/ref/bitcoin/memo"
)

func push(data []byte) []byte {
	return append([]byte{byte(len(data))}, data...)
}

// slpSendScript builds a strict-parseable type-1 SEND vout-0 script; token
// hash is given in internal (little-endian) order like TxHash.
func slpSendScript(tokenHash [32]byte, amounts ...uint64) []byte {
	script := []byte{0x6a}
	script = append(script, push(memo.PrefixSlp)...)
	script = append(script, push([]byte{memo.SlpDefaultTokenType})...)
	script = append(script, push([]byte(memo.SlpTxTypeSend))...)
	script = append(script, push(jutil.ByteReverse(tokenHash[:]))...)
	for _, amount := range amounts {
		var quantity [8]byte
		binary.BigEndian.PutUint64(quantity[:], amount)
		script = append(script, push(quantity[:])...)
	}
	return script
}

func hash32(b byte) (h [32]byte) {
	for i := range h {
		h[i] = b
	}
	return
}

// resolveFakes stubs every backing-read seam over an in-memory dataset,
// counting fetches per key so tests can pin the resolve cache's contract.
type resolveFakes struct {
	validities map[[32]byte]*item_slp.Validity
	outputs    map[outPoint]*item_slp.Output
	geneses    map[[32]byte]*item_slp.Genesis
	vouts      map[[32]byte][]byte // vout-0 lock scripts (chain rows)

	validityFetches map[[32]byte]int
	outputFetches   map[outPoint]int
	batonFetches    map[outPoint]int
	genesisFetches  map[[32]byte]int
	voutFetches     map[[32]byte]int
	saved           []db.Object
	totalKeys       int // every key requested from any seam
}

func newResolveFakes() *resolveFakes {
	return &resolveFakes{
		validities:      make(map[[32]byte]*item_slp.Validity),
		outputs:         make(map[outPoint]*item_slp.Output),
		geneses:         make(map[[32]byte]*item_slp.Genesis),
		vouts:           make(map[[32]byte][]byte),
		validityFetches: make(map[[32]byte]int),
		outputFetches:   make(map[outPoint]int),
		batonFetches:    make(map[outPoint]int),
		genesisFetches:  make(map[[32]byte]int),
		voutFetches:     make(map[[32]byte]int),
	}
}

func (f *resolveFakes) install(t *testing.T) {
	t.Helper()
	origValidities, origOutputs, origBatons := getValidities, getSlpOutputs, getSlpBatons
	origGeneses, origVouts, origSave := getSlpGeneses, getVoutOutputs, saveObjects
	t.Cleanup(func() {
		getValidities, getSlpOutputs, getSlpBatons = origValidities, origOutputs, origBatons
		getSlpGeneses, getVoutOutputs, saveObjects = origGeneses, origVouts, origSave
	})
	getValidities = func(_ context.Context, txHashes [][32]byte) ([]*item_slp.Validity, error) {
		var found []*item_slp.Validity
		for _, txHash := range txHashes {
			f.validityFetches[txHash]++
			f.totalKeys++
			if validity, ok := f.validities[txHash]; ok {
				found = append(found, validity)
			}
		}
		return found, nil
	}
	getSlpOutputs = func(_ context.Context, outs []memo.Out) ([]*item_slp.Output, error) {
		var found []*item_slp.Output
		for _, out := range outs {
			var op = outPoint{index: out.Index}
			copy(op.hash[:], out.TxHash)
			f.outputFetches[op]++
			f.totalKeys++
			if output, ok := f.outputs[op]; ok {
				found = append(found, output)
			}
		}
		return found, nil
	}
	getSlpBatons = func(_ context.Context, outs []memo.Out) ([]*item_slp.Baton, error) {
		for _, out := range outs {
			var op = outPoint{index: out.Index}
			copy(op.hash[:], out.TxHash)
			f.batonFetches[op]++
			f.totalKeys++
		}
		return nil, nil
	}
	getSlpGeneses = func(_ context.Context, tokenHashes [][32]byte) ([]*item_slp.Genesis, error) {
		var found []*item_slp.Genesis
		for _, tokenHash := range tokenHashes {
			f.genesisFetches[tokenHash]++
			f.totalKeys++
			if genesis, ok := f.geneses[tokenHash]; ok {
				found = append(found, genesis)
			}
		}
		return found, nil
	}
	getVoutOutputs = func(_ context.Context, outs []memo.Out) ([]*chain.TxOutput, error) {
		var found []*chain.TxOutput
		for _, out := range outs {
			var txHash [32]byte
			copy(txHash[:], out.TxHash)
			f.voutFetches[txHash]++
			f.totalKeys++
			if script, ok := f.vouts[txHash]; ok {
				found = append(found, &chain.TxOutput{TxHash: txHash, Index: out.Index, LockScript: script})
			}
		}
		return found, nil
	}
	saveObjects = func(objects []db.Object) error {
		f.saved = append(f.saved, objects...)
		return nil
	}
}

// TestResolveCacheReadContract pins both sides of the resolve cache's
// contract across the rounds of one shared cache (the fan-in re-validation
// pattern): final facts — a decided parent's verdict and row state, a found
// genesis row, known vout-0 answers — are fetched exactly once, while genuine
// unknowns — an undecided parent's verdict and its absent rows — are fetched
// again every round until they decide.
func TestResolveCacheReadContract(t *testing.T) {
	var tokenT = hash32(0x70)
	var parentD = hash32(0xd0) // decided valid, transcribed
	var parentU = hash32(0xa0) // undecided SLP (vout-0 lokad, no rows yet)
	var parentN = hash32(0xb0) // not SLP (plain vout-0 script)
	var spenderS = hash32(0x50)

	fakes := newResolveFakes()
	fakes.install(t)
	fakes.validities[parentD] = &item_slp.Validity{TxHash: parentD, Status: item_slp.ValidityStatusValid}
	fakes.outputs[outPoint{hash: parentD, index: 1}] = &item_slp.Output{
		TxHash: parentD, Index: 1, TokenHash: tokenT, Quantity: 100}
	fakes.geneses[tokenT] = &item_slp.Genesis{TxHash: tokenT, TokenType: memo.SlpDefaultTokenType}
	fakes.vouts[parentU] = slpSendScript(tokenT, 100)
	fakes.vouts[parentN] = []byte{0x76, 0xa9, 0x14} // p2pkh-ish, no lokad

	// S declares 150 of token T from inputs D:1 (100, valid), U:1 (unknown),
	// N:0 (not slp): pending until U decides and its row lands
	var spendTx = Tx{
		TxHash: spenderS,
		Inputs: []*wire.TxIn{
			wire.NewTxIn(wire.NewOutPoint((*chainhash.Hash)(&parentD), 1), nil),
			wire.NewTxIn(wire.NewOutPoint((*chainhash.Hash)(&parentU), 1), nil),
			wire.NewTxIn(wire.NewOutPoint((*chainhash.Hash)(&parentN), 0), nil),
		},
		Outputs: []*wire.TxOut{{PkScript: slpSendScript(tokenT, 150)}},
	}

	var cache = newResolveCache()
	var roundKeys []int
	var run = func(name string) *Result {
		t.Helper()
		before := fakes.totalKeys
		result, err := validateTxsCached(context.Background(), []Tx{spendTx}, cache)
		if err != nil {
			t.Fatalf("%s: error validating; %v", name, err)
		}
		roundKeys = append(roundKeys, fakes.totalKeys-before)
		return result
	}

	if result := run("round 1"); result.Pending != 1 || result.Decided() != 0 {
		t.Fatalf("round 1: expected pending, got %+v", result)
	}
	if result := run("round 2"); result.Pending != 1 || result.Decided() != 0 {
		t.Fatalf("round 2: expected still pending, got %+v", result)
	}

	// U decides valid and its transcription row lands; round 3 must re-read
	// exactly the unknowns, find them, and decide S
	fakes.validities[parentU] = &item_slp.Validity{TxHash: parentU, Status: item_slp.ValidityStatusValid}
	fakes.outputs[outPoint{hash: parentU, index: 1}] = &item_slp.Output{
		TxHash: parentU, Index: 1, TokenHash: tokenT, Quantity: 100}
	if result := run("round 3"); result.Valid != 1 {
		t.Fatalf("round 3: expected valid, got %+v", result)
	}
	if len(fakes.saved) != 1 {
		t.Fatalf("expected 1 saved verdict, got %d", len(fakes.saved))
	}

	// Final facts: fetched exactly once across all three rounds
	for name, got := range map[string]int{
		"decided parent D verdict":  fakes.validityFetches[parentD],
		"decided parent D:1 output": fakes.outputFetches[outPoint{hash: parentD, index: 1}],
		"decided parent D:1 baton":  fakes.batonFetches[outPoint{hash: parentD, index: 1}],
		"not-slp parent N:0 output": fakes.outputFetches[outPoint{hash: parentN, index: 0}],
		"genesis row for token T":   fakes.genesisFetches[tokenT],
		"vout-0 answer for U":       fakes.voutFetches[parentU],
		"vout-0 answer for N":       fakes.voutFetches[parentN],
	} {
		if got != 1 {
			t.Errorf("%s: fetched %d times, expected exactly 1", name, got)
		}
	}
	// Genuine unknowns: re-fetched every round until decided
	if got := fakes.validityFetches[parentU]; got != 3 {
		t.Errorf("undecided parent U verdict: fetched %d times, expected 3 (every round)", got)
	}
	if got := fakes.outputFetches[outPoint{hash: parentU, index: 1}]; got != 3 {
		t.Errorf("undecided parent U:1 output: fetched %d times, expected 3 (absent rows of undecided parents re-read)", got)
	}
	if got := fakes.validityFetches[parentN]; got != 3 {
		t.Errorf("verdict-less parent N: fetched %d times, expected 3 (verdict absence is never cached)", got)
	}
	// The spender's own existing-verdict pre-check is uncached by design
	if got := fakes.validityFetches[spenderS]; got != 3 {
		t.Errorf("spender pre-check: fetched %d times, expected 3", got)
	}
	// Fan-in shape: later rounds do strictly less backing work
	if roundKeys[1] >= roundKeys[0] {
		t.Errorf("round 2 fetched %d keys, expected fewer than round 1's %d", roundKeys[1], roundKeys[0])
	}
	t.Logf("backing keys fetched per round: %v", roundKeys)
}
