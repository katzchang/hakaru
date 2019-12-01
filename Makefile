export AWS_PROIFLE        ?= sunrise201911
export AWS_DEFAULT_REGION := ap-northeast-1

.PHONY: all install imports fmt test run build clean upload

GOOS   ?=
GOARCH ?=
GOSRC  := $(GOPATH)/src

all: install run

install:

imports:
	goimports -w .

fmt:
	gofmt -w .

test:
	go test -v -tags=unit $$(go list ./... | grep -v '/vendor/')

run: main.go
	go run main.go

build: test
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o hakaru

clean:
	rm -rf hakaru *.tgz

# deployment
artifacts.tgz: provisioning/instance
	$(MAKE) build GOOS=linux GOARCH=amd64
	tar czf artifacts.tgz hakaru db provisioning/instance

ARTIFACTS_BUCKET := $(AWS_PROFILE)-hakaru-artifacts

# ci からアップロードできなくなった場合のターゲット
upload: clean artifacts.tgz
	aws s3 cp artifacts.tgz s3://$(ARTIFACTS_BUCKET)/latest/artifacts.tgz
	aws s3 cp artifacts.tgz s3://$(ARTIFACTS_BUCKET)/$$(git rev-parse HEAD)/artifacts.tgz
