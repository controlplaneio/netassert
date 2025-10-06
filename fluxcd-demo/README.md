# 🚀 FluxCD Demo Guide

This guide walks you through setting up a **FluxCD** demo environment using **kind** (Kubernetes in Docker) and a **local Helm chart registry**.  
You’ll see how Flux automates Helm releases and how to observe its reconciliation behavior in action while running tests with **NetAssert**.

---

## 🧰 Prerequisites

Before starting, make sure you have the following tools installed:

- [Docker](https://docs.docker.com/get-docker/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kind](https://kind.sigs.k8s.io/)
- [Helm](https://helm.sh/docs/intro/install/)

---

## 🏗️ Step 1: Set Up the Environment

### 1.1 Start a Local Docker Registry

FluxCD can work with OCI-based Helm registries. Start a local Docker registry to host your Helm charts:

```bash
docker run -d -p 5000:5000 --restart=always --name registry-5000 registry:2
```

This creates a local registry accessible at `localhost:5000`.

---

### 1.2 Create a Kind Cluster

Create a local Kubernetes cluster using your configuration file:

```bash
kind create cluster --config kind-cluster.yaml
```

Once complete, verify the cluster is ready:

```bash
kubectl cluster-info
kubectl get nodes
```

---

## ⚙️ Step 2: Install FluxCD

Refer to the official documentation for detailed installation instructions:  
👉 [FluxCD Installation Guide](https://fluxcd.io/flux/installation/)

For this demo, you can use the following command:

```bash
kubectl apply -f https://github.com/fluxcd/flux2/releases/download/v2.7.2/install.yaml
```

Verify that FluxCD is running:

```bash
kubectl get pods -n flux-system
```

Expected output should include components like:

```
helm-controller
kustomize-controller
notification-controller
source-controller
```

All should reach the `Running` state.

---

## 📦 Step 3: Package and Push the Helm Chart

### 3.1 Update Chart Versions

Before packaging, update the NetAssert subchart to a version available in the packages section of this repo.  

---

### 3.2 Package the Helm Chart

Run the following command to package your chart into a `.tgz` archive:

```bash
helm package ./helm -d .
```

This produces a packaged chart file, for example:

```
./fluxcd-demo-0.0.1-dev.tgz
```

---

### 3.3 Push the Chart to the Local Registry

Push the packaged Helm chart to your local OCI registry:

```bash
helm push ./fluxcd-demo-0.0.1-dev.tgz oci://localhost:5000/fluxcd/
```

---

### 3.4 Apply the FluxCD configs

Apply the fluxcd-helmconfig.yaml file so FluxCD can release the charts:

```bash
kubectl apply -f fluxcd-helmconfig.yaml
```

---

## 🔄 Step 4: Watch Flux Reconcile the Release with NetAssert Tests

Flux continuously monitors and applies Helm releases defined in your cluster.  
To observe its behavior, list Helm releases managed by Flux:

```bash
kubectl get helmreleases
```

Flux will automatically pull your Helm chart from the registry and apply it.

---

### 🧩 What to Observe

- The **init container** in your k8s deployment object intentionally delay completion.  
- The **Netassert** job will not be created until the deployment finishes.  
- Once the deployments completes, Netassert will start running as a Job, and once finished it is going to make the release marked as successful or failed.

---

## 🔁 Step 5: Demonstrate an Upgrade

You can simulate a Helm chart upgrade to observe Flux’s automated update handling.

1. **Update chart version** — bump your chart version.  
2. **Repackage** the chart:

   ```bash
   helm package ./helm -d .
   ```

3. **Push** the new version to the registry:

   ```bash
   helm push ./fluxcd-demo-0.0.2-dev.tgz oci://localhost:5000/fluxcd/
   ```

4. **Watch** Flux detect and reconcile the new version:

   ```bash
   kubectl get helmreleases -w
   ```

You’ll see Flux automatically roll out the new chart and update your resources in place, and then run the NetAssert tests.
