BIN=finance-planner-tui

build:
	CGO_ENABLED=0 go build -v

build-prod:
	CGO_ENABLED=0 go build -v -o $(BIN)-prod -ldflags="-w -s -buildid=" -trimpath

run:
	./$(BIN)

compress-prod:
	rm -f $(BIN)-compressed
	upx --best -o ./$(BIN)-compressed $(BIN)

build-optimized: build-prod compress-prod

build-mac:
	CGO_ENABLED=0 GOARCH=arm64 GOOS=darwin go build -v -o $(BIN)-prod-darwin-arm64 -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-prod-darwin-arm64-compressed
	upx --best -o ./$(BIN)-prod-darwin-arm64-compressed $(BIN)-prod-darwin-arm64

build-win-amd64:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -v -o $(BIN)-prod-win-amd64 -ldflags="-w -s -buildid=" -trimpath
	rm -f $(BIN)-prod-win-amd64-compressed
	upx --best -o ./$(BIN)-prod-win-amd64-compressed $(BIN)-prod-win-amd64

lint:
	golangci-lint run ./...
