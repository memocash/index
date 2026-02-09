# GQL Gen

```bash
# Run generator from repo root
GOFLAGS=-mod=mod go run github.com/99designs/gqlgen generate --config graph/gqlgen.yml

# Clean up extra generated files (only graph/generated/generated.go is needed)
rm -f graph/model/models_gen.go
rm -rf resolver/
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
