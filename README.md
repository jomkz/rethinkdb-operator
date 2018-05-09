# RethinkDB Operator

A Kubernetes operator to manage RethinkDB instances.

## Overview

This Operator is built using the [Operator SDK](https://github.com/operator-framework/operator-sdk), which is part of the [Operator Framework](https://github.com/operator-framework/) and manages one or more RethinkDB instances deployed on Kubernetes.

## Usage

The first step is to deploy the RethinkDB Operator into the cluster where it
will watch for requests to create `RethinkDB` resources, much like the native
Kubernetes Deployment Controller watches for Deployment resource requests.

#### Deploy RethinkDB Operator

The `deploy` directory contains the manifests needed to properly install the
Operator.

```
kubectl apply -f deploy
```

You can watch the list of pods and wait until the Operator pod is in a Running
state, it should not take long.

```
kubectl get pods -wl name=rethinkdb-operator
```

You can have a look at the logs for troubleshooting if needed.

```
kubectl logs -l name=rethinkdb-operator
```

Once the RethinkDB Operator is deployed, Have a look in the `examples` directory for example manifests that create `RethinkDB` resources.

#### Create RethinkDB Cluster

Once the Operator is deployed and running, we can create an example RethinkDB
cluster. The `example` directory contains several example manifests for creating
RethinkDB clusters using the Operator.

```
kubectl apply -f example/rethinkdb-minimal.yaml
```

Watch the list of pods to see that each requested node starts successfully.

```
kubectl get pods -wl cluster=rethinkdb-minimal-example
```

#### Destroy RethinkDB Cluster

Simply delete the `RethinkDB` Custom Resource to remove the cluster.

```
kubectl delete -f example/rethinkdb-minimal.yaml
```

#### Persistent Volumes

The RethinkDB Operator supports the use of Persistent Volumes for each node in
the RethinkDB cluster. See [rethinkdb-custom.yaml](example/rethinkdb-custom.yaml)
for the syntax to enable.

```
kubectl apply -f example/rethinkdb-custom.yaml
```

When deleting a RethinkDB cluster that uses Persistent Volumes, remember to
remove the left-over volumes when the cluster is no longer needed, as these will
not be removed automatically.

```
kubectl delete rethinkdb,pvc -l cluster=rethinkdb-custom-example
```

## Development

Clone the repository to a location on your workstation, generally this should be in someplace like `$GOPATH/src/github.com/ORG/REPO`.

Navigate to the location where the repository has been cloned and install the dependencies.

```
cd YOUR_REPO_PATH
dep ensure
```

#### Minikube

When using minikube for local development and testing, it may be necessary to increase the resources for the minikube VM.

```
minikube start --cpus 2 --memory 8192 --disk-size 40g
```

#### Testing

Run the tests.

```
./hack/test
```

Generate the code coverage report.

```
./hack/cover
```

#### Release

Build the operator and push the new image, remember to update the version.

```
./hack/release
```
