default:
    just --list

version := "0.0.1"

# build the binary in ./bin folder
build:
	go build -o bin/netassert cmd/netassert/cli/*.go

# build and run the binary
run: build
	bin/netassert

# run go test(s)
test:
	go test -v -race ./...

# run the linter
lint:
	golangci-lint run ./...

# remove the binary from ./bin folder
clean:
	@rm -rf ./bin

# create a new kind k8s cluster called packet-test
kind-up:
	kind create cluster --name packet-test --config ./e2e/clusters/kind/kind-config.yaml

# delete the kind k8s cluster called packet-test
kind-down:
	kind delete clusters packet-test
# deployObj kubernetes manifests
k8s-apply:
	kubectl apply -f ./e2e/manifests/workload.yaml

# delete kubernetes deployObj
k8s-rm-apply:
	kubectl delete -f ./e2e/manifests/workload.yaml

# build docker image and tag it 0.0.01
docker-build:
	docker build -f Dockerfile --no-cache --tag packet-capture:{{version}} .

# import image into the local kind cluster called packet-test
kind-import-image:
    kind load docker-image packet-capture:{{version}} --name packet-test && kind load docker-image netassert-client:{{version}} --name packet-test

