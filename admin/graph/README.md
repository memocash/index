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
go run github.com/vektah/dataloaden TxRawLoader string string
go run github.com/vektah/dataloaden ProfileLoader string *github.com/memocash/index/admin/graph/model.Profile

go get github.com/vektah/dataloaden@none
go mod tidy
```

# GQL Gen

```bash
go get github.com/99designs/gqlgen
go generate ./...
```

### Adding field resolvers

Add to `gqlgen.yml` (replace with desired model/fields):

```yaml
models:
  Profile:
    fields:
      following:
        resolver: true
      followers:
        resolver: true
```
