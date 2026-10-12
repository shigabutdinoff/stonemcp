VERSION_RE := (0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(\.[1-9][0-9]*)?(-(alpha|beta|rc)\.[1-9][0-9]*)?
STAGE_KEY := BEGIN{r["alpha"]=1;r["beta"]=2;r["rc"]=3} {p=0;s=$$3;n=$$4} $$3~/^[0-9]/{p=$$3;s=$$4;n=$$5} {print $$1,$$2,p,(s in r?r[s]:4),n+0,$$0}
TAG       := $(shell git tag --points-at HEAD 2>/dev/null | grep -Ex '$(VERSION_RE)' \
               | awk -F'[.-]' '$(STAGE_KEY)' | sort -k1,1n -k2,2n -k3,3n -k4,4n -k5,5n | tail -n1 | cut -d' ' -f6)
DIRTY     := $(shell git status --porcelain --untracked-files=normal 2>/dev/null)
VERSION   := $(if $(DIRTY),dev,$(or $(TAG),dev))
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64
GOBUILD   := go build -tags release -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"

.PHONY: build release lint fmt fakegame

build:
	$(GOBUILD) ./cmd/stonemcp

release:
	rm -rf dist
	for p in $(PLATFORMS); do \
		os=$${p%%/*}; arch=$${p#*/}; ext=; \
		if [ $$os = windows ]; then ext=.exe; fi; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 $(GOBUILD) \
			-o dist/stonemcp-$(VERSION)-$$os-$$arch$$ext ./cmd/stonemcp || exit 1; \
	done

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
