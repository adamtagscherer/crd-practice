# Custom Resource practice

With this repository you can define a custom resource with the following fields:

replicas - the replica count of the pods

host - the domain on which you want the application to be available

image - the image to use for your application

After installing everything you should be able to visit the domain you specified through HTTP and HTTPS protocols.

## Prerequisites

- [go](https://go.dev/dl) - version v1.19.0+
- [docker](https://docs.docker.com/get-docker) - version 17.03+
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) - version v1.11.3+
- [kind](https://kind.sigs.k8s.io)

## Installation

### [kind cluster](https://kind.sigs.k8s.io/docs/user/ingress)

First create a kind cluster with the following script.
This script will expose the ports 80 and 443 to your host machine so you will be able to direct traffic from your machine to your cluster.

```bash
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
```
### [nginx ingress controller](https://docs.nginx.com/nginx-ingress-controller)

You should install an nginx ingress controller to control traffic to you services.

To apply the manifests of the controller run the following script.

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```

Wait for the installation to finish.

```bash
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

### [cert manager](https://cert-manager.io)

Cert manager provides HTTPS access to the server.
To apply the manifests run the following script.

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
```

Wait for the installation to finish.
You will need [cmctl](https://cert-manager.io/docs/reference/cmctl/#installation) for this.

```bash
cmctl check api --wait=2m
```

### cluster issuer, crd, controller manager

Install a cluster issuer, CRD and the controller manager on the cluster with the following script.

```bash
kubectl apply -f https://raw.githubusercontent.com/adamtagscherer/crd-practice/main/config/release.yaml
```

## Usage

This is a sample custom resource which you can install on your cluster.

```bash
cat <<EOF | kubectl apply -f -
apiVersion: webapp.my.domain/v1
kind: Guestbook
metadata:
  labels:
    app.kubernetes.io/name: guestbook
    app.kubernetes.io/instance: guestbook-sample
    app.kubernetes.io/part-of: guestbook
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: guestbook
  name: guestbook-sample
spec:
  replicas: 2
  host: palacsinta.org
  image: nginx:latest
EOF
```

If you want to access the pods from your browser you have to add a new entry for `palacsinta.org` to your `/etc/hosts` file.

You can apply the same CRD with other parameters and the controller will alter the state of the cluster to match your yaml.

## Improvements

- add a webhook to validate what objects the user applies and to give it default values if they are not given
- add metrics endpoint to the pods and the controller
- add e2e tests to test if everything is set up properly
