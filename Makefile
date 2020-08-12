NAME := netassert
GITHUB_ORG := controlplaneio
DOCKER_HUB_ORG := controlplane

### github.com/controlplaneio/ensure-content.git makefile-header START ###
ifeq ($(NAME),)
  $(error NAME required, please add "NAME := project-name" to top of Makefile)
else ifeq ($(GITHUB_ORG),)
    $(error GITHUB_ORG required, please add "GITHUB_ORG := controlplaneio" to top of Makefile)
else ifeq ($(DOCKER_HUB_ORG),)
    $(error DOCKER_HUB_ORG required, please add "DOCKER_HUB_ORG := controlplane" to top of Makefile)
endif

PKG := github.com/$(GITHUB_ORG)/$(NAME)
DOCKER_REGISTRY_FQDN ?= docker.io
DOCKER_HUB_URL := $(DOCKER_REGISTRY_FQDN)/$(DOCKER_HUB_ORG)/$(NAME)

SHELL := /bin/bash
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

GIT_MESSAGE := $(shell git -c log.showSignature=false \
	log --max-count=1 --pretty=format:"%H")
GIT_SHA := $(shell git -c log.showSignature=false \
	log -1 --format=%h)
GIT_TAG := $(shell bash -c 'TAG=$$(git -c log.showSignature=false \
	describe --tags --exact-match --match "v*" 2>/dev/null | sort | tail -n1); \
	echo "$${TAG:-none}"')
GIT_UNTRACKED_CHANGES := $(shell git -c log.showSignature=false \
	status --porcelain --untracked-files=no)

ifndef CONTAINER_TAG
	ifeq ($(GIT_TAG),none)
		CONTAINER_TAG := $(GIT_SHA)
	else
		CONTAINER_TAG := $(GIT_TAG)
	endif
	ifneq ($(GIT_UNTRACKED_CHANGES),)
		CONTAINER_TAG := $(CONTAINER_TAG)-dirty
	endif
endif

CONTAINER_NAME ?= $(DOCKER_HUB_URL):$(CONTAINER_TAG)
CONTAINER_NAME_LATEST ?= $(DOCKER_HUB_URL):latest
CONTAINER_NAME_TESTING ?= $(DOCKER_HUB_URL):testing

# golang buildtime, more at https://github.com/jessfraz/pepper/blob/master/Makefile
CTIMEVAR=-X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

export NAME DOCKER_HUB_URL BUILD_DATE GIT_MESSAGE GIT_SHA GIT_TAG \
  CONTAINER_TAG CONTAINER_NAME CONTAINER_NAME_LATEST CONTAINER_NAME_TESTING
### github.com/controlplaneio/ensure-content.git makefile-header END ###

TEST_FILE := "test/test-localhost-remote.yaml"

.PHONY: all test
.SILENT:

all: help

.PHONY: cluster
cluster: ## creates a test GKE cluster
	@echo "+ $@"
	gcloud container clusters create \
		--zone europe-west2-a \
		--machine-type n1-highcpu-16 \
		--enable-autorepair \
		--no-enable-legacy-authorization \
		--num-nodes 1 \
		--preemptible \
		--enable-network-policy \
		netassert-test

.PHONY: cluster-kill
cluster-kill: ## deletes a test GKE cluster
	@echo "+ $@"
	yes | gcloud container clusters delete \
		--zone europe-west2-a \
		netassert-test

.PHONY: build
build: ## builds a docker image
	@echo "+ $@"
	docker build --tag "${CONTAINER_NAME}" .

.PHONY: run
run: ## runs the last build docker image
	@echo "+ $@"
	docker run -i "${CONTAINER_NAME}" ${ARGS}

.PHONY: push
push: ## pushes a docker image
	@echo "+ $@"
	docker push "${CONTAINER_NAME}"

.PHONY: get-container-tag
get-container-tag: ## get the container's tag
	echo "${CONTAINER_NAME}"

.PHONY: run-in-docker
run-in-docker: ## runs the last build docker image inside docker
	@echo "+ $@"; \
		set -x ; \
		LINK=$(shell readlink -f ~/.ssh/config); \
		docker run -i \
			--net=host \
			--cap-add NET_ADMIN \
			--cap-add NET_RAW \
			${DOCKER_ARGS} \
			-v ~/.config/gcloud/:/root/.config/gcloud/ \
			-v ~/.ssh/:/tmp/.ssh/:ro \
			-v ~/.kube/:/root/.kube:ro \
			-v $${LINK:-/dev/null}:/tmp/ssh-config:ro \
			-v /var/run/docker.sock:/var/run/docker.sock:ro \
			\
			-v "$${KUBECONFIG:-/dev/null}:$${KUBECONFIG:-/dev/null}" \
			-e KUBECONFIG="${KUBECONFIG}" \
			\
			-e DEBUG=0 \
			\
			"${CONTAINER_NAME}" ${ARGS}

.PHONY: jenkins
jenkins: ## run acceptance tests
	@echo "+ $@"
	make build
	make run-in-docker \
		ARGS='netassert --offline --image ${CONTAINER_NAME} ${TEST_FILE}'

.PHONY: rollcage-test
rollcage-test: ## build, test, and push container, then run local tests
	@echo "+ $@"
	make rollcage && ./netassert test/test-all.yaml

.PHONY: test-local-docker
test-local-docker: ## test against local container
	@echo "+ $@"

	# test against local container
	make build CONTAINER_NAME="$(CONTAINER_NAME_TESTING)"

	docker rm --force netassert-http-echo 2>/dev/null || true;

	# start echo server
	docker run -d -p 59942:59942 \
		--name netassert-http-echo \
		hashicorp/http-echo \
		-listen=:59942 -text 'netassert test endpoint'

	COUNT=0; \
	until [[ "$$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' netassert-http-echo)" != "" ]]; do \
			sleep 0.5; \
			if [[ $$((COUNT++)) -gt 10 ]]; then echo 'Container did not start'; exit 1; fi; \
	done;

	# get IP to template into test file
	bash -euo pipefail -c " \
		TMP_TEST_FILE=$$(mktemp); \
		\
		cat test/test-localhost-docker-TEMPLATE.yaml \
			| DOCKER_HOST_IP=$$(docker inspect \
        	--format='{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' \
        	netassert-http-echo) \
        	\
        	envsubst \
			| tee -a \$${TMP_TEST_FILE}; \
		\
		./netassert \
			--verbose \
			--image $(CONTAINER_NAME_TESTING) \
			--no-pull \
			--ssh-user $${SSH_USER:-root} \
			--ssh-options \"-o StrictHostKeyChecking=no\" \
			\$${TMP_TEST_FILE}; \
	"

	docker rm --force netassert-http-echo 2>/dev/null || true;

.PHONY: test
test: ## build, test, and push container, then run local tests
	# Run unit tests
	npm run test:unit

	# test against local container
	make test-local-docker

	# test against remote hosts
	make test-deploy
	make push CONTAINER_NAME="$(CONTAINER_NAME_TESTING)"
	./netassert \
		--image $(CONTAINER_NAME_TESTING) \
		--ssh-user $${SSH_USER:-root} \
		--ssh-options "-o StrictHostKeyChecking=no" \
		test/test-all.yaml

	# test in docker against remote hosts
	bash -c "make run-in-docker \
		CONTAINER_NAME=$(CONTAINER_NAME_TESTING) \
		ARGS='netassert \
			--image $(CONTAINER_NAME_TESTING) \
			--ssh-user $${SSH_USER:-root} \
			--ssh-options \"-o StrictHostKeyChecking=no\" \
			test/test-all.yaml'"

.PHONY: test-local
test-local: ## test from the local machine
	@echo "+ $@"
	./netassert \
		--image ${CONTAINER_NAME_TESTING} \
		test/test-all.yaml

.PHONY: test-deploy
test-deploy: ## deploy test services
	@echo "+ $@"
	set -x;
	kubectl apply \
						-f resource/deployment/demo.yaml \
						-f resource/net-pol/web-deny-all.yaml \
						-f resource/net-pol/test-services-allow.yaml;


.PHONY: rollcage
rollcage: ## build, test, and push the container
	@echo "+ $@"
	rollcage build run push \
		--interactive false \
	  --tag controlplane/netassert:none \
	  --pull=false "npm test"  \
		-- \
	  --net=host \
		--env DEBUG="" \
		--env "TEST_YAML=$$(cat test/test.yaml | base64 -w0)"

.PHONY: rollcage-docker
rollcage-docker: ## experimental, does not currently work with gcloud
	@echo "+ $@"
	rollcage build run push \
		--interactive false \
		--tag controlplane/netassert:none \
		--pull=false "npm test" \
		-- \
	  --net=host \
		--env DEBUG=1 \
		--env "TEST_YAML=$$(base64 -w0 test/test.yaml)" \
		--volume /var/run/docker.sock:/var/run/docker.sock:ro  \
		--volume $${HOME}/.ssh:/root/.ssh \
		--volume $${HOME}/.andy_sync/conf/.ssh/config:/opt/ssh_config \
		--volume $${HOME}/.kube:/root/.kube:ro

# ---

.PHONY: add-make-rule
add-make-rule: ## add a new rule to this Makefile
	@echo "+ $@"
	set -x ; \
	if [[ "$${ACTION:-}" != "" ]]; then \
		sed -E "/^.PHONY: new-make-rule$$/i \
.PHONY: $${ACTION}\n\
$${ACTION}: \#\# help\n\
\t\@echo \"+ \$$\@\"\n\
\techo \"Not implemented\"\n\
" \
		-i Makefile; \
		LINE=$$(grep -E '^.PHONY: rollcage' Makefile  --line-number | head -n 1 | cut  -d: -f1); \
		vim Makefile \
			"+call cursor($${LINE}, 14)"; \
	else \
		echo "ACTION required"; \
	fi

.PHONY: help
help: ## parse jobs and descriptions from this Makefile
	@grep -E '^[ a-zA-Z0-9_-]+:([^=]|$$)' $(MAKEFILE_LIST) \
    | grep -Ev '^(all|help)\b[[:space:]]*:' \
    | sort \
    | awk 'BEGIN {FS = ":.*?##"}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
