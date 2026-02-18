package test_tx_test

import (
	"fmt"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/bitcoin/tx/gen"
	"github.com/memocash/index/ref/bitcoin/tx/hs"
	"github.com/memocash/index/ref/bitcoin/util/testing/test_tx"
	"github.com/memocash/index/ref/bitcoin/wallet"
	"log"
)

func printTx(value int64) {
	test_tx.ResetUTXOIndex()
	memoTx, err := gen.TxUnsigned(gen.TxRequest{
		Getter: gen.GetWrapper(&test_tx.TestGetter{UTXOs: []memo.UTXO{test_tx.GetUTXO(value)}}, test_tx.Address1pkHash),
		Change: wallet.GetChange(test_tx.Address1),
	})
	if err != nil {
		log.Fatalf("fatal error getting unsigned; %v", err)
	}
	fmt.Println(hs.GetTxString(memoTx.GetHash()))
	//fmt.Printf("value: %d, tx_hash: %s\n", value, hs.GetTxString(memoTx.GetHash()))
	//parse.GetTxInfo(memoTx).Print()
}

func Example_emptyTx1000() {
	printTx(1000)
	// Output: f35f64bba6e99fe94b3fbed0c75f47f4d1ba6791d4f2f216c9fdb2e30810d19b
}

func Example_emptyTx1001() {
	printTx(1001)
	// Output: 3fe109fb374c7bbbccf12a77e98d81d955393284e6c338f705163c4ee9c18189
}

func Example_emptyTx1002() {
	printTx(1002)
	// Output: e136fef169cbee7ad43d3ccfd27c2c8ff421184d3e3a4990b53f5f6ee52389b0
}

func Example_emptyTx1003() {
	printTx(1003)
	// Output: e3c085463385f646b9a495551b316d79b7d445f742f35da0eb611aab78017903
}

func Example_emptyTx1004() {
	printTx(1004)
	// Output: 3f9d0cae29b93e40d7fa8f0f38f5c31804a862ccfa03c97b400eb143cab4965c
}

func Example_emptyTx1005() {
	printTx(1005)
	// Output: 0e7aeed9ea63921dd26d0d7c6f21bd7de9c9baba4697d2a0c2d3802253f40fa2
}

func Example_emptyTx1006() {
	printTx(1006)
	// Output: bf4bc03df9cba52e849a9fac87a7cd08afbedf9287edccd5d5dc7cd08b05b51b
}

func Example_emptyTx1007() {
	printTx(1007)
	// Output: bb0113745b55ffd22e6776ad02e8152fd298094a0889759419aed2ea26f68826
}

func Example_emptyTx1008() {
	printTx(1008)
	// Output: 90d9f0394685ff762d9f94c191fa61995c261ebc4651bde1973fe67ad04014fb
}

func Example_emptyTx1009() {
	printTx(1009)
	// Output: 94a3569c2782a3caf1ab2e9433bcc7d9c30483166448a95c993cebef6bd49c99
}
