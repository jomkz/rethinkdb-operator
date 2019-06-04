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

	"github.com/coreos/go-semver/semver"
	rethinkdbv1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/rethinkdb/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

	// Watch for changes to secondary resource PersistentVolumeClaims and requeue the owner RethinkDBCluster
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, &handler.EnqueueRequestForOwner{
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
		if k8serrors.IsNotFound(err) {
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

	// Reconcile the admin service
	err = r.reconcileAdminService(cluster)
	if err != nil {
		reqLogger.Error(err, "unable to reconcile admin service")
		return reconcile.Result{}, err
	}

	// Reconcile the driver service
	err = r.reconcileDriverService(cluster)
	if err != nil {
		reqLogger.Error(err, "unable to reconcile driver service")
		return reconcile.Result{}, err
	}

	// Reconcile the cluster TLS secrets
	err = r.reconcileTLSSecrets(cluster, caSecret)
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

	// Reconcile the cluster persistent volume claims
	// err = r.reconcilePersistentVolumeClaims(cluster)
	// if err != nil {
	// 	reqLogger.Error(err, "unable to reconcile persistent volume claims")
	// 	return reconcile.Result{}, err
	// }

	// Reconcile the cluster server pods
	err = r.reconcileServerPods(cluster)
	if err != nil {
		reqLogger.Error(err, "unable to reconcile server pods")
		return reconcile.Result{}, err
	}

	// No errors, return and don't requeue
	return reconcile.Result{}, nil
}

// addPVC will add a new Pod to the cluster.
func (r *ReconcileRethinkDBCluster) addPVC(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
	log.Info("creating new persistent volume claim")
	pvc := newPVC(cr)

	// Set RethinkDB instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, pvc, r.scheme); err != nil {
		return err
	}

	return r.client.Create(context.TODO(), newPVC(cr))
}

// addServer will add a new Pod to the cluster.
func (r *ReconcileRethinkDBCluster) addServer(cr *rethinkdbv1alpha1.RethinkDBCluster, members []corev1.Pod) error {
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

// listPVCs will return a slice containing the persistent volume claims for the cluster.
func (r *ReconcileRethinkDBCluster) listPVCs(cr *rethinkdbv1alpha1.RethinkDBCluster) ([]corev1.PersistentVolumeClaim, error) {
	found := &corev1.PersistentVolumeClaimList{}
	labelSelector := labels.SelectorFromSet(labelsForCluster(cr))
	listOps := &client.ListOptions{Namespace: cr.Namespace, LabelSelector: labelSelector}
	err := r.client.List(context.TODO(), listOps, found)
	if err != nil {
		log.Error(err, "failed to list persistent volume claims")
		return nil, err
	}
	return found.Items, nil
}

// listServers will return a slice containing the server Pods in the cluster.
func (r *ReconcileRethinkDBCluster) listServers(cr *rethinkdbv1alpha1.RethinkDBCluster) ([]corev1.Pod, error) {
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

// reconcileAdminSecret ensures the cluster admin user credentials are present.
func (r *ReconcileRethinkDBCluster) reconcileAdminSecret(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
	name := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, RethinkDBAdminKey)
	found := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, found)
	if err != nil && k8serrors.IsNotFound(err) {
		log.Info("creating new secret", "secret", name)
		secret, err := newUserSecret(cr, RethinkDBAdminKey)
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

// reconcileAdminService ensures the admin Service is created.
func (r *ReconcileRethinkDBCluster) reconcileAdminService(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
	found := &corev1.Service{}
	name := fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, RethinkDBAdminKey)

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, found)
	if err != nil && k8serrors.IsNotFound(err) {
		if !cr.Spec.WebAdminEnabled {
			// Admin not enabled, do not create service
			return nil
		}

		log.Info("creating new service", "service", name)
		svc := newAdminService(cr)

		// Set RethinkDBCluster instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, svc, r.scheme); err != nil {
			return err
		}

		// Create the Service and return
		return r.client.Create(context.TODO(), svc)
	} else if err != nil {
		return err
	}

	// Service exists, verify that it should...
	if !cr.Spec.WebAdminEnabled {
		log.Info("removing existing service", "service", name)
		return r.client.Delete(context.TODO(), found)
	}

	log.Info("service exists", "service", found.Name)
	return nil
}

// reconcileCAConfigMap ensures the cluster CA certificate ConfigMap is present, based on the given CA Secret.
func (r *ReconcileRethinkDBCluster) reconcileCAConfigMap(cr *rethinkdbv1alpha1.RethinkDBCluster, caSecret *corev1.Secret) error {
	name := fmt.Sprintf("%s-ca", cr.Name)
	found := &corev1.ConfigMap{}

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, found)
	if err != nil && k8serrors.IsNotFound(err) {
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
	if err != nil && k8serrors.IsNotFound(err) {
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

// reconcileDriverService ensures the driver Service is present.
func (r *ReconcileRethinkDBCluster) reconcileDriverService(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, found)
	if err != nil && k8serrors.IsNotFound(err) {
		log.Info("creating new service", "service", cr.Name)
		svc := newDriverService(cr)

		// Set RethinkDBCluster instance as the owner and controller
		if err = controllerutil.SetControllerReference(cr, svc, r.scheme); err != nil {
			return err
		}

		// Create the Service
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return err
		}

		// Service created successfully, update status and return
		cr.Status.ServiceName = svc.Name
		return r.client.Status().Update(context.TODO(), cr)
	} else if err != nil {
		return err
	}

	log.Info("service exists", "service", found.Name)
	return nil
}

// reconcilePersistentVolumeClaims ensures the requested number of PVCs are created.
func (r *ReconcileRethinkDBCluster) reconcilePersistentVolumeClaims(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
	if !isPVEnabled(cr) {
		return nil
	}

	pvcs, err := r.listPVCs(cr)
	if err != nil {
		log.Error(err, "unable to list persistent volume claims")
		return err
	}
	pvcCount := int32(len(pvcs))

	if pvcCount < cr.Spec.Size {
		return r.addPVC(cr)
	} else if pvcCount > cr.Spec.Size {
		return r.removePVC(cr, pvcs)
	}

	log.Info("correct persistent volume claim count reached", "count", pvcCount)
	return nil
}

// reconcileServers ensures the requested number of server Pods are created.
func (r *ReconcileRethinkDBCluster) reconcileServerPods(cr *rethinkdbv1alpha1.RethinkDBCluster) error {
	servers, err := r.listServers(cr)
	if err != nil {
		log.Error(err, "unable to list servers")
		return err
	}
	serverCount := int32(len(servers))

	if serverCount < cr.Spec.Size {
		// Ensure all existing Pods are running before adding a new Pod.
		for _, pod := range servers {
			if pod.Status.Phase != corev1.PodRunning {
				log.Info("waiting for existing server pods to enter running phase", "pod", pod.ObjectMeta.Name)
				return nil
			}
		}
		return r.addServer(cr, servers)
	} else if serverCount > cr.Spec.Size {
		return r.removeServer(cr, servers)
	}

	log.Info("correct cluster size reached", "size", serverCount)
	return r.reconcileServerUpgrade(cr, servers)
}

// reconcileServerUpgrade will upgrade each server Pod to the correct version for the given RethinkDBCluster.
func (r *ReconcileRethinkDBCluster) reconcileServerUpgrade(cr *rethinkdbv1alpha1.RethinkDBCluster, servers []corev1.Pod) error {
	// Ensure all pods ready before upgrade
	for _, pod := range servers {
		log.Info("verify pod phase", "pod", pod.ObjectMeta.Name, "phase", pod.Status.Phase)
		if pod.Status.Phase != corev1.PodRunning {
			log.Info("waiting for existing server pods to enter running phase", "pod", pod.ObjectMeta.Name)
			return nil
		}

		for _, status := range pod.Status.ContainerStatuses {
			log.Info("verify container status", "pod", pod.ObjectMeta.Name, "ready", status.Ready)
			if !status.Ready {
				log.Info("waiting for existing server pods to become ready", "pod", pod.ObjectMeta.Name)
				return nil
			}
		}
	}

	// Upgrade the Pod by chaging the tag for the container image
	for _, pod := range servers {
		image, version := parseContainerImage(pod.Spec.Containers[0].Image)
		oldVersion := semver.New(version)
		newVersion := semver.New(cr.Spec.Version)

		if oldVersion.LessThan(*newVersion) {
			log.Info("upgrading server pod", "pod", pod.ObjectMeta.Name, "old", oldVersion, "new", newVersion)
			pod.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, newVersion)
			return r.client.Update(context.TODO(), &pod)
		}
	}
	return nil
}

// reconcileCertificates ensures the TLS secrets are created for the given RethinkDBCluster.
func (r *ReconcileRethinkDBCluster) reconcileTLSSecrets(cr *rethinkdbv1alpha1.RethinkDBCluster, caSecret *corev1.Secret) error {
	// Reconcile the cluster certificate Secret
	err := r.reconcileTLSSecretWithSuffix(cr, caSecret, RethinkDBClusterKey)
	if err != nil {
		return err
	}

	// Reconcile the driver certificate Secret
	err = r.reconcileTLSSecretWithSuffix(cr, caSecret, RethinkDBDriverKey)
	if err != nil {
		return err
	}

	// Reconcile the http (web-admin) certificate Secret
	err = r.reconcileTLSSecretWithSuffix(cr, caSecret, RethinkDBHttpKey)
	if err != nil {
		return err
	}

	// Reconcile the client certificate Secret
	err = r.reconcileTLSSecretWithSuffix(cr, caSecret, RethinkDBClientKey)
	return err
}

// reconcileTLSSecretWithSuffix ensures the TLS Secret is created for the given Service with the given suffix.
func (r *ReconcileRethinkDBCluster) reconcileTLSSecretWithSuffix(cr *rethinkdbv1alpha1.RethinkDBCluster, caSecret *corev1.Secret, suffix string) error {
	found := &corev1.Secret{}
	name := fmt.Sprintf("%s-%s", cr.Name, suffix)

	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: cr.Namespace}, found)
	if err != nil && k8serrors.IsNotFound(err) {
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

		return r.client.Create(context.TODO(), secret)
	} else if err != nil {
		return err
	}

	log.Info("secret exists", "secret", found.Name)
	return nil
}

// removePVC will delete a PVC from the cluster. The first unbound PVC in the provided slice of PVCs will be deleted.
func (r *ReconcileRethinkDBCluster) removePVC(cr *rethinkdbv1alpha1.RethinkDBCluster, pvcs []corev1.PersistentVolumeClaim) error {
	if len(pvcs) <= 0 {
		return nil
	}

	for _, pvc := range pvcs {
		if pvc.Status.Phase == corev1.ClaimPending {
			log.Info("removing existing persistent volume claim", "pvc", pvc.ObjectMeta.Name)
			err := r.client.Delete(context.TODO(), &pvc)
			if err != nil {
				return err
			}
		}
	}

	// Nothing to remove...
	return nil
}

// removeServer will delete a server Pod from the cluster. The first Pod in the provided slice of servers will be deleted.
func (r *ReconcileRethinkDBCluster) removeServer(cr *rethinkdbv1alpha1.RethinkDBCluster, servers []corev1.Pod) error {
	if len(servers) <= 0 {
		return nil
	}

	pod := servers[0]
	log.Info("removing existing server pod", "pod", pod.Name)

	err := r.client.Delete(context.TODO(), &pod)
	if err != nil {
		return err
	}

	// Pod deleted successfully, update status and return
	servers = append(servers[:0], servers[1:]...)
	cr.Status.Servers = []string{}
	for _, pod := range servers {
		cr.Status.Servers = append(cr.Status.Servers, pod.Name)
	}
	return r.client.Status().Update(context.TODO(), cr)
}
