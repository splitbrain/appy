.PHONY: build test

build:
	go build -o appy ./cmd/appy

testdata/test.AppImage:
	@echo "Downloading test.AppImage..."
	@mkdir -p testdata
	@curl -L https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage -o testdata/test.AppImage
	@chmod +x testdata/test.AppImage
	@echo "Downloaded test.AppImage successfully"

test: testdata/test.AppImage
	go test -v ./...
