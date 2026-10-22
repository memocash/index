# GQL Gen

```bash
# Run the generator from the repo root. gqlgen resolves the schema glob and
# output paths in gqlgen.yml relative to the working directory, so it must run
# inside graph/; the subshell keeps your shell in the repo root.
(cd graph && GOFLAGS=-mod=mod go run github.com/99designs/gqlgen generate --config gqlgen.yml)

# Only graph/generated/generated.go should change. graph/resolver/ and
# graph/model/models.go are tracked, hand-maintained sources: if git status
# shows them modified or new files appeared under graph/, inspect before
# committing. Never rm -rf graph/resolver.
git status --short graph/
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
