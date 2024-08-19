PKG := github.com/tapmon/tapmond


include make/release_flags.mk
include make/testing_flags.mk


GOBUILD := go build -v
GOINSTALL := go install -v
GOTEST := go test -v
GOMOD := go mod

# =========
# UTILITIES
# =========

gen: rpc sqlc

sqlc:
	@$(call print, "Generating sql models and queries in Go")
	./scripts/gen_sqlc_docker.sh

sqlc-check: sqlc
	@$(call print, "Verifying sql code generation.")
	if test -n "$$(git status --porcelain '*.go')"; then echo "SQL models not properly generated!"; git status --porcelain '*.go'; exit 1; fi

rpc:
	@$(call print, "Compiling protos.")
	cd ./tapmonrpc; ./gen_protos_docker.sh

build:
	@$(call print, "Building lightning-terminal.")
	$(GOBUILD) -tags="$(LND_RELEASE_TAGS)" -o tapmond $(PKG)/cmd/tapmond
	$(GOBUILD) -tags="$(LND_RELEASE_TAGS)" -o tapmoncli $(PKG)/cmd/tapmoncli
