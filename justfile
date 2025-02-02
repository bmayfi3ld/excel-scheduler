# for the addin dev server, afterwards open https://localhost:22234/manifest.xml and trust the cert
addin-dev:
  cd addin && npm run dev-server

# for the docs dev server
docs-dev:
  cd docs && hugo server