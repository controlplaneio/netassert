# netassert

`netassert`: network security testing for DevSecOps workflows

**NOTE:** this framework is in beta state as we move towards our first 1.0 release. Please file any issues you find and note the version used.

This is a security testing framework for fast, safe iteration on firewall, routing, and NACL rules for Kubernetes ([Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/), services) and non-containerised hosts (cloud provider instances, VMs, bare metal). It aggressively parallelises `nmap` to test outbound network connections and ports from any accessible host, container, or Kubernetes pod by joining the same network namespace as the instance under test.

- [netassert](#netassert)
  * [Why?](#why)
  * [CLI](#cli)
- [Example](#example)
    + [Deploy fake mini-microservices](#deploy-fake-mini-microservices)
    + [Run netassert (this should fail)](#run-netassert-this-should-fail)
    + [Apply network policies](#apply-network-policies)
    + [Run netassert (this should pass)](#run-netassert-this-should-pass)
    + [Manually test the pods](#manually-test-the-pods)
- [Configuration](#configuration)
    + [Test outbound connections from localhost](#test-outbound-connections-from-localhost)
    + [Test outbound connections from a remote server](#test-outbound-connections-from-a-remote-server)
    + [Test localhost can reach a remote server, and that the remote server can reach another host](#test-localhost-can-reach-a-remote-server-and-that-the-remote-server-can-reach-another-host)
    + [Test a Kubernetes pod](#test-a-kubernetes-pod)
    + [Test Kubernetes pods' intercommunication](#test-kubernetes-pods-intercommunication)
  * [Example flow for K8S pods](#example-flow-for-k8s-pods)


## Why?

The alternative is to `exec` into a container and `curl`, or spin up new pods with the same selectors and `curl` from there. This has lots of problems (extra tools in container image, or tool installation despite immutable root filesystems, or egress prevention). `netassert` aims to fix this:
- does not rely on a dedicated tool speaking the correct target protocol (e.g. doesn't need `curl`, GRPC client, etc)
- does not bloat the pod under test or increase the pod's attack surface with non-production tooling
- works with `FROM scratch` containers
- is parallelised to run in near-constant time for large or small test suites
- does not appear to the Kubernetes API server that it's changing the system under test
- uses TCP/IP (layers 3 and 4) so does not show up in HTTP logs (e.g. `nginx` access logs)
- produces TAP output for humans and build servers

More information and background in [this presentation](https://www.binarysludge.com/2018/02/05/assertion-or-it-didnt-happen-in-cloud-networks-cfgmgmtcamp-feburary-2018/) from Configuration Management Camp 2018.

## CLI

```
Usage: netassert [options] [filename]

Options:
  --image                Name of test image
  --no-pull              Don't pull test container on target nodes
  --timeout              Integer time to wait before giving up on tests (default 120)
  --ssh-user             SSH user for kubelet host
  --known-hosts          A known_hosts file (default: ${HOME}/.ssh/known_hosts)
  --gcloud-ssh-options   Optional options to pass to the 'gcloud compute ssh' command

  --debug                More debug

  -h --help              Display this message
```

# Example

## Prerequisites on host machine

- `jq`
- `yq`
- `parallel`
- `timeout`

> These will be moved into a container runner in the future

## Prerequisites on target

- `docker`

### Deploy fake mini-microservices
```bash
for DEPLOYMENT_TYPE in \
  frontend \
  microservice \
  database\
  ; do
  DEPLOYMENT="test-${DEPLOYMENT_TYPE}"

  kubectl run "${DEPLOYMENT}" \
    --image=busybox \
    --labels=app=web,role="${DEPLOYMENT_TYPE}" \
    --requests='cpu=10m,memory=32Mi' \
    --expose \
    --port 80 \
    -- sh -c "while true; do { printf 'HTTP/1.1 200 OK\r\n\n I am a ${DEPLOYMENT_TYPE}\n'; } | nc -l -p  80; done"

  kubectl scale deployment "${DEPLOYMENT}" --replicas=3
done
```

### Run netassert (this should fail)

As we haven't applied network policies, this should **FAIL**.

```bash
./netassert test/test-k8s.yaml
```

> Ensure your user has SSH access to the node names listed by `kubectl get nodes`. To change the SSH user set
> `--ssh-user MY_USER`. To configure your ssh keys, use DNS resolvable names (or `/etc/hosts` entries) for the nodes,
> and/or add login directives to `~/.ssh/config`:
> ```bash
> # ~/.ssh/config
> Host node-1
>   HostName 192.168.10.1
>   User sublimino
>   IdentityFile ~/.ssh/node-1-key.pem
> ```

### Apply network policies
```
kubectl apply -f resource/net-pol/web-deny-all.yaml
kubectl apply -f resource/net-pol/test-services-allow.yaml
```

### Run netassert (this should pass)

Now that we've applied the policies that these tests reflect, this should pass:

```bash
./netassert test/test-k8s.yaml
```

For manual verification of the test results we can `exec` and `curl` in the pods under test (see [why] above for reasons that this is a bad idea).

### Manually test the pods
```
kubectl exec -it test-frontend-$YOUR_POD_ID -- wget -qO- --timeout=2 http://test-microservice
kubectl exec -it test-microservice-$YOUR_POD_ID -- wget -qO- --timeout=2 http://test-database
kubectl exec -it test-database-$YOUR_POD_ID -- wget -qO- --timeout=2 http://test-frontend
```

These should all pass as they have equivalent network policies.

The network policies do not allow the `frontend` pods to communicate with the `database` pods.

Let's verify that manually - this should **FAIL**:

```
kubectl exec -it test-frontend-$YOUR_POD_ID -- wget -qO- --timeout=2 http://test-database
```


# Configuration

netassert takes a single YAML file as input. This file lists the hosts to test from, and describes the hosts and ports that it should be able to reach.

It can test from any reachable host, and from Kubernetes pods.

A simple example:

```yaml
host: # used for ssh-accessible hosts
  localhost: # host to run test from, can be anything accessible via SSH
    8.8.8.8: UDP:53 # host and ports to test for access
```

A full example:

```yaml
host: # used for ssh-accessible hosts
  localhost: # host to run test from, can be a remote host
    8.8.8.8: UDP:53 # host and ports to test from localhost
    google.co.uk: 443 # if no protocol is specified then TCP is implied
    control-plane.io: 80, 81, 443, 22 # ports can be comma or space delimited
    kubernetes.io: # this can be anything SSH can access
      - 443 # ports can be provided as a list
      - 80
    localhost: # this tests ports on the local machine
      - 22
      - -999       # ports can be negated with `-`, this checks that 999 TCP is not open
      - -TCP:30731 # TCP is implied, but can be specified
      - -UDP:1234  # UDP must be explicitly stated, otherwise TCP assumed
      - -UDP:555

  control-plane.io: # this must be accessible via ssh (perhaps via ssh-agent), or `localhost`
    8.8.8.8: UDP:53 # this tests 8.8.8.8:53 is accesible from control-plane.io
    8.8.4.4: UDP:53 # this tests 8.8.4.4:53 is accesible from control-plane.io
    google.com: 443 # this tests google.com:443 is accesible from control-plane.io


k8s: # used for Kubernetes pods
  deployment: # only deployments currently supported
    test-frontend: # pod name, defaults to `default` namespace
      test-microservice: 80  # `test-microservice` is the DNS name of the target service
      test-database: -80     # test-frontend should not be able to access test-database port 80

    new-namespace:test-microservice: # `new-namespace` is the namespace name
      test-database.new-namespace: 80 # longer DNS names can be used for other namespaces
      test-frontend.default: 80

    default:test-database:
      test-frontend.default.svc.cluster.local: 80 # full DNS names can be used
      test-microservice.default.svc.cluster.local: -80
```

### Test outbound connections from localhost

To test that `localhost` can reach `8.8.8.8` and `8.8.4.4` on port 53 UDP:

```yaml
host:
  localhost:
    8.8.8.8: UDP:53
    8.8.4.4: UDP:53
```

What this test does:
1. Starts on the test runner host
1. Pull the test container
1. Check port `UDP:53` is open on `8.8.8.8` and `8.8.4.4`
1. Shows TAP results

### Test outbound connections from a remote server

Test that `control-plane.io` can reach `github.com`:

```yaml
host:
  control-plane.io:
    github.com:
      - 22
      - 443
```

What this test does:
1. Starts on the test runner host
1. SSH to `control-plane.io`
1. Pull the test container
1. Check ports `22` and `443` are open
1. Returns TAP results to the test runner host


### Test localhost can reach a remote server, and that the remote server can reach another host

```yaml
host:
  localhost:
    control-plane.io:
      - 22
  control-plane.io:
    github.com:
      - 22
```

### Test a Kubernetes pod

Test that a pod can reach `8.8.8.8`:

```yaml
k8s:
  deployment:
    some-namespace:my-pod:
      8.8.8.8: UDP:53
```

### Test Kubernetes pods' intercommunication

Test that `my-pod` in namespace `default` can reach `other-pod` in `other-namespace`, and that `other-pod` cannot reach
`my-pod`:


```yaml
k8s:
  deployment:
    default:my-pod:
      other-namespace:other-pod: 80

    other-namespace:other-pod:
      default:my-pod: -80
```

## Example flow for K8S pods

1. from test host: `nettest test/test-k8s.yaml`
1. look up deployments, pods, and namespaces to test in Kube API
1. for each pod, SSH to a worker node running an instance
1. connect a test container to the container's network namespace
1. run that pod's test suite from inside the network namespace
1. report results via TAP
1. test host gathers TAP results and reports
1. the same process applies to non-Kubernetes instances accessible via ssh


