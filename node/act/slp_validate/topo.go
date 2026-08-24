package slp_validate

import "container/heap"

// SortTopological orders txs so every in-slice parent precedes its spenders,
// preserving the given order among independent txs (smallest original index
// first). Needed because CTOR blocks (BCH ≥ Nov 2018) order txs by txid, not
// topologically, so a same-block parent can sort after its child. Cycles are
// impossible in a tx spend graph; any leftovers are appended defensively.
func SortTopological(txs []Tx) []Tx {
	var indexByHash = make(map[[32]byte]int, len(txs))
	for i, tx := range txs {
		indexByHash[tx.TxHash] = i
	}
	var children = make(map[int][]int)
	var indegree = make([]int, len(txs))
	for i, tx := range txs {
		var parents = make(map[int]bool)
		for _, txIn := range tx.Inputs {
			parentIndex, ok := indexByHash[txIn.PreviousOutPoint.Hash]
			if !ok || parentIndex == i || parents[parentIndex] {
				continue
			}
			parents[parentIndex] = true
			children[parentIndex] = append(children[parentIndex], i)
			indegree[i]++
		}
	}
	var ready = new(indexHeap)
	for i := range txs {
		if indegree[i] == 0 {
			heap.Push(ready, i)
		}
	}
	var ordered = make([]Tx, 0, len(txs))
	var emitted = make([]bool, len(txs))
	for ready.Len() > 0 {
		i := heap.Pop(ready).(int)
		ordered = append(ordered, txs[i])
		emitted[i] = true
		for _, child := range children[i] {
			indegree[child]--
			if indegree[child] == 0 {
				heap.Push(ready, child)
			}
		}
	}
	for i := range txs {
		if !emitted[i] {
			ordered = append(ordered, txs[i])
		}
	}
	return ordered
}

type indexHeap []int

func (h indexHeap) Len() int            { return len(h) }
func (h indexHeap) Less(i, j int) bool  { return h[i] < h[j] }
func (h indexHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *indexHeap) Push(x interface{}) { *h = append(*h, x.(int)) }
func (h *indexHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}
