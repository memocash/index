# Dataloader

Create new:

```bash
cd github.com/memocash/index/admin/graph/dataloader
go get github.com/vektah/dataloaden
go run github.com/vektah/dataloaden TxLostLoader string *github.com/memocash/index/admin/graph/model.TxLost
go run github.com/vektah/dataloaden TxSuspectLoader string *github.com/memocash/index/admin/graph/model.TxSuspect
go run github.com/vektah/dataloaden BlockLoader string []*github.com/memocash/index/admin/graph/model.Block
go run github.com/vektah/dataloaden TxSeenLoader string *github.com/memocash/index/admin/graph/model.Date
go run github.com/vektah/dataloaden TxOutputLoader github.com/memocash/index/admin/graph/model.HashIndex *github.com/memocash/index/admin/graph/model.TxOutput
go get github.com/vektah/dataloaden@none
```

# GQL Gen

```bash
go get github.com/99designs/gqlgen
go generate ./...
```
