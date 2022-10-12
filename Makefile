DIR:=$(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
LOCAL_BIN:=$(DIR)/bin

describe_job = @echo "=====================\n$1...\n====================="

GOLANG_CI_LINT_VERSION ?= v1.46.2
lint-deps:
ifeq ("$(wildcard $(LOCAL_BIN)/golangci-lint)","")
	GOBIN=$(LOCAL_BIN) go install -mod=mod github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANG_CI_LINT_VERSION)
endif

IMPORTS_REVISER_VERSION ?= v2.5.1
imports-deps:
ifeq ("$(wildcard $(LOCAL_BIN)/goimports-reviser)","")
	GOBIN=$(LOCAL_BIN) go install -mod=mod github.com/incu6us/goimports-reviser/v2@$(IMPORTS_REVISER_VERSION)
endif

deps:
	$(call describe_job,"Installing dependencies")
	$(MAKE) lint-deps
	$(MAKE) imports-deps
	go mod tidy

git-hooks:
	$(call describe_job,"Setting up git hooks")
	/bin/sh ./scripts/hooks.sh

environment:
	$(call describe_job,"Local development setup")
	$(MAKE) git-hooks
	$(MAKE) deps

imports:
	$(call describe_job,"Running imports")
	$(MAKE) imports-deps
	find . -name \*.go \
	    -not -path "./vendor/*" \
	    -exec $(LOCAL_BIN)/goimports-reviser -file-path {} -rm-unused -set-alias \
	    -format -local github.com/Propertyfinder/ \;

lint:
	$(call describe_job,"Running linter")
	$(MAKE) lint-deps
	$(LOCAL_BIN)/golangci-lint run
