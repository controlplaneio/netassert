NAME := talk-net-assert
PKG := github.com/controlplane/$(NAME)
REGISTRY := docker.io

SHELL := /bin/bash
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

GIT_MESSAGE := $(shell git -c log.showSignature=false log --max-count=1 --pretty=format:"%H")
GIT_SHA := $(shell git log -1 --format=%h)
GIT_TAG ?= $(shell bash -c 'TAG=$$(git tag | tail -n1); echo "$${TAG:-none}"')
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif

# golang buildtime, more at https://github.com/jessfraz/pepper/blob/master/Makefile
CTIMEVAR=-X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

CONTAINER_TAG ?= $(GIT_TAG)
CONTAINER_NAME := $(REGISTRY)/$(NAME):$(CONTAINER_TAG)

TEST_FILE := "test/test-localhost-remote.yaml"

export NAME REGISTRY BUILD_DATE GIT_MESSAGE GIT_SHA GIT_TAG CONTAINER_TAG CONTAINER_NAME

.PHONY: cluster
cluster: ## builds a test cluster image
	@echo "+ $@"
	gcloud container clusters create \
	--zone europe-west2-a \
	--machine-type n1-highcpu-16 \
	--enable-autorepair \
	--no-enable-legacy-authorization \
	--num-nodes 1 \
	--preemptible \
	--enable-network-policy \
	netpol3

.PHONY: cluster-def
cluster-def: ## builds a test cluster
	@echo "+ $@"
	gcloud beta container clusters create np-2 \
    --enable-network-policy \
		--preemptible

.PHONY: build
build: ## builds a docker image
	@echo "+ $@"
	docker build --tag "${CONTAINER_NAME}" .

.PHONY: run
run: ## runs the last build docker image
	@echo "+ $@"
	docker run -i "${CONTAINER_NAME}" ${ARGS}

.PHONY: run-in-docker
run-in-docker: ## runs the last build docker image inside docker
	@echo "+ $@"
	docker run -i \
		--net=host \
		${DOCKER_ARGS} \
		-v /var/run/docker.sock:/var/run/docker.sock:ro \
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

.PHONY: test
test: ## build, test, and push container, then run local tests
	@echo "+ $@"
	make rollcage && ./netassert test/test-all.yaml

.PHONY: rollcage
rollcage: ## build, test, and push the container
	@echo "+ $@"
	rollcage build run push \
		--interactive false \
	  --tag sublimino/scratch:dev --pull=false "npm test"  \
		-- \
	  --net=host \
		--env DEBUG="" \
		--env "TEST_YAML=$$(cat test/test.yaml | base64 -w0)"

.PHONY: rollcage-docker
rollcage-docker: ## experimental, does not currently work with gcloud
	@echo "+ $@"
	rollcage build run push \
		--interactive false \
		--tag sublimino/scratch:dev --pull=false "npm test" \
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
    | grep -Ev '^help\b[[:space:]]*:' \
    | sort \
    | awk 'BEGIN {FS = ":.*?##"}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

