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