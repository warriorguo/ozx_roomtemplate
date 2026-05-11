# Testing Guide (local-client branch)

This branch has no database dependency, so there are no integration tests —
only Go unit tests against the in-process packages.

## Layout

| Package | Tests | Notes |
|---|---|---|
| `internal/validate/` | `TestValidateTemplate_*` | Structure + logical rules |
| `internal/generate/` | Generators, stage rules, mainpath, static placement invariants | Some probabilistic tests can be flaky |
| `internal/http/` | Handler tests with `MockStore` | Mocks the `store.Store` interface |
| `internal/store/` | (none yet) | Filesystem store tests land in **ORT-66** |

## Running tests

```bash
make test-unit             # everything under internal/...
make test-coverage         # generates coverage.html
make test-bench            # benchmarks

go test -v ./internal/validate/
go test -v ./internal/http/
go test -v ./internal/generate/
```

## Mocking the store

`internal/http/handlers_test.go` defines `MockStore`, a `testify/mock`
implementation of `store.Store`. New handler tests should reuse it:

```go
mockStore := &MockStore{}
mockStore.On("Get", mock.Anything, id).Return(&template, nil)
handler := NewTemplateHandler(mockStore, zap.NewNop())
```

## Coverage

```bash
make test-coverage
open coverage.html
```
