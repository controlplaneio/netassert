Fluxcd Demo

doc to be finished


docker run -d -p 5000:5000 --restart=always --name registry-5000 registry:2
kind create cluster --config kind-cluster.yaml


kubectl apply -f https://github.com/fluxcd/flux2/releases/download/v2.7.2/install.yaml

update chart version
helm package ./helm -d .
helm push ./fluxcd-demo-0.0.1-dev.tgz oci://localhost:5000/fluxcd/
kubectl get helmreleases


kind delete cluster