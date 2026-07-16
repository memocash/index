package maint

import (
	"bufio"
	"context"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"

	"github.com/jchavannes/jgo/jutil"
	"github.com/memocash/index/node/act/maint"
	"github.com/spf13/cobra"
)

var populateOpReturnsCmd = &cobra.Command{
	Use:   "populate-op-returns",
	Short: "Run the op_return saver over stored Index txs for tx hashes read from stdin",
	Long: "Reads big-endian hex tx hashes (one per line) from stdin, reconstructs each tx from " +
		"data already stored in the Index (version, inputs, outputs, seen), and runs it through the " +
		"op_return saver. Populates op_return-derived data (e.g. new memo item types) without a BCH " +
		"node scan. Feed hashes from MySQL, e.g.:\n" +
		"  infra-tool mysql local \"select hex(tx_hash) from link_accepts\" | ./index maint populate-op-returns",
	Run: func(c *cobra.Command, args []string) {
		verbose, _ := c.Flags().GetBool(FlagVerbose)
		hashes, err := readTxHashes(os.Stdin)
		if err != nil {
			log.Fatalf("error reading tx hashes; %v", err)
		}
		log.Printf("Read %d tx hashes\n", len(hashes))
		populate := maint.NewPopulateOpReturn(context.Background(), verbose)
		result, err := populate.Run(hashes)
		if err != nil {
			log.Fatalf("error running populate op returns; %v", err)
		}
		for _, txHash := range result.Missing {
			log.Printf("missing/skipped (not fully in index): %s\n", hashString(txHash))
		}
		log.Printf("Done. Processed %d txs, %d missing/skipped.\n", result.Processed, len(result.Missing))
	},
}

// readTxHashes reads internal-order hex tx hashes, one per line, skipping blanks and
// non-hash lines (e.g. a column header). This is the order stored by the Index chain
// topics and returned by MySQL's hex(tx_hash) (the reverse of the explorer txid).
func readTxHashes(r io.Reader) ([][32]byte, error) {
	var hashes [][32]byte
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) != 64 {
			continue
		}
		raw, err := hex.DecodeString(line)
		if err != nil {
			continue
		}
		var txHash [32]byte
		copy(txHash[:], raw)
		hashes = append(hashes, txHash)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return hashes, nil
}

// hashString renders an internal-order tx hash as the explorer txid (big-endian).
func hashString(txHash [32]byte) string {
	return hex.EncodeToString(jutil.ByteReverse(txHash[:]))
}
