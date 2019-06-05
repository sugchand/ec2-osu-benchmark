GO ?= go
GOPATH := $(CURDIR)
GOSRCPATH := ./src/ec2-osu-benchmark
GOBINPATH := ./bin
GOOUTPUTBIN := $(GOBINPATH)/ec2-osu-benchmark
SHELL := /bin/bash
DEP := $(shell command -v dep  2> /dev/null)

export GOPATH
export GOSRCPATH
export
.PHONY : clean build debug

all: build

clean:
	rm -rf $(GOBINPATH)/*

debug: GCFLAGS = -gcflags=all="-N -l"

debug:
	@echo -e "\nCompiling OSU benchmarking on EC2 with gdb symbols..."
	$(MAKE) build

build:
ifndef DEP
$(error "dep is not available please install go dep package manager")
endif
	#Make sure the 'dep status' shows right data
	#-@(cd $(GOSRCPATH);$(DEP) status 2> /dev/null)
	@echo -e "\n\tSet 'GOPATH' to '$(GOPATH)'"
	@echo -e "\tRun 'dep ensure' in $(GOSRCPATH) to install missing third party packages\n"
	@echo -e "\tRun 'env DEPNOLOCK=1 dep ensure' in $(GOSRCPATH) in case running the application in vbox shared directory"
	$(GO) build $(GCFLAGS) -o $(GOOUTPUTBIN) $(GOSRCPATH)
	@echo -e "\n\t**** RESULT : $$? : Build completed!!! ****\n\t**** Binary is at $$PWD/bin ****"

tests:
	@echo -e "\n\tSet 'GOPATH' to '$(GOPATH)'"
	@echo -e "\tRun 'dep ensure' in $(GOSRCPATH) to install missing third party packages\n"
	$(GO) test -v $(GOSRCPATH)/...
