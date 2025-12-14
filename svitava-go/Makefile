SHELL := /bin/bash

.PHONY: default clean build

SOURCES:=$(shell find . -name '*.go')
BINARY:=svitava-go
#OUTDIR:=.
OUTDIR:=/tmp/ramdisk

default: build

clean: ## Run go clean
	@go clean
	# cleanup raster images
	rm *.bmp

build:	${OUTDIR}/${BINARY} ## Build binary containing service executable

run:	${OUTDIR}/${BINARY}
	${OUTDIR}/${BINARY} -s

${OUTDIR}/${BINARY}:	${SOURCES}
	@go build -o ${OUTDIR}

benchmark: ## Run benchmarks
	@go test -bench . ./...

uml.png:	uml.puml  ## Generate image with UML diagram
	java -jar plantuml.jar uml.puml

uml.puml:	## Generate UML diagram
	goplantuml --recursive --show-aggregations --show-connection-labels . > uml.puml

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z0-9_.-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-33s\033[0m %s\n", $$1, $$2}'
	@echo ''
