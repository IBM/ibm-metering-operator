# Copyright 2020 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This repo is build locally for dev/test by default;
# Override this variable in CI env.
BUILD_LOCALLY ?= 1

# set these variables to the tag or SHA for the ubi image used in the Dockerfile
UBI_IMAGE_TAG_AMD=8.2-345
UBI_IMAGE_SHA_PPC=d397c81371b235816967b1140ace8cb0b621f1872b8a1c1131658e7aac5fb95c
UBI_IMAGE_SHA_390=cc1a7031050564fe76969d249a6d89aadbd5ea7a8969b21718c09ceb77d57715

# Image URL to use all building/pushing image targets;
# Use your own docker registry and image name for dev/test by overriding the IMG and REGISTRY environment variable.
IMG ?= ibm-metering-operator
ifeq ($(BUILD_LOCALLY),0)
REGISTRY ?= "hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com/ibmcom"
endif
CSV_VERSION ?= $(VERSION)

# Set the registry and tags for the operand images
OPERAND_REGISTRY ?= $(REGISTRY)
OPERAND_TAG_DM ?= 3.6.0
OPERAND_TAG_UI ?= 3.6.0
OPERAND_TAG_MCMUI ?= 3.6.0
OPERAND_TAG_REPORT ?= 3.6.0

# Github host to use for checking the source tree;
# Override this variable ue with your own value if you're working on forked repo.
GIT_HOST ?= github.com/IBM

PWD := $(shell pwd)
BASE_DIR := $(shell basename $(PWD))

# Keep an existing GOPATH, make a private one if it is undefined
GOPATH_DEFAULT := $(PWD)/.go
export GOPATH ?= $(GOPATH_DEFAULT)
GOBIN_DEFAULT := $(GOPATH)/bin
export GOBIN ?= $(GOBIN_DEFAULT)
TESTARGS_DEFAULT := "-v"
export TESTARGS ?= $(TESTARGS_DEFAULT)
DEST := $(GOPATH)/src/$(GIT_HOST)/$(BASE_DIR)
#Pushing with release tag after moving to artifactory
VERSION ?= $(shell cat version/version.go | grep "Version =" | awk '{ print $3}' | tr -d '"')
#VERSION ?= $(shell git describe --exact-match 2> /dev/null || \
#                 git describe --match=$(git rev-parse --short=8 HEAD) --always --dirty --abbrev=8)

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
    TARGET_OS ?= linux
    XARGS_FLAGS="-r"
else ifeq ($(LOCAL_OS),Darwin)
    TARGET_OS ?= darwin
    XARGS_FLAGS=
else
    $(error "This system's OS $(LOCAL_OS) isn't recognized/supported")
endif

ARCH := $(shell uname -m)
LOCAL_ARCH := "amd64"
ifeq ($(ARCH),x86_64)
    LOCAL_ARCH="amd64"
else ifeq ($(ARCH),ppc64le)
    LOCAL_ARCH="ppc64le"
else ifeq ($(ARCH),s390x)
    LOCAL_ARCH="s390x"
else
    $(error "This system's ARCH $(ARCH) isn't recognized/supported")
endif

# Setup DOCKER_BUILD_OPTS after all includes complete
#Variables for redhat ubi certification required labels
IMAGE_NAME=$(IMG)
IMAGE_DISPLAY_NAME=IBM Metering Operator
IMAGE_MAINTAINER=sgrube@us.ibm.com
IMAGE_VENDOR=IBM
IMAGE_VERSION=$(VERSION)
IMAGE_DESCRIPTION=Operator used to install a service to meter workloads in a kubernetes cluster
IMAGE_SUMMARY=$(IMAGE_DESCRIPTION)
IMAGE_OPENSHIFT_TAGS=metering
$(eval WORKING_CHANGES := $(shell git status --porcelain))
$(eval BUILD_DATE := $(shell date +%m/%d@%H:%M:%S))
$(eval GIT_COMMIT := $(shell git rev-parse --short HEAD))
$(eval VCS_REF := $(GIT_COMMIT))
IMAGE_RELEASE=$(VCS_REF)
GIT_REMOTE_URL = $(shell git config --get remote.origin.url)
DOCKER_BUILD_OPTS := --build-arg "IMAGE_NAME=$(IMAGE_NAME)" --build-arg "IMAGE_DISPLAY_NAME=$(IMAGE_DISPLAY_NAME)" --build-arg "IMAGE_MAINTAINER=$(IMAGE_MAINTAINER)" --build-arg "IMAGE_VENDOR=$(IMAGE_VENDOR)" --build-arg "IMAGE_VERSION=$(IMAGE_VERSION)" --build-arg "IMAGE_RELEASE=$(IMAGE_RELEASE)" --build-arg "IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION)" --build-arg "IMAGE_SUMMARY=$(IMAGE_SUMMARY)" --build-arg "IMAGE_OPENSHIFT_TAGS=$(IMAGE_OPENSHIFT_TAGS)" --build-arg "VCS_REF=$(VCS_REF)" --build-arg "VCS_URL=$(GIT_REMOTE_URL)" --build-arg "SELF_METER_IMAGE_TAG=$(SELF_METER_IMAGE_TAG)"


all: fmt check test coverage build images

ifneq ("$(realpath $(DEST))", "$(realpath $(PWD))")
    $(error Please run 'make' from $(DEST). Current directory is $(PWD))
endif

include common/Makefile.common.mk

############################################################
# work section
############################################################
$(GOBIN):
	@echo "create gobin"
	@mkdir -p $(GOBIN)

work: $(GOBIN)

############################################################
# format section
############################################################

# All available format: format-go format-protos format-python
# Default value will run all formats, override these make target with your requirements:
#    eg: fmt: format-go format-protos
fmt: format-go format-protos format-python

############################################################
# check section
############################################################

check: lint

# All available linters: lint-dockerfiles lint-scripts lint-yaml lint-copyright-banner lint-go lint-python lint-helm lint-markdown lint-sass lint-typescript lint-protos
# Default value will run all linters, override these make target with your requirements:
#    eg: lint: lint-go lint-yaml
lint: lint-all

############################################################
# test section
############################################################

test:
	@go test ${TESTARGS} ./...

############################################################
# coverage section
############################################################

coverage:
	@common/scripts/codecov.sh ${BUILD_LOCALLY}

############################################################
# install operator sdk section
############################################################

install-operator-sdk:
	@operator-sdk version 2> /dev/null ; if [ $$? -ne 0 ]; then ./common/scripts/install-operator-sdk.sh; fi

############################################################
# build section
############################################################

build: build-amd64 build-ppc64le build-s390x

build-amd64:
	@echo "Building the ${IMG} amd64 binary..."
	@GOARCH=amd64 common/scripts/gobuild.sh build/_output/bin/$(IMG) ./cmd/manager

build-ppc64le:
	@echo "Building the ${IMG} ppc64le binary..."
	@GOARCH=ppc64le common/scripts/gobuild.sh build/_output/bin/$(IMG)-ppc64le ./cmd/manager

build-s390x:
	@echo "Building the ${IMG} s390x binary..."
	@GOARCH=s390x common/scripts/gobuild.sh build/_output/bin/$(IMG)-s390x ./cmd/manager

local:
	@GOOS=darwin common/scripts/gobuild.sh build/_output/bin/$(IMG) ./cmd/manager

############################################################
# images section
############################################################

ifeq ($(BUILD_LOCALLY),0)
    export CONFIG_DOCKER_TARGET = config-docker
config-docker:
endif


build-image-amd64: build-amd64
	@docker build -t $(REGISTRY)/$(IMG)-amd64:$(VERSION) $(DOCKER_BUILD_OPTS) --build-arg "IMAGE_NAME_ARCH=$(IMAGE_NAME)-amd64" --build-arg "UBI_IMAGE_TAG_AMD=$(UBI_IMAGE_TAG_AMD)" -f build/Dockerfile .

build-image-ppc64le: build-ppc64le
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(REGISTRY)/$(IMG)-ppc64le:$(VERSION) $(DOCKER_BUILD_OPTS) --build-arg "IMAGE_NAME_ARCH=$(IMAGE_NAME)-ppc64le" --build-arg "UBI_IMAGE_SHA_PPC=$(UBI_IMAGE_SHA_PPC)" -f build/Dockerfile.ppc64le .

build-image-s390x: build-s390x
	@docker run --rm --privileged multiarch/qemu-user-static:register --reset
	@docker build -t $(REGISTRY)/$(IMG)-s390x:$(VERSION) $(DOCKER_BUILD_OPTS) --build-arg "IMAGE_NAME_ARCH=$(IMAGE_NAME)-s390x" --build-arg "UBI_IMAGE_SHA_390=$(UBI_IMAGE_SHA_390)" -f build/Dockerfile.s390x .

push-image-amd64: $(CONFIG_DOCKER_TARGET) build-image-amd64
	@docker push $(REGISTRY)/$(IMG)-amd64:$(VERSION)

push-image-ppc64le: $(CONFIG_DOCKER_TARGET) build-image-ppc64le
	@docker push $(REGISTRY)/$(IMG)-ppc64le:$(VERSION)

push-image-s390x: $(CONFIG_DOCKER_TARGET) build-image-s390x
	@docker push $(REGISTRY)/$(IMG)-s390x:$(VERSION)

############################################################
# multiarch-image section
############################################################

images: push-image-amd64 push-image-ppc64le push-image-s390x multiarch-image

multiarch-image:
	@curl -L -o /tmp/manifest-tool https://github.com/estesp/manifest-tool/releases/download/v1.0.0/manifest-tool-linux-amd64
	@chmod +x /tmp/manifest-tool
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(REGISTRY)/$(IMG)-ARCH:$(VERSION) --target $(REGISTRY)/$(IMG) --ignore-missing
	/tmp/manifest-tool push from-args --platforms linux/amd64,linux/ppc64le,linux/s390x --template $(REGISTRY)/$(IMG)-ARCH:$(VERSION) --target $(REGISTRY)/$(IMG):$(VERSION) --ignore-missing

############################################################
# CSV section
############################################################
csv: ## Push CSV package to the catalog
	@RELEASE=${CSV_VERSION} common/scripts/push-csv.sh

############################################################
# SHA section
############################################################
.PHONY: get-all-image-sha
get-all-image-sha: get-report-image-sha get-mcmui-image-sha get-ui-image-sha get-dm-image-sha
	@echo Got SHAs for all operand images

.PHONY: get-dm-image-sha
get-dm-image-sha:
	@echo Get SHA for metering-data-manager:$(OPERAND_TAG_DM)
	@scripts/get-image-sha.sh DM $(OPERAND_REGISTRY)/metering-data-manager $(OPERAND_TAG_DM)

.PHONY: get-ui-image-sha
get-ui-image-sha:
	@echo Get SHA for metering-ui:$(OPERAND_TAG_UI)
	@scripts/get-image-sha.sh UI $(OPERAND_REGISTRY)/metering-ui $(OPERAND_TAG_UI)

.PHONY: get-mcmui-image-sha
get-mcmui-image-sha:
	@echo Get SHA for metering-mcmui:$(OPERAND_TAG_MCMUI)
	@scripts/get-image-sha.sh MCMUI $(OPERAND_REGISTRY)/metering-mcmui $(OPERAND_TAG_MCMUI)

.PHONY: get-report-image-sha
get-report-image-sha:
	@echo Get SHA for metering-report:$(OPERAND_TAG_REPORT)
	@scripts/get-image-sha.sh REPORT $(OPERAND_REGISTRY)/metering-report $(OPERAND_TAG_REPORT)

############################################################
# Red Hat certification section
############################################################
.PHONY: bundle
bundle:
	@echo --- Updating the bundle directory with latest yamls from olm-catalog ---
	-mkdir bundle
	rm -rf bundle/*
	cp -r -p deploy/olm-catalog/ibm-metering-operator/$(CSV_VERSION)/* bundle/
	cp -p deploy/olm-catalog/ibm-metering-operator/ibm-metering-operator.package.yaml bundle/
	# if old CSVs are listed in other channels in ibm-metering-operator.package.yaml, need to copy them
	#cp -p deploy/olm-catalog/ibm-metering-operator/3.6.2/ibm-metering-operator.v3.6.2.clusterserviceversion.yaml bundle/
	# need certificate-crd.yaml in the bundle so that the operator can be started during the RH scan
	cp scripts/rh-scan/certificate-crd.yaml bundle/
	@echo Add certmanager info to CSV for scan
	@scripts/rh-scan/add-certman-info.sh $(CSV_VERSION)
	cd bundle && zip ibm-metering-metadata ./*.yaml && cd ..

.PHONY: install-operator-courier
install-operator-courier:
	@echo --- Installing Operator Courier ---
	pip3 install operator-courier

.PHONY: verify-bundle
verify-bundle:
	@echo --- Verify bundle is ready for Red Hat certification ---
	operator-courier --verbose verify --ui_validate_io bundle/

#redhat-certify-ready: bundle install-operator-courier verify-bundle
.PHONY: redhat-certify-ready
redhat-certify-ready: bundle verify-bundle

############################################################
# clean section
############################################################
clean:
	rm -rf build/_output

############################################################
# application section
############################################################

install: ## Install all resources (CR/CRD's, RBCA and Operator)
	@echo ....... Set environment variables ......
	- export DEPLOY_DIR=deploy/crds
	- export WATCH_NAMESPACE=${NAMESPACE}
	# @echo ....... Creating namespace .......
	# - kubectl create namespace ${NAMESPACE}
	@echo ....... Applying CRDS and Operator .......
	- for crd in $(shell ls deploy/crds/*_crd.yaml); do kubectl apply -f $${crd}; done
	@echo ....... Applying RBAC .......
	- kubectl apply -f deploy/service_account.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/role_binding.yaml -n ${NAMESPACE}
	@echo ....... Applying Operator .......
	- kubectl apply -f deploy/olm-catalog/${BASE_DIR}/${CSV_VERSION}/${BASE_DIR}.v${CSV_VERSION}.clusterserviceversion.yaml -n ${NAMESPACE} --validate=false
	@echo ....... Creating the Instances .......
	- kubectl apply -f deploy/crds/operator.ibm.com_v1alpha1_metering_cr.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/crds/operator.ibm.com_v1alpha1_meteringui_cr.yaml -n ${NAMESPACE}
	- kubectl apply -f deploy/crds/operator.ibm.com_v1alpha1_meteringreportserver_cr.yaml -n ${NAMESPACE}

uninstall: ## Uninstall all resources created in the $ make install
	@echo ....... Uninstalling .......
	@echo ....... Deleting CRs .......
	- kubectl delete -f deploy/crds/operator.ibm.com_v1alpha1_metering_cr.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/crds/operator.ibm.com_v1alpha1_meteringui_cr.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/crds/operator.ibm.com_v1alpha1_meteringreportserver_cr.yaml -n ${NAMESPACE}
	@echo ....... Deleting Operator .......
	- kubectl delete -f deploy/olm-catalog/${BASE_DIR}/${CSV_VERSION}/${BASE_DIR}.v${CSV_VERSION}.clusterserviceversion.yaml -n ${NAMESPACE}
	@echo ....... Deleting CRDs.......
	- for crd in $(shell ls deploy/crds/*_crd.yaml); do kubectl delete -f $${crd}; done
	@echo ....... Deleting Roles and Service Account .......
	- kubectl delete -f deploy/role_binding.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/service_account.yaml -n ${NAMESPACE}
	- kubectl delete -f deploy/role.yaml -n ${NAMESPACE}
	# @echo ....... Deleting namespace .......
	# - kubectl delete namespace ${NAMESPACE}

.PHONY: all work build check lint test coverage images multiarch-image
