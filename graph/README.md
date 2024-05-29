# GQL Gen

```bash
go get github.com/99designs/gqlgen@v0.17.20
go generate ./...
```

### Adding field resolvers

Automatic: Move model from `models_gen.go` to `models.go`.
Remove fields from model and regenerate resolvers.

Manual: Add to `gqlgen.yml` (replace with desired model/fields).

```yaml
models:
  Profile:
    fields:
      following:
        resolver: true
      followers:
        resolver: true
```
