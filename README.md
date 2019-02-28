# RethinkDB Operator

A Kubernetes operator to manage RethinkDB clusters.

## Overview

This Operator is built using the [Operator SDK](https://github.com/operator-framework/operator-sdk), which is part of the [Operator Framework](https://github.com/operator-framework/) and manages one or more RethinkDB instances deployed on Kubernetes.

## Usage

The first step is to deploy the RethinkDB Operator into the cluster where it
will watch for requests to create `RethinkDBCluster` resources, much like the native
Kubernetes Deployment Controller watches for Deployment resource requests.

#### Deploy RethinkDB Operator

The `deploy` directory contains the manifests needed to properly install the
Operator.

Create the service account for the operator.

```
kubectl create -f deploy/service_account.yaml
```

Next, create the RBAC role and role-binding that grants the permissions
necessary for the operator to function.

```
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
```

Add the CRD to the cluster that defines the RethinkDB resource.

```
kubectl create -f deploy/crds/rethinkdb_v1alpha1_rethinkdbcluster_crd.yaml
```

Finally, deploy the operator into the cluster.

```
$ kubectl create -f deploy/operator.yaml
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

Local development is usually done with [minikube](https://github.com/kubernetes/minikube) or [minishift](https://www.okd.io/minishift/).

#### Minikube

When using minikube for local development and testing, it may be necessary to increase the resources for the minikube VM.

```
minikube start --cpus 2 --memory 8192 --disk-size 40g
```

#### Source Code

Clone the repository to a location on your workstation, generally this should be in someplace like `$GOPATH/src/github.com/ORG/REPO`.

Navigate to the location where the repository has been cloned and install the dependencies.

```
cd YOUR_REPO_PATH
dep ensure && dep status
```

#### Run Locally

Once the dependencies are present, ensure the service account, role, role binding and CRD are added to your local cluster.

```
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/crds/rethinkdb_v1alpha1_rethinkdbcluster_crd.yaml
```

Once the CRD is present, we can start the operator locally and begin development.

```
operator-sdk up local
```

You can now create a RethinkDB resource to test your changes.

```
kubectl create -f example/rethinkdb-minimal.yaml
```

Keep in mind that when you make changes to the code, you must restart the operator. Use `Ctrl+c` to kill the process and restart.

## License

RethinkDB Operator is released under the Apache 2.0 license. See the [LICENSE][license_file] file for details.

[license_file]:./LICENSE
