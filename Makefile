NAME := netassert
PKG := github.com/controlplane/$(NAME)
REGISTRY := docker.io/controlplane

SHELL := /bin/bash
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

GIT_MESSAGE := $(shell git -c log.showSignature=false log --max-count=1 --pretty=format:"%H")
GIT_SHA := $(shell git log -1 --format=%h)
GIT_TAG ?= $(shell bash -c 'TAG=$$(git tag | tail -n1); echo "$${TAG:-none}"')

GIT_UNTRACKED_CHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GIT_UNTRACKED_CHANGES),)
	GIT_COMMIT := $(GIT_COMMIT)-dirty
endif

CONTAINER_TAG ?= $(GIT_TAG)
CONTAINER_NAME := $(REGISTRY)/$(NAME):$(CONTAINER_TAG)
TEST_CONTAINER_TAG := "testing"
CONTAINER_NAME_TESTING := $(REGISTRY)/$(NAME):$(TEST_CONTAINER_TAG)

TEST_FILE := "test/test-localhost-remote.yaml"

export NAME REGISTRY BUILD_DATE GIT_MESSAGE GIT_SHA GIT_TAG CONTAINER_TAG CONTAINER_NAME

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

.PHONY: run-in-docker
run-in-docker: ## runs the last build docker image inside docker
	@echo "+ $@"
	set -x ;	docker run -i \
		--net=host \
		--cap-add NET_ADMIN \
		--cap-add NET_RAW \
		${DOCKER_ARGS} \
		-v ~/.config/gcloud/:/root/.config/gcloud/ \
		-v ~/.ssh/:/tmp/.ssh/:ro \
		-v ~/.kube/:/root/.kube:ro \
		-v $(shell readlink -f ~/.ssh/config):/tmp/ssh-config:ro \
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
test: test-deploy ## build, test, and push container, then run local tests
	@echo "+ $@"
	make build push CONTAINER_TAG="$(TEST_CONTAINER_TAG)" \
		&& ./netassert \
			--image ${CONTAINER_NAME_TESTING} \
			test/test-all.yaml \
		&& make run-in-docker \
			CONTAINER_NAME=$(CONTAINER_NAME_TESTING) \
			ARGS='netassert --image ${CONTAINER_NAME_TESTING} test/test-all.yaml'

.PHONY: test-local
test-local: test-deploy ## test from the local machine
	@echo "+ $@"
	./netassert \
		--image ${CONTAINER_NAME_TESTING} \
		test/test-all.yaml

.PHONY: test-deploy
test-deploy: ## deploy test services
	@echo "+ $@"
	set -x; for DEPLOYMENT_TYPE in \
    frontend \
    microservice \
    database \
    ; do \
  	\
    DEPLOYMENT="test-$${DEPLOYMENT_TYPE}"; \
    kubectl run "$${DEPLOYMENT}" \
      --image=busybox \
      --labels=app=web,role="$${DEPLOYMENT_TYPE}" \
      --requests='cpu=10m,memory=32Mi' \
      --expose \
      --port 80 \
      -- sh -c "while true; do { printf 'HTTP/1.1 200 OK\r\n\n I am a $${DEPLOYMENT_TYPE}\n'; } | nc -l -p  80; done"; \
  	\
    kubectl scale deployment "$${DEPLOYMENT}" --replicas=3; \
  done; \
  \
  kubectl apply -f resource/net-pol/web-deny-all.yaml -f resource/net-pol/test-services-allow.yaml;


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
