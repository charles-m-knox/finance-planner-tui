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
