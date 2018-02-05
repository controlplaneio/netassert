# netassert

`netassert`: network testing for humans

**NOTE:** this framework is in beta state as we move towards our first 1.0 release. Please file any issues you find and note the version used.

This project is thin, distributed wrapper around nmap, ssh, and kubectl. It asserts the state of ports on hosts, with traffic originating from anywhere the test runner can access via SSH.

## CLI

```bash
$ ./netassert --help
Usage: netassert [options] [filename]

Options:
  --offline           Assume image is already on target nodes
  --image             Name of test image (for private/offline registies)

  --debug             More debug

  -h --help           Display this message
```

## configuration

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
    8.8.8.8: UDP:53 # host and ports to test
    google.co.uk: 443 # if no protocol is specified then TCP is implied
    control-plane.io: 80, 81, 443, 22 # ports can be comma- or space- delimited
    kubernetes.io: # this can be anything SSH can access
      - 443 # ports can be provided as a list
      - 80
    localhost: # this tests ports on the local machine
      - 22
      - -999 # ports can be negated with `-`
      - -TCP:30731 # TCP is implied, but can be specified
      - -UDP:1234 # UDP ports must be specified
      - -UDP:555

  control-plane.io: # this must be accessible via ssh (perhaps via ssh-agent), or `localhost`
    8.8.8.8: UDP:53 # this tests 8.8.8.8:53 from control-plane.io
    8.8.4.4: UDP:53
    google.com: 443


k8s: # used for Kubernetes pods
  deployment: # only deployments currently supported
    test-frontend: # pod name, defaults to `default` namespace
      test-microservice: 80 # `test-microservice` is the DNS name of the target service
      test-database: -80

    new-namespace:test-microservice: # `new-namespace` is the namespace name
      test-database.new-namespace: 80 # longer DNS names can be used for other namespaces
      test-frontend.default: 80

    default:test-database:
      test-frontend.default.svc.cluster.local: 80 # full DNS names can be used
      test-microservice.default.svc.cluster.local: -80


```


### test from localhost

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

### test from a remote server

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


### test localhost can reach a remote server, and that the remote server can reach another host

```yaml
host:
  localhost:
    control-plane.io:
      - 22
  control-plane.io:
    github.com:
      - 22
```

### test Kubernetes pod

Test that a pod can reach `8.8.8.8`:

```yaml
k8s:
  deployment:
    some-namespace:my-pod:
      8.8.8.8: UDP:53
```

### test Kubernetes pods' intercommunication

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


# example

### deploy our micro-microservices
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

### run netassert - this should fail

```bash
./netassert test/test-k8s.yaml
```

As we haven't applied network policies, this will fail


###

### apply network policies
```
kubectl apply -f resource/net-pol/web-deny-all.yaml
kubectl apply -f resource/net-pol/test-services-allow.yaml
```


### manually test the pods
```
kubectl exec -it test-frontend-5cc944689f-rzv4f -- wget -qO- --timeout=2 http://test-microservice
kubectl exec -it test-microservice-5cc944689f-rzv4f -- wget -qO- --timeout=2 http://test-database
kubectl exec -it test-database-5cc944689f-rzv4f -- wget -qO- --timeout=2 http://test-frontend
```
