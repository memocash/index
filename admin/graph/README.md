# Dataloader

Create new:

```bash
cd github.com/memocash/server/admin/graph/dataloader
go get github.com/vektah/dataloaden
go run github.com/vektah/dataloaden TxLostLoader string *github.com/memocash/server/admin/graph/model.TxLost
go run github.com/vektah/dataloaden TxSuspectLoader string *github.com/memocash/server/admin/graph/model.TxSuspect
go run github.com/vektah/dataloaden BlockLoader string []*github.com/memocash/server/admin/graph/model.Block
go get github.com/vektah/dataloaden@none
```
