# RethinkDB Operator

A Kubernetes operator to manage RethinkDB instances.

## Overview

This Operator is based on the Operator Framework and is used for managing
RethinkDB clusters on Kubernetes.

## Usage

#### Deploy RethinkDB Operator

The first step is to install the RethinkDB Operator in the Kubernetes cluster.
The `deploy` directory contains the manifests needed to properly install the
Operator.

```
kubectl apply -f deploy
```

#### Create RethinkDB Cluster

Once the Operator is deployed and running, we can create an example RethinkDB
cluster. The `example` directory contains several example manifests for creating
RethinkDB clusters using the Operator.

```
kubectl apply -f example/rethinkdb-minimal.yaml
```

#### Destroy RethinkDB Cluster

Simply delete the `RethinkDB` Custom Resource to remove the cluster.

```
kubectl delete -f example/rethinkdb-minimal.yaml
```

#### Persistent Volumes

The RethinkDB Operator supports the use of Persistent Volumes for each node in
the RethinkDB cluster. See (rethinkdb-custom.yaml)[example/rethinkdb-custom.yaml]
for the syntax to enable.

```
kubectl apply -f example/rethinkdb-custom.yaml
```

Delete a RethinkDB cluster using Persistent Volumes. Remember to remove the
left-over volumes when the cluster is no longer needed, as these will not be
removed automatically when the RethinkDB cluster is deleted.

```
kubectl delete rethinkdb,pvc -l cluster=rethinkdb-custom-example
```

## Development

Build the operator and push the new image, remember to update the version.

```
./hack/release
```

When using minikube for local testing, it may be necessary to increase the resources.

```
minikube start --cpus 2 --memory 8192 --disk-size 40g
```

## Testing

Run the tests.

```
./hack/test
```

Generate the code coverage report.

```
./hack/cover
```
