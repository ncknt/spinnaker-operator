# Copyright 2019, Armory
#
# Licensed under the Apache License, Version 2.0 (the "License")
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Inspired by MySQL Operator Makefile: https://github.com/oracle/mysql-operator/blob/master/Makefile

VERSION ?= $(shell git describe --always --dirty)
REGISTRY_ORG ?= "armory"
ARCH    ?= amd64
OS      ?= linux
UNAME_S := $(shell uname -s)
NAMESPACE ?= "default"
PWD = $(shell pwd)

PKG             := github.com/armory/spinnaker-operator
REGISTRY        := docker.io
SRC_DIRS        := cmd pkg
COMMAND         := cmd/manager/main
BUILD_DIR       := ${PWD}/bin/$(OS)_$(ARCH)
BINARY 			:= ${BUILD_DIR}/spinnaker-operator


.PHONY: all
all: build

.PHONY: test
test: build-dirs Makefile
	@go test -cover ./...

.PHONY: test-docker
test-docker: build-dirs Makefile
	@docker build \
	-t $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator-test:$(VERSION) \
	-f build/Dockerfile-test .
	@echo "Running tests..."
	@docker run \
	-v $(PWD):/opt/spinnaker-operator-test \
	$(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator-test:$(VERSION) \
	go test -cover ./...

.PHONY: build-dirs
build-dirs:
	@echo "Creating build directories ${BUILD_DIR}"
	@mkdir -p $(BUILD_DIR)

# Regenerates CRD yamls out of any changes in spinnakerservice_types.go
.PHONY: generate
generate: build-dirs
	operator-sdk generate k8s
	operator-sdk generate openapi

.PHONY: build
build: build-dirs Makefile
	@echo "Building: $(BINARIES)"
	@go build -i ${LDFLAGS} -o ${BINARY} cmd/manager/main.go

.PHONY: build-docker
build-docker:
	@docker build \
	--build-arg OPERATOR_VERSION=${VERSION} \
	--build-arg OPERATOR_PATH=bin/$(OS)_$(ARCH)/spinnaker-operator \
	-t $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:$(VERSION) \
	-f build/Dockerfile .

# Note: Only used for development, i.e. in CI the images are pushed using Wercker.
.PHONY: push
push:
	@docker push $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:$(VERSION)

.PHONY: publish
publish:
	@docker tag $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:$(VERSION) $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:latest
	@docker push $(REGISTRY)/$(REGISTRY_ORG)/spinnaker-operator:latest

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: lint
lint:
	@find pkg cmd -name '*.go' | grep -v 'generated' | xargs -L 1 golint

.PHONY: clean
clean:
	rm -rf bin

.PHONY: run-dev
run-dev:
	@go run \
	    cmd/manager/main.go \
	    --kubeconfig=${KUBECONFIG} \
	    --namespace=${NAMESPACE}

# Depends on operator-sdk for now
.PHONE: debug
debug:
	OPERATOR_NAME=local-operator \
    WATCH_NAMESPACE=operator \
	dlv debug --headless  --listen=:2345 --headless --log --api-version=2 cmd/manager/main.go -- \
	--kubeconfig ~/.kube/config --disable-admission-controller

