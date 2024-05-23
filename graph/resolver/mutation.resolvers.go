package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/memocash/index/graph/generated"
	"github.com/memocash/index/ref/bitcoin/memo"
	"github.com/memocash/index/ref/broadcast/broadcast_client"
)

// Broadcast is the resolver for the broadcast field.
func (r *mutationResolver) Broadcast(ctx context.Context, raw string) (bool, error) {
	rawBytes, err := hex.DecodeString(raw)
	if err != nil {
		return false, fmt.Errorf("error decoding raw tx for graphql broadcast; %w", err)
	}
	client := broadcast_client.NewBroadcast()
	msgTx, err := memo.GetMsgFromRaw(rawBytes)
	if err != nil {
		return false, fmt.Errorf("error getting msg tx from raw for broadcast mutation; %w", err)
	}
	log.Printf("Broadcasting tx: %s\n", msgTx.TxHash())
	if err := client.Broadcast(ctx, rawBytes); err != nil {
		log.Printf("Broadcast tx failed: %s\n", msgTx.TxHash())
		return false, fmt.Errorf("error broadcasting tx for graphql; %w", err)
	}
	return true, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
