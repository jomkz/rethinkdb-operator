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
	"fmt"

	rethinkdbv1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/rethinkdb/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
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

// Add creates a new RethinkDBCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRethinkDBCluster{client: mgr.GetClient(), config: mgr.GetConfig(), scheme: mgr.GetScheme()}
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

	// Watch for changes to secondary resource ConfigMap and requeue the owner RethinkDBCluster
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &rethinkdbv1alpha1.RethinkDBCluster{},
	})
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

	// Watch for changes to secondary resource Secret and requeue the owner RethinkDBCluster
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
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
	// that reads objects from the cache and writes to the apiserver.
	// The REST configuration has been added for use by SDK TLS functionality
	client client.Client
	config *rest.Config
	scheme *runtime.Scheme
}

// Reconcile compares the actual state of the cluster to the desired state
// for the given RethinkDBCluster request.
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRethinkDBCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("namespace", request.Namespace, "name", request.Name)
	reqLogger.Info("reconciling RethinkDBCluster")

	// Fetch the RethinkDBCluster instance
	cluster := &rethinkdbv1alpha1.RethinkDBCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Verify resource defaults have been initialized
	if setDefaults(cluster) {
		// defaults updated, update and requeue
		reqLogger.Info("default spec values initialized")
		return reconcile.Result{Requeue: true}, r.client.Update(context.TODO(), cluster)
	}

	// Reconcile the cluster CA secret
	caSecret, err := r.reconcileCASecret(cluster)
	if err != nil {
		reqLogger.Error(err, "unable to reconcile ca secret")
		return reconcile.Result{}, err
	}

	// Reconcile the cluster CA configmap
	err = r.reconcileCAConfigMap(cluster, caSecret)
	if err != nil {
		reqLogger.Error(err, "unable to reconcile ca configmap")
		return reconcile.Result{}, err
	}

	// Reconcile the cluster service
	svc, err := r.reconcileService(cluster)
	if err != nil {
		reqLogger.Error(err, "unable to reconcile service")
		return reconcile.Result{}, err
	}

	// Reconcile the cluster TLS secrets
	err = r.reconcileTLSSecrets(cluster, svc, caSecret)
	if err != nil {
		reqLogger.Error(err, "unable to reconcile tls secrets")
		return reconcile.Result{}, err
	}

	// Reconcile the cluster admin secret
	err = r.reconcileAdminSecret(cluster)
	if err != nil {
		reqLogger.Error(err, "unable to reconcile admin secret")
		return reconcile.Result{}, err
	}

	// Reconcile the cluster server pods
	err = r.reconcileServerPods(cluster)
	if err != nil {
		reqLogger.Error(err, "unable to reconcile server pods")
		return reconcile.Result{}, err
	}

	// No errors, return and don't requeue
	return reconcile.Result{}, nil
}

func (r *ReconcileRethinkDBCluster) reconcileAdminSecret(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
	name := fmt.Sprintf("%s-admin", cr.ObjectMeta.Name)
	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("creating new secret", "secret", name)
		secret, err := newUserSecret(cr, "admin")
		if err != nil {
			return err
		}

		// Set RethinkDBCluster instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, secret, r.scheme); err != nil {
			return err
		}

		// Create the Secret and return
		return r.client.Create(context.TODO(), secret)
	} else if err != nil {
		return err
	}
	log.Info("secret exists", "secret", found.Name)
	return nil
}

// reconcileCAConfigMap ensures the cluster CA certificate ConfigMap is present, based on the given CA Secret.
func (r *ReconcileRethinkDBCluster) reconcileCAConfigMap(cr *rethinkdbv1alpha1.RethinkDBCluster, caSecret *corev1.Secret) error {
	name := fmt.Sprintf("%s-ca", cr.Name)
	found := &corev1.ConfigMap{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("creating new configmap", "configmap", name)
		cm, err := newCAConfigMap(cr, caSecret)
		if err != nil {
			return err
		}

		// Set RethinkDBCluster instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, cm, r.scheme); err != nil {
			return err
		}

		return r.client.Create(context.TODO(), cm)
	} else if err != nil {
		return err
	}

	log.Info("configmap exists", "configmap", found.Name)
	return nil
}

// reconcileCASecret ensures the CA TLS Secret is created.
// The secret is returned upon success to be used by other reconcilers.
func (r *ReconcileRethinkDBCluster) reconcileCASecret(cr *rethinkdbv1alpha1.RethinkDBCluster) (*corev1.Secret, error) {
	name := fmt.Sprintf("%s-ca", cr.Name)
	found := &corev1.Secret{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("creating new ca secret", "secret", name)

		secret, err := newCASecret(cr, name)
		if err != nil {
			return nil, err
		}

		// Set RethinkDB instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, secret, r.scheme); err != nil {
			return nil, err
		}

		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return nil, err
		}

		return secret, nil
	} else if err != nil {
		return nil, err
	}

	log.Info("ca secret exists", "secret", found.Name)
	return found, nil
}

// reconcileCertificates ensures the TLS secrets are created for the given RethinkDBCluster.
func (r *ReconcileRethinkDBCluster) reconcileTLSSecrets(cr *rethinkdbv1alpha1.RethinkDBCluster, svc *corev1.Service, caSecret *corev1.Secret) error {
	// Reconcile the cluster certificate Secret
	err := r.reconcileTLSSecretWithSuffix(cr, svc, caSecret, "cluster")
	if err != nil {
		return err
	}

	// Reconcile the driver certificate Secret
	err = r.reconcileTLSSecretWithSuffix(cr, svc, caSecret, "driver")
	if err != nil {
		return err
	}

	// Reconcile the http (web-admin) certificate Secret
	err = r.reconcileTLSSecretWithSuffix(cr, svc, caSecret, "http")
	return err
}

// reconcileTLSSecretWithSuffix ensures the TLS Secret is created for the given Service with the given suffix.
func (r *ReconcileRethinkDBCluster) reconcileTLSSecretWithSuffix(cr *rethinkdbv1alpha1.RethinkDBCluster, svc *corev1.Service, caSecret *corev1.Secret, suffix string) error {
	found := &corev1.Secret{}
	name := fmt.Sprintf("%s-%s", cr.Name, suffix)

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("creating new secret", "secret", name)

		caCert, err := parsePEMEncodedCert(caSecret.Data[corev1.TLSCertKey])
		if err != nil {
			return err
		}

		caKey, err := parsePEMEncodedPrivateKey(caSecret.Data[corev1.TLSPrivateKeyKey])
		if err != nil {
			return err
		}

		secret, err := newCertificateSecret(cr, name, caCert, caKey)
		if err != nil {
			return err
		}

		// Set RethinkDB instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, secret, r.scheme); err != nil {
			return err
		}

		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			return err
		}

		return nil
	} else if err != nil {
		return err
	}

	log.Info("secret exists", "secret", found.Name)
	return nil
}

// reconcileServers ensures the requested number of server Pods are created.
func (r *ReconcileRethinkDBCluster) reconcileServerPods(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
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

	log.Info("correct cluster size reached", "size", memberCount)
	return nil
}

// reconcileService ensures the Service is created.
func (r *ReconcileRethinkDBCluster) reconcileService(cr *rethinkdbv1alpha1.RethinkDBCluster) (*corev1.Service, error) {
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("creating new service", "service", cr.Name)
		svc := newService(cr)

		// Set RethinkDBCluster instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, svc, r.scheme); err != nil {
			return nil, err
		}

		// Create the Service
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return nil, err
		}

		// Service created successfully, update status and return
		cr.Status.ServiceName = svc.Name
		return svc, r.client.Status().Update(context.TODO(), cr)
	} else if err != nil {
		return nil, err
	}

	log.Info("service exists", "service", found.Name)
	return found, nil
}

// addMember will add a new Pod to the cluster.
func (r *ReconcileRethinkDBCluster) addMember(cr *rethinkdbv1alpha1.RethinkDBCluster, members []corev1.Pod) error {
	log.Info("creating new server pod")
	pod := newPod(cr, members)

	// Set RethinkDB instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, pod, r.scheme); err != nil {
		return err
	}

	err := r.client.Create(context.TODO(), pod)
	if err != nil {
		return err
	}

	// Pod created successfully, update status and return
	cr.Status.Servers = append(cr.Status.Servers, pod.Name)
	return r.client.Status().Update(context.TODO(), cr)
}

// listMembers will return a slice containing the server Pods in the cluster.
func (r *ReconcileRethinkDBCluster) listMembers(cr *rethinkdbv1alpha1.RethinkDBCluster) ([]corev1.Pod, error) {
	found := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForCluster(cr))
	listOps := &client.ListOptions{Namespace: cr.Namespace, LabelSelector: labelSelector}
	err := r.client.List(context.TODO(), listOps, found)
	if err != nil {
		log.Error(err, "failed to list server pods")
		return nil, err
	}
	return found.Items, nil
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

	// Pod deleted successfully, update status and return
	members = append(members[:0], members[1:]...)
	cr.Status.Servers = []string{}
	for _, pod := range members {
		cr.Status.Servers = append(cr.Status.Servers, pod.Name)
	}
	return r.client.Status().Update(context.TODO(), cr)
}
