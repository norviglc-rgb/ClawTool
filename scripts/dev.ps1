param(
    [Parameter(Position = 0)]
    [string]$Task = "check"
)

switch ($Task) {
    "fmt" { go fmt ./...; break }
    "lint" { go vet ./...; break }
    "test" { go test ./...; break }
    "test-unit" { go test ./...; break }
    "test-integration" { go test ./...; break }
    "test-e2e" { go test ./...; break }
    "test-i18n" { go test ./internal/i18n/... ./internal/schema/...; break }
    "build" { go build ./cmd/clawtool; break }
    "build-all" { go build ./...; break }
    "generate" { go test ./...; break }
    "check" {
        go fmt ./...
        go vet ./...
        go test ./...
        break
    }
    "release-dry-run" {
        go build ./...
        break
    }
    default {
        Write-Error "Unsupported task: $Task"
        exit 1
    }
}

