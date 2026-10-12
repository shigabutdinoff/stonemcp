SNAPSHOT   := $(if $(shell git tag --points-at HEAD 2>/dev/null),,--snapshot)
GORELEASER := go run github.com/goreleaser/goreleaser/v2@v2.18.2
PLATFORMS  := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64

.PHONY: build release lint fmt fakegame

build:
	$(GORELEASER) build --clean --single-target --auto-snapshot $(SNAPSHOT) -o stonemcp$(shell go env GOEXE)

release:
	$(GORELEASER) build --clean --auto-snapshot $(SNAPSHOT)

lint:
	go run ./cmd/staticlint ./...

fmt:
	go tool goimports -local github.com/shigabutdinoff/stonemcp -w .

fakegame:
	mkdir -p internal/fakegame/bin
	for p in $(PLATFORMS); do \
		os=$${p%%/*}; arch=$${p#*/}; ext=; \
		if [ $$os = windows ]; then ext=.exe; fi; \
		(cd internal/fakegame/src && GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build \
			-trimpath -buildvcs=false -ldflags "-s -w -buildid=" \
			-o $(CURDIR)/internal/fakegame/bin/fakegame-$$os-$$arch$$ext ./cmd/fakegame) || exit 1; \
	done
