GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
GOBINDATA=go-bindata
BINARY_NAME=naruko
VERSION := 0.0.1
.DEFAULT_GOAL := help
.PHONY: setup resource resource/local clean help bin bin/arm

deps:
	$(GOGET) -u github.com/jteeuwen/go-bindata/...
	$(GOGET) -u github.com/golang/dep/cmd/dep
	$(GOGET) github.com/Songmu/make2help/cmd/make2help

## Create resource for Production
resource:
	$(GOBINDATA) -pkg resource -o resource/asset.go resource

## Create resource for local
resource/local:
	$(GOBINDATA) -debug -pkg resource -o resource/asset.go resource

bin: resource
	$(GOBUILD) -o bin/$(BINARY_NAME) -v

bin/arm7: resource
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -o bin/$(BINARY_NAME)_arm7

run: resource/local
	$(GORUN) main.go

clean:
	$(GOCLEAN)
	rm -rf ./bin/*

## Show help
help:
	@make2help $(MAKEFILE_LIST)

