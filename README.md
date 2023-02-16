# Custom Resource practice

With this repository you can define a custom resource with the following fields:

replica - the replica count of the pods

host - the domain on which you want the application to be available

image - the image to use for your application

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

### [cert manager](https://cert-manager.io)

Cert manager provides HTTP access to our server.
To apply the manifests run the following script.

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
```

### cluster issuer, crd, controller

Install a cluster, CRD and the controller on the cluster.

```bash
kubectl apply -f https://raw.githubusercontent.com/adamtagscherer/crd-practice/main/config/release.yaml
```

## Usage

This is a sample custom resource which you can install on your cluster.

```bash
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

```

If you want to access the pods from your browser you have to add a new entry for `palacsinta.org` to your `/etc/hosts` file.
