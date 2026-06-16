# for the addin dev server, afterwards open https://localhost:22234/manifest.xml and trust the cert
addin-dev:
  npm install -g yo generator-office
  npm install
  cd addin && npm run dev-server

# to test the release build
release-build:
  cd addin && npm run build

# for the docs dev server
docs-dev:
  cd docs && hugo server

# build the scheduler CLI to ./bin (bin/ is git-ignored)
cli-build:
  go build -o bin/scheduler ./cmd/scheduler

# run all unit tests and the linter (Go)
validate:
  go test ./...
  go vet ./...
  @command -v govulncheck >/dev/null 2>&1 || { echo "govulncheck not found — install it: go install golang.org/x/vuln/cmd/govulncheck@latest"; exit 1; }
  govulncheck ./...
  @command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found — install it: https://golangci-lint.run/welcome/install/"; exit 1; }
  golangci-lint run

# point git at the committed hooks so `just validate` runs before each commit
setup-hooks:
  git config core.hooksPath .githooks