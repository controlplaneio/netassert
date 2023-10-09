# Netassert

[![Testing Workflow][testing_workflow_badge]][testing_workflow_badge]
[![Release Workflow][release_workflow_badge]][release_workflow_badge]

`NetAssert` is a command line tool that enables you to check the network connectivity between Kubernetes objects such as Pods, Deployments, DaemonSets, and StatefulSets, as well as test their connectivity to remote hosts or IP addresses. `NetAssert` v2 is a rewrite of original `NetAssert` tool in Go that utilises the ephemeral container support in Kubernetes to verify network connectivity. `NetAssert` test(s) are defined in YAML format. `NetAssert` **currently supports TCP and UDP protocols**:

- To perform a TCP test, only a [`scanner`](https://github.com/controlplaneio/netassertv2-l4-client) container is used. This container requires no privileges nor any Linux capabilities.

- To run a UDP test, a [`sniffer`](https://github.com/controlplaneio/netassertv2-packet-sniffer) ephemeral container is injected into the target Pod which requires `cap_raw` capabilities to read data from the network interface. During UDP testing, `NetAssert` runs both container `scanner` and `sniffer` container images which are injected as `ephemeral` containers into running Pods.

The [`sniffer`](https://github.com/controlplaneio/netassertv2-packet-sniffer) and [`scanner`](https://github.com/controlplaneio/netassertv2-l4-client)  container images can be downloaded from:

- `docker.io/controlplane/netassertv2-l4-client:latest`
  - Used for both TCP and UDP testing and acts as a Layer 4 (TCP/UDP) client
  - Requires no privileges nor any Linux capabilities.
- `docker.io/controlplane/netassertv2-packet-sniffer:latest`
  - Used for UDP testing only, injected at the destination to capture packet and search for specific string in the payload
  - requires `cap_raw` capabilities to read data from the network interface

`NetAssert` utilises the above containers during test and configures them using *environment variables*. The list of environment variables that are used can be found [here](https://github.com/controlplaneio/netassertv2-packet-sniffer) and [here](https://github.com/controlplaneio/netassertv2-l4-client). It is possible to override the `sniffer` and `scanner` images from command line during a run, so one can also bring their own container image(s) as long as they support the same environment variables.

<img src="./img/demo.gif">

## Installation

Download the latest stable version of `NetAssert` from the [releases](https://github.com/controlplaneio/netassert/releases) page

## Test specification

`NetAssert` v2 tests are written in YAML format. Each test is a YAML document which supports the following mappings:

- A YAML document is a list of `NetAssert` test. Each test has the following keys:
  - **name**: a scalar representing the name of the connection
  - **type**: a scalar representing the type of connection, only "k8s" is supported at this time
  - **protocol**: a scalar representing the protocol used for the connection, which must be "tcp" or "udp"
  - **targetPort**: an integer scalar representing the target port used by the connection
  - **timeoutSeconds**: an integer scalar representing the timeout for the connection in seconds
  - **attempts**: an integer scalar representing the number of connection attempts for the test
  - **exitCode**: an integer scalar representing the expected exit code from the ephemeral/debug container(s)
  - **src**: a mapping representing the source Kubernetes resource, which has the following keys:
    - **k8sResource**: a mapping representing a Kubernetes resource with the following keys:
      - **kind**: a scalar representing the kind of the Kubernetes resource, which can be `deployment`, `statefulset`, `daemonset` or `pod`
      - **name**: a scalar representing the name of the Kubernetes resource
      - **namespace**: a scalar representing the namespace of the Kubernetes resource
  - **dst**: a mapping representing the destination Kubernetes resource or host, **which can have one of the the following keys** i.e both `k8sResource` and `host` **are not supported at the same time** :
    - **k8sResource**: a mapping representing a Kubernetes resource with the following keys:
      - **kind**: a scalar representing the kind of the Kubernetes resource, which can be `deployment`, `statefulset`, `daemonset` or `pod`
      - **name**: a scalar representing the name of the Kubernetes resource
      - **namespace**: a scalar representing the namespace of the Kubernetes resource. (Note: Only allowed when protocol is "tcp")
    - **host**: a mapping representing a host/node with the following key:
      - **name**: a scalar representing the name or IP address of the host/node. (Note: Only allowed when protocol is "tcp" or "udp", but not both at the same time)

<details><summary>This is an example of a test that can be consumed by `NetAssert` utility</summary>

```yaml
---
- name: busybox-deploy-to-echoserver-deploy
  type: k8s
  protocol: tcp
  targetPort: 8080
  timeoutSeconds: 67
  attempts: 3
  exitCode: 0
  src:
    k8sResource:
      kind: deployment
      name: busybox
      namespace: busybox
  dst:
    k8sResource:
      kind: deployment
      name: echoserver
      namespace: echoserver
#######
#######
- name: busybox-deploy-to-core-dns
  type: k8s
  protocol: udp
  targetPort: 53
  timeoutSeconds: 67
  attempts: 3
  exitCode: 0
  src:
    k8sResource:
      kind: deployment
      name: busybox
      namespace: busybox
  dst:
    k8sResource:
      kind: deployment
      name: coredns
      namespace: kube-system
######
######
- name: busybox-deploy-to-web-statefulset
  type: k8s
  protocol: tcp
  targetPort: 80
  timeoutSeconds: 67
  attempts: 3
  exitCode: 0
  src:
    k8sResource: # this is type endpoint
      kind: deployment
      name: busybox
      namespace: busybox
  dst:
    k8sResource: ## this is type endpoint
      kind: statefulset
      name: web
      namespace: web
###
###
- name: fluentd-daemonset-to-web-statefulset
  type: k8s
  protocol: tcp
  targetPort: 80
  timeoutSeconds: 67
  attempts: 3
  exitCode: 0
  src:
    k8sResource: # this is type endpoint
      kind: daemonset
      name: fluentd
      namespace: fluentd
  dst:
    k8sResource: ## this is type endpoint
      kind: statefulset
      name: web
      namespace: web
###
####
- name: busybox-deploy-to-control-plane-dot-io
  type: k8s
  protocol: tcp
  targetPort: 80
  timeoutSeconds: 67
  attempts: 3
  exitCode: 0
  src:
    k8sResource: # type endpoint
      kind: deployment
      name: busybox
      namespace: busybox
  dst:
    host: # type host or node or machine
      name: control-plane.io
###
###
- name: test-from-pod1-to-pod2
  type: k8s
  protocol: tcp
  targetPort: 80
  timeoutSeconds: 67
  attempts: 3
  exitCode: 0
  src:
    k8sResource: ##
      kind: pod
      name: pod1
      namespace: pod1
  dst:
    k8sResource:
      kind: pod
      name: pod2
      namespace: pod2
###
###
- name: busybox-deploy-to-fake-host
  type: k8s
  protocol: tcp
  targetPort: 333
  timeoutSeconds: 67
  attempts: 3
  exitCode: 1
  src:
    k8sResource: # type endpoint
      kind: deployment
      name: busybox
      namespace: busybox
  dst:
    host: # type host or node or machine
      name: 0.0.0.0
...
```

</details>

## Components

`NetAssert` has three main components:

- [NetAssert](https://github.com/controlplaneio/netassert): This is responsible for orchestrating the tests and is also known as `Netassert-Engine` or simply the `Engine`
- [NetAssertv2-packet-sniffer](https://github.com/controlplaneio/netassertv2-packet-sniffer): This is the sniffer component that is utilised during a UDP test and is injected to the destination/target Pod as an ephemeral container
- [NetAssertv2-l4-client](https://github.com/controlplaneio/netassertv2-l4-client): This is the scanner component that is injected as the scanner ephemeral container onto the source Pod and is utilised during both TCP and UDP tests

## Detailed steps/flow of tests

All the tests are read from an YAML file or a directory (step **1**) and the results are written following the [TAP format](https://testanything.org/) (step **5** for UDP and step **4** for TCP). The tests are performed in two different manners depending on whether a TCP or UDP connection is used

### UDP test

<img src="./img/udp-test.svg">

- Validate the test spec and ensure that the `src` and `dst` fields are correct: for udp tests both of them must be of type `k8sResource`
- Find a running Pod called `dstPod` in the object defined by the `dst.k8sResource` field. Ensure that the Pod is in running state and has an IP address allocated by the CNI
- Find a running Pod called `srcPod` in the object defined by the `src.k8sResource` field. Ensure that the Pod is in running state and has an IP address allocated by the CNI
- Generate a random UUID, which will be used by both ephemeral containers
- Inject the `netassert-l4-client` as an ephemeral container in the `srcPod` (step **2**) and set the port and protocol according to the test specifications. Provide also the target host equal to the previously found dstPod IP address, and the random UUID that was generated in the previous step as the message to be sent over the udp connection. At the same time, inject the `netassertv2-packet-sniffer` (step **3**) as an ephemeral container in the `dstPod` using the protocol, search string, number of matches and timeout defined in the test specifications. The search_string environment variable is equal to the UUID that was generated in the previous step which is expected to be found in the data sent by the scanner when the connections are successful.
- Poll that status of the ephemeral containers (step **4**)
- Ensure that the `netassertv2-packet-sniffer` ephemeral sniffer container’s exit status matches the one defined in the test specification
- Ensure that the `netassert-l4-client`, exits with exit status of zero. This should always be the case as UDP is not a connection oriented protocol.

### TCP test

<img src="./img/tcp-test.svg">

- Validate the test spec and ensure that the `src` field is of type `k8sResource`
- Find a running Pod called `srcPod` in the object defined by the `src.k8sResource` field. Ensure that the Pod is in running state and has an IPAddress
- Check if `dst` has `k8sResource` defined as a child object. If so then find a running Pod defined by the `dst.K8sResource`
- Inject the `netassert-l4-client` as an ephemeral container in the `srcPod` (step **2**). Configure the `netassert-l4-client` similarly to the udp case. If the `dst` field is set to `host` then use the host `name` field as the scanner target host
- Poll that status of the ephemeral containers (step **3**)
- Ensure that the exit code of that container matches the `exitCode` field defined in the test specification

## Development

- You will need Go version 1.20.x or higher. Download the latest version of [just](https://github.com/casey/just/releases). To build the project you can use `just build`. The resulting binary will be in `cmd/netassert/cli/netassert`. To run `unit` tests you can use `just test`. There is a separate [README.md](./e2e/README.md) that details `end-to-end` testing.

## Quick testing

- Make sure you have installed [`kind`](https://kind.sigs.k8s.io/) and its prerequisites
- Make sure you have also installed [`just`](https://github.com/casey/just/releases)
- Download the `NetAssert` binary from the [release](https://github.com/controlplaneio/netassert/releases) page

- If you want to quickly test `NetAssert`, you can make use of the sample test(s) and manifests provided

- You will also need a working kubernetes cluster with ephemeral/debug container support, you can spin one quickly using the `justfile` included in the repo

```bash
❯ just kind-down ; just kind-up
```

- In order to use the sample tests, you need to create kubernetes resources:

```bash
❯ just k8s-apply

  kubectl apply -f ./kubernetes/manifests
  namespace/fluentd created
  daemonset.apps/fluentd created
  namespace/echoserver created
  namespace/busybox created
  deployment.apps/echoserver created
  deployment.apps/busybox created
  namespace/web created
  statefulset.apps/web created
```

- Run the netassert binary pointing it to the test cases:

```bash
❯ ./netassert run --input-file cmd/netassert/cli/test-cases.yaml

❯ cat results.tap
1..4
ok 1 - busybox-deploy-to-echoserver-deploy
ok 2 - busybox-deploy-to-core-dns
ok 3 - test-from-busybox-to-web-statefulset
not ok 4 - test-from-busybox-to-host
  ---
  reason: ephemeral container netassertv2-client-u8dqy3qwo exit code for test test-from-busybox-to-host
  is 1 instead of 0
  ...


```

## Compatibility

Netassert has been tested with the following flavours of Kubernetes:

| K8s Distribution | Version | CNI                             | Working |
|------------------|---------|---------------------------------|---------|
| AWS EKS          | 1.25    | AWS VPC CNI                     | Yes     |
| AWS EKS          | 1.24    | AWS VPC CNI                     | Yes     |
| AWS EKS          | 1.25    | Calico Version 3.25             | Yes     |
| AWS EKS          | 1.24    | Calico version 3.25             | Yes     |
| GCP GKE          | 1.24    | GCP VPC CNI                     | Yes     |
| GCP GKE          | 1.24    | GCP Cilium 1.11 (Dataplane v2)  | Yes     |

## Checking for ephemeral container support

You can check for ephemeral container support using the following command:

```bash
❯ netassert ping
2023-03-27T11:25:28.421+0100 [INFO]  [NetAssert-v2.0.0]: ✅ Successfully pinged /healthz endpoint of the Kubernetes server
2023-03-27T11:25:28.425+0100 [INFO]  [NetAssert-v2.0.0]: ✅ Ephemeral containers are supported by the Kubernetes server
```

## Increasing logging verbosity

You can increase the logging level to `debug` by passing `--log-level` argument:

```bash
❯ netassert run --input-file ./sample-tests/test-cases/test-cases.yaml --log-level=debug
```

## RBAC Configuration

This tool can be run according to the Principle of Least Privilege (PoLP) by properly configuring the RBAC.

The list of required permissions can be found in the `netassert` ClusterRole `kubernetes/rbac/cluster-role.yaml`, which could be redefined as a Role for namespacing reasons if needed. This role can then be bound to a "principal" either through a RoleBinding or a ClusterRoleBinding, depending on whether the scope of the role is supposed to be namespaced or not. The ClusterRoleBinding `kubernetes/rbac/cluster-rolebinding.yaml` is an example where the user `netassert-user` is assigned the role `netassert` using a cluster-wide binding called `netassert`

## Limitations

- When performing UDP scanning, the sniffer container [image](https://github.com/controlplaneio/netassertv2-packet-sniffer) needs `cap_net_raw` capability so that it can bind and read packets from the network interface. As a result, admission controllers or other security mechanisms must be modified to allow the `sniffer` image to run with this capability. Currently, the Security context used by the ephemeral sniffer container looks like the following:

```yaml
...
...
   securityContext:
     allowPrivilegeEscalation: false
     capabilities:
       add:
       - NET_RAW
     runAsNonRoot: true
...
...
```

- Although they do not consume any resources, ephemeral containers that are injected as part of the test(s) by `NetAssert` will remain in the Pod specification

- Service meshes are not be currently supported

## E2E Tests

- Please check this [README.md](./e2e/README.md)

[testing_workflow_badge]: https://github.com/controlplaneio/netassert/workflows/Lint%20and%20Build/badge.svg

[release_workflow_badge]: https://github.com/controlplaneio/netassert/workflows/goreleaser/badge.svg
