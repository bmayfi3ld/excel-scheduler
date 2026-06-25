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

# build the quilt CLI to ./bin (bin/ is git-ignored)
cli-build:
  go build -o bin/quilt ./cmd/quilt

# build the quilt CLI to ./bin (bin/ is git-ignored)
cli-install:
  go install ./cmd/quilt

# local goreleaser dry-run — builds all 6 target archives into ./dist
release-snapshot:
  @command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not found — install it: https://goreleaser.com/install/"; exit 1; }
  goreleaser release --snapshot --clean

# zip the Claude Desktop Extension (.dxt): binary + docs + views + icon
dxt: cli-build
  rm -rf dist/dxt && mkdir -p dist/dxt/docs dist/dxt/live-views
  cp bin/quilt dist/dxt/quilt
  cp packaging/dxt/manifest.json packaging/dxt/icon.png packaging/dxt/icon.svg dist/dxt/
  cp docs/content/docs/*.md dist/dxt/docs/
  cp packaging/live-views/*.html dist/dxt/live-views/
  cd dist/dxt && python3 -c "import shutil,os; shutil.make_archive('../quilt','zip','.'); os.replace('../quilt.zip','../quilt.dxt')"
  @echo "built dist/quilt.dxt"

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
