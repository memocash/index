package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/hex"

	"github.com/jchavannes/jgo/jerr"
	"github.com/memocash/index/admin/graph/generated"
	"github.com/memocash/index/ref/broadcast/broadcast_client"
)

// Broadcast is the resolver for the broadcast field.
func (r *mutationResolver) Broadcast(ctx context.Context, raw string) (bool, error) {
	rawBytes, err := hex.DecodeString(raw)
	if err != nil {
		return false, jerr.Get("error decoding raw tx for graphql broadcast", err)
	}
	client := broadcast_client.NewBroadcast()
	if err := client.Broadcast(ctx, rawBytes); err != nil {
		return false, jerr.Get("error broadcasting tx for graphql", err)
	}
	return true, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

type mutationResolver struct{ *Resolver }
