# Contributing

Keep package definitions declarative and platform-specific commands reviewable. New recipes should include detection, installation, and a test that does not mutate the host. Avoid shell commands that download and execute remote code.

Before opening a pull request:

```sh
gofmt -w .
go test ./...
go vet ./...
```

Commit messages use an imperative subject with a small scope, for example `catalog: add Bun recipe`.
