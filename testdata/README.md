# Test Data

This directory contains test AppImages used for the test suite.

## test.AppImage

The `test.AppImage` file is appimagetool, which is **automatically downloaded** when running `make test` if not already present.

- **Application:** appimagetool
- **Purpose:** Tool to generate AppImages from AppDirs
- **Size:** ~14.4 MB
- **Compression:** XZ
- **Source:** https://github.com/AppImage/appimagetool/releases

To manually download:

```bash
curl -L https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage -o testdata/test.AppImage
chmod +x testdata/test.AppImage
```

## Running Tests

```bash
# Run all tests (automatically downloads test.AppImage if missing)
make test

# Or with go directly (requires test.AppImage to be present)
go test -v ./...
```
