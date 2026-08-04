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

k8s-rm-apply:
	kubectl delete -f ./e2e/manifests/workload.yaml

netpol-apply:
	kubectl apply -f ./e2e/manifests/networkpolicies.yaml

netpol-rm-apply:
	kubectl delete -f ./e2e/manifests/networkpolicies.yaml

calico-apply:
	kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.31.3/manifests/calico.yaml

calico-rm-apply:
	kubectl delete -f https://raw.githubusercontent.com/projectcalico/calico/v3.31.3/manifests/calico.yaml

docker-build:
	docker build -f Dockerfile \
		--no-cache \
		--build-arg VERSION={{version}} \
		--build-arg SCANNER_IMG_VERSION=latest \
		--build-arg SNIFFER_IMG_VERSION=latest \
		--tag netassert-local:{{version}} .
