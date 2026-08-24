package slp_validate

import (
	"testing"

	"github.com/jchavannes/btcd/wire"
)

func topoTx(id byte, parentIds ...byte) Tx {
	var tx = Tx{TxHash: [32]byte{id}}
	for _, parentId := range parentIds {
		tx.Inputs = append(tx.Inputs, &wire.TxIn{
			PreviousOutPoint: wire.OutPoint{Hash: [32]byte{parentId}},
		})
	}
	return tx
}

func checkTopoOrder(t *testing.T, txs []Tx, expectedIds []byte) {
	t.Helper()
	ordered := SortTopological(txs)
	if len(ordered) != len(expectedIds) {
		t.Fatalf("ordered count = %d, want %d", len(ordered), len(expectedIds))
	}
	for i := range ordered {
		if ordered[i].TxHash != [32]byte{expectedIds[i]} {
			var gotIds []byte
			for _, tx := range ordered {
				gotIds = append(gotIds, tx.TxHash[0])
			}
			t.Fatalf("order = %v, want %v", gotIds, expectedIds)
		}
	}
}

func TestSortTopological(t *testing.T) {
	t.Run("already ordered is stable", func(t *testing.T) {
		checkTopoOrder(t, []Tx{topoTx(1), topoTx(2, 1), topoTx(3, 2)}, []byte{1, 2, 3})
	})
	t.Run("child before parent swaps", func(t *testing.T) {
		checkTopoOrder(t, []Tx{topoTx(3, 2), topoTx(2, 1), topoTx(1)}, []byte{1, 2, 3})
	})
	t.Run("independent txs keep original order", func(t *testing.T) {
		checkTopoOrder(t, []Tx{topoTx(5), topoTx(3), topoTx(4)}, []byte{5, 3, 4})
	})
	t.Run("diamond", func(t *testing.T) {
		// 4 spends 2 and 3, both spending 1; 2 and 3 keep their given order
		checkTopoOrder(t, []Tx{topoTx(4, 2, 3), topoTx(3, 1), topoTx(2, 1), topoTx(1)}, []byte{1, 3, 2, 4})
	})
	t.Run("external parents ignored", func(t *testing.T) {
		checkTopoOrder(t, []Tx{topoTx(2, 99), topoTx(1, 2)}, []byte{2, 1})
	})
	t.Run("duplicate parent edges", func(t *testing.T) {
		// two inputs spending different outputs of the same in-chunk parent
		checkTopoOrder(t, []Tx{topoTx(2, 1, 1), topoTx(1)}, []byte{1, 2})
	})
	t.Run("interleaved chains preserve height order between them", func(t *testing.T) {
		// chain A: 1 -> 2, chain B: 3 -> 4, given interleaved with children first
		checkTopoOrder(t, []Tx{topoTx(2, 1), topoTx(4, 3), topoTx(1), topoTx(3)}, []byte{1, 2, 3, 4})
	})
	t.Run("empty", func(t *testing.T) {
		checkTopoOrder(t, nil, nil)
	})
}
