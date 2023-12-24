.PHONY=build

BUILDDIR=build
BIN=$(BUILDDIR)/finance-planner-tui

build-dev:
	CGO_ENABLED=0 go build -v

mkbuilddir:
	mkdir -p $(BUILDDIR)

build-prod: mkbuilddir
	CGO_ENABLED=0 go build -v -o $(BIN)-prod -ldflags="-w -s -buildid=" -trimpath

run:
	./$(BIN)

lint:
	golangci-lint run ./...

compress-prod: mkbuilddir
	rm -f $(BIN)-compressed
	upx --best -o ./$(BIN)-compressed $(BIN)

build-optimized: mkbuilddir build-prod compress-prod

build-mac-arm64: mkbuilddir
	CGO_ENABLED=0 GOARCH=arm64 GOOS=darwin go build -v -o $(BIN)-prod-darwin-arm64 -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-prod-darwin-arm64-compressed
# note for mac m1 - this seems to taint the binary, it doesn't work;
# you'll probably have to do without upx for now
# upx --best -o ./$(BIN)-prod-darwin-arm64-compressed $(BIN)-prod-darwin-arm64

build-mac-amd64: mkbuilddir
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -v -o $(BIN)-prod-darwin-amd64 -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-prod-darwin-arm64-compressed
	upx --best -o ./$(BIN)-prod-darwin-arm64-compressed $(BIN)-prod-darwin-arm64

build-win-amd64: mkbuilddir
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -v -o $(BIN)-prod-win-amd64 -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-prod-win-amd64-compressed
	upx --best -o ./$(BIN)-prod-win-amd64-compressed $(BIN)-prod-win-amd64

build-linux-arm64: mkbuilddir
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux go build -v -o $(BIN)-prod-linux-arm64 -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-prod-linux-arm64-compressed
	upx --best -o ./$(BIN)-prod-linux-arm64-compressed $(BIN)-prod-linux-arm64

build-linux-amd64: mkbuilddir
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -v -o $(BIN)-prod-linux-amd64 -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-prod-linux-amd64-compressed
	upx --best -o ./$(BIN)-prod-linux-amd64-compressed $(BIN)-prod-linux-amd64
