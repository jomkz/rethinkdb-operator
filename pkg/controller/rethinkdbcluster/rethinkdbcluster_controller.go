// Copyright 2018 The rethinkdb-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rethinkdbcluster

import (
	"context"

	rethinkdbv1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/rethinkdb/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_rethinkdbcluster")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new RethinkDBCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRethinkDBCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rethinkdbcluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RethinkDBCluster
	err = c.Watch(&source.Kind{Type: &rethinkdbv1alpha1.RethinkDBCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner RethinkDBCluster
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rethinkdbv1alpha1.RethinkDBCluster{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Service and requeue the owner RethinkDBCluster
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rethinkdbv1alpha1.RethinkDBCluster{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRethinkDBCluster{}

// ReconcileRethinkDBCluster reconciles a RethinkDBCluster object
type ReconcileRethinkDBCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a RethinkDBCluster object and makes changes based on the state read
// and what is in the RethinkDBCluster.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRethinkDBCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("reconciling RethinkDBCluster...")

	// Fetch the RethinkDB instance
	instance := &rethinkdbv1alpha1.RethinkDBCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Set the default values for properties not specified in the manifest
	setDefaults(instance)

	// Reconcile the cluster pods
	err = r.reconcilePods(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Reconcile the cluster service
	err = r.reconcileService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// reconcilePods ensures the requested number of Pods are created.
func (r *ReconcileRethinkDBCluster) reconcilePods(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
	members, err := r.listMembers(cr)
	if err != nil {
		return err
	}
	memberCount := int32(len(members))

	if memberCount < cr.Spec.Size {
		// Ensure all existing Pods are running before adding a new Pod.
		for _, pod := range members {
			if pod.Status.Phase != corev1.PodRunning {
				log.Info("waiting for existing server pods to become ready...")
				return nil
			}
		}
		return r.addMember(cr, members)
	} else if memberCount > cr.Spec.Size {
		return r.removeMember(cr, members)
	}

	log.Info("skip reconcile: correct number of server pods exist for the cluster", "size", memberCount)
	return nil
}

// reconcileService ensures the Service is created.
func (r *ReconcileRethinkDBCluster) reconcileService(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("creating a new service", "service", cr.Name)
		svc := newService(cr)

		// Set RethinkDB instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, svc, r.scheme); err != nil {
			return err
		}

		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return err
		}

		// Service created successfully, update status
		cr.Status.ServiceName = svc.Name
		err = r.client.Status().Update(context.TODO(), cr)
		if err != nil {
			return err
		}

		return nil
	} else if err != nil {
		return err
	}

	log.Info("skip reconcile: service already exists", "service", found.Name)
	return nil
}

// addMember will add a new Pod to the cluster.
func (r *ReconcileRethinkDBCluster) addMember(cr *rethinkdbv1alpha1.RethinkDBCluster, members []corev1.Pod) error {
	log.Info("adding a new server pod")
	pod := newPod(cr, members)

	// Set RethinkDB instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, pod, r.scheme); err != nil {
		return err
	}

	err := r.client.Create(context.TODO(), pod)
	if err != nil {
		return err
	}

	// Pod created successfully, update status
	cr.Status.Servers = append(cr.Status.Servers, pod.Name)
	log.Info("Updating status...")
	err = r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		return err
	}

	return nil
}

// listMembers will return a slice containing the Pods in the cluster.
func (r *ReconcileRethinkDBCluster) listMembers(cr *rethinkdbv1alpha1.RethinkDBCluster) ([]corev1.Pod, error) {
	// List the pods for the RethinkDB cluster
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForCluster(cr))
	listOps := &client.ListOptions{Namespace: cr.Namespace, LabelSelector: labelSelector}
	err := r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		log.Error(err, "failed to list server pods")
		return nil, err
	}
	return podList.Items, nil
}

// removeMember will delete a Pod from the cluster. The first Pod in the provided slice of Members will be deleted.
func (r *ReconcileRethinkDBCluster) removeMember(cr *rethinkdbv1alpha1.RethinkDBCluster, members []corev1.Pod) error {
	if len(members) <= 0 {
		return nil
	}

	pod := members[0]
	log.Info("removing existing server pod", "pod", pod.Name)

	err := r.client.Delete(context.TODO(), &pod)
	if err != nil {
		return err
	}

	// Pod deleted successfully, update status
	members = append(members[:0], members[1:]...)
	cr.Status.Servers = make([]string, 0)
	for _, pod := range members {
		cr.Status.Servers = append(cr.Status.Servers, pod.Name)
	}

	err = r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		return err
	}

	return nil
}
