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

package rethinkdb

import (
	"context"
	"log"

	operatorv1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
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
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new RethinkDB Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRethinkDB{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rethinkdb-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RethinkDB
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.RethinkDB{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner RethinkDB
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.RethinkDB{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileRethinkDB{}

// ReconcileRethinkDB reconciles a RethinkDB object
type ReconcileRethinkDB struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a RethinkDB object and makes changes based on the state read
// and what is in the RethinkDB.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRethinkDB) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling RethinkDB %s/%s\n", request.Namespace, request.Name)

	// Fetch the RethinkDB instance
	instance := &operatorv1alpha1.RethinkDB{}
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

// labelsForCluster returns the labels for all cluster resources.
func labelsForCluster(cr *operatorv1alpha1.RethinkDB) map[string]string {
	return map[string]string{
		"app":     "rethinkdb",
		"cluster": cr.Name,
	}
}

// reconcilePods ensures the requested number of Pods are created.
func (r *ReconcileRethinkDB) reconcilePods(cr *operatorv1alpha1.RethinkDB) error {
	// Set the default values for properties not specified in the manifest
	cr.SetDefaults()

	// List the pods for the RethinkDB cluster
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForCluster(cr))
	listOps := &client.ListOptions{Namespace: cr.Namespace, LabelSelector: labelSelector}
	err := r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return err
	}

	if int32(len(podList.Items)) < cr.Spec.Size {
		// Ensure all existing Pods have an IP address before creating a new Pod.
		for _, pod := range podList.Items {
			if len(pod.Status.ContainerStatuses) <= 0 || !pod.Status.ContainerStatuses[0].Ready {
				log.Println("Waiting for existing Pods to become ready...")
				return nil
			}
		}

		log.Printf("Creating a new Pod")
		pod := newPod(cr, podList.Items)

		// Set RethinkDB instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, pod, r.scheme); err != nil {
			return err
		}

		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return err
		}

		// Pod created successfully
		return nil
	}

	log.Printf("Skip reconcile: %d Pods already exist", cr.Spec.Size)
	return nil
}

// reconcileService ensures the Service is created.
func (r *ReconcileRethinkDB) reconcileService(cr *operatorv1alpha1.RethinkDB) error {
	// Check if this Service already exists
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Printf("Creating a new Service %s/%s", cr.Namespace, cr.Name)
		svc := newService(cr)

		// Set RethinkDB instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, svc, r.scheme); err != nil {
			return err
		}

		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return err
		}

		// Service created successfully
		return nil
	} else if err != nil {
		return err
	}

	log.Printf("Skip reconcile: Service %s/%s already exists", found.Namespace, found.Name)
	return nil
}
