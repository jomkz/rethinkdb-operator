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

package stub

import (
	"fmt"
	"reflect"

	"github.com/coreos/operator-sdk/pkg/sdk/action"
	"github.com/coreos/operator-sdk/pkg/sdk/query"
	v1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"github.com/sirupsen/logrus"
	apps_v1beta2 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
)

func createOrUpdateStatefulSet(r *v1alpha1.RethinkDB) error {
	name := r.Name
	replicas := r.Spec.Size
	cluster := name + "-cluster"
	terminationSeconds := int64(5)

	ss := &apps_v1beta2.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1beta2",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.ObjectMeta.Namespace,
			Labels:    labelsForRethinkDB(r),
		},
	}

	err := query.Get(ss)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get statefulset: %v", err)
	}

	if apierrors.IsNotFound(err) {
		logrus.Infof("creating statefulset: %v", name)
		ss.Spec = apps_v1beta2.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForRethinkDB(r),
			},
			ServiceName: cluster,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsForRethinkDB(r),
				},
				Spec: v1.PodSpec{
					Containers:                    addContainers(r),
					InitContainers:                addInitContainers(r),
					TerminationGracePeriodSeconds: &terminationSeconds,
					Volumes: addVolumes(r),
				},
			},
			VolumeClaimTemplates: addPVCs(r),
		}

		addOwnerRefToObject(ss, asOwner(r))
		err = action.Create(ss)
		if err != nil {
			return fmt.Errorf("failed to create statefulset: %v", err)
		}

		err = query.Get(ss)
		if err != nil {
			return fmt.Errorf("failed to get statefulset: %v", err)
		}
	}

	if *ss.Spec.Replicas != replicas {
		logrus.Infof("updating statefulset: %v", name)
		ss.Spec.Replicas = &replicas
		err = action.Update(ss)
		if err != nil {
			return fmt.Errorf("failed to update statefulset: %v", err)
		}
	}

	return nil
}

func createOrUpdateClusterService(r *v1alpha1.RethinkDB) error {
	name := r.Name + "-cluster"
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.ObjectMeta.Namespace,
			Labels:    labelsForRethinkDB(r),
		},
	}

	err := query.Get(svc)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get cluster service: %v", err)
	}

	if apierrors.IsNotFound(err) {
		logrus.Infof("creating cluster service: %v", name)
		svc.Spec = v1.ServiceSpec{
			ClusterIP: "None",
			Selector:  labelsForRethinkDB(r),
			Ports: []v1.ServicePort{{
				Port: 29015,
				Name: "cluster",
			}},
		}

		addOwnerRefToObject(svc, asOwner(r))
		err = action.Create(svc)
		if err != nil {
			return fmt.Errorf("failed to create cluster service: %v", err)
		}
	}

	return nil
}

func createOrUpdateDriverService(r *v1alpha1.RethinkDB) error {
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.ObjectMeta.Namespace,
			Labels:    labelsForRethinkDB(r),
		},
	}

	err := query.Get(svc)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get driver service: %v", err)
	}

	if apierrors.IsNotFound(err) {
		logrus.Infof("creating driver service: %v", r.Name)
		svc.Spec = v1.ServiceSpec{
			Selector:        labelsForRethinkDB(r),
			SessionAffinity: "ClientIP",
			Type:            "NodePort",
			Ports: []v1.ServicePort{{
				Port: 8080,
				Name: "http",
			},
				{
					Port: 28015,
					Name: "driver",
				}},
		}

		addOwnerRefToObject(svc, asOwner(r))
		err = action.Create(svc)
		if err != nil {
			return fmt.Errorf("failed to create driver service: %v", err)
		}
	}

	return nil
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given rethinkdb CR name.
func labelsForRethinkDB(r *v1alpha1.RethinkDB) map[string]string {
	return map[string]string{
		"app":     "rethinkdb",
		"cluster": r.Name,
	}
}

func updateStatus(r *v1alpha1.RethinkDB) error {
	var podNames []string
	podList := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}

	labelSelector := apilabels.SelectorFromSet(r.ObjectMeta.Labels).String()
	listOps := &metav1.ListOptions{LabelSelector: labelSelector}

	err := query.List(r.ObjectMeta.Namespace, podList, query.WithListOptions(listOps))
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}

	if !reflect.DeepEqual(podNames, r.Status.Pods) {
		r.Status.Pods = podNames
		err := action.Update(r)
		if err != nil {
			return fmt.Errorf("failed to update rethinkdb status: %v", err)
		}
	}

	return nil
}
