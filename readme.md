## User Documentation

https://docs.excelscheduler.field3.systems/

or check out the `docs/content/_index.md` file in this repo.

## Dev Notes

### Addin Development

**Dependencies**

- Node22
- npm
- just (https://github.com/casey/just)

**Quick Setup**

Check out the overview
```
just -l
```

Then go ahead and start dev'ing

```
just addin-dev
```

afterwards open https://localhost:22234/manifest.xml and trust the cert

Files are in the addin folder

**Random**

Neat Place to Find Fabric Icon Names https://uifabricicons.azurewebsites.net/

### Docs

**Dependencies**

- Hugo
- just (https://github.com/casey/just)


Then startup the dev server

```
just docs-dev
```

Files are in the docs folder, content for the pages are in docs/content.

### Release

Can do a quick test of the release build

```
just release-build
```

Make sure the `changelog.md` is updated with the correct release date and
version number.

Then update the version number in `addin/src/taskpane/taskpane.html`.

After everything is merged into main, the pipeline will build and run.


