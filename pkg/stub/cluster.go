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

type RethinkDBCluster struct {
	Resource *v1alpha1.RethinkDB
}

func NewRethinkDBCluster(r *v1alpha1.RethinkDB) RethinkDBCluster {
	return RethinkDBCluster{Resource: r}
}

func (c *RethinkDBCluster) CreateOrUpdateCluster() error {
	err := c.CreateOrUpdateClusterService()
	if err != nil {
		return fmt.Errorf("failed to create or update cluster service: %v", err)
	}

	err = c.CreateOrUpdateDriverService()
	if err != nil {
		return fmt.Errorf("failed to create or update driver service: %v", err)
	}

	err = c.CreateOrUpdateStatefulSet()
	if err != nil {
		return fmt.Errorf("failed to create or update statefulset: %v", err)
	}

	err = c.UpdateStatus()
	if err != nil {
		return fmt.Errorf("failed to update rethinkdb status: %v", err)
	}

	return nil
}

func (c *RethinkDBCluster) CreateOrUpdateStatefulSet() error {
	name := c.Resource.Name
	replicas := c.Resource.Spec.Size
	cluster := name + "-cluster"
	terminationSeconds := int64(5)

	ss := &apps_v1beta2.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1beta2",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Resource.ObjectMeta.Namespace,
			Labels:    c.LabelsForCluster(),
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
				MatchLabels: c.LabelsForCluster(),
			},
			ServiceName: cluster,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: c.LabelsForCluster(),
				},
				Spec: v1.PodSpec{
					Containers:                    c.AddContainers(),
					InitContainers:                c.AddInitContainers(),
					TerminationGracePeriodSeconds: &terminationSeconds,
					Volumes: c.AddVolumes(),
				},
			},
			VolumeClaimTemplates: c.AddPVCs(),
		}

		c.AddOwnerRefToObject(ss, c.AsOwner())
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

func (c *RethinkDBCluster) CreateOrUpdateClusterService() error {
	name := c.Resource.Name + "-cluster"
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.Resource.ObjectMeta.Namespace,
			Labels:    c.LabelsForCluster(),
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
			Selector:  c.LabelsForCluster(),
			Ports: []v1.ServicePort{{
				Port: 29015,
				Name: "cluster",
			}},
		}

		c.AddOwnerRefToObject(svc, c.AsOwner())
		err = action.Create(svc)
		if err != nil {
			return fmt.Errorf("failed to create cluster service: %v", err)
		}
	}

	return nil
}

func (c *RethinkDBCluster) CreateOrUpdateDriverService() error {
	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Resource.Name,
			Namespace: c.Resource.ObjectMeta.Namespace,
			Labels:    c.LabelsForCluster(),
		},
	}

	err := query.Get(svc)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to get driver service: %v", err)
	}

	if apierrors.IsNotFound(err) {
		logrus.Infof("creating driver service: %v", c.Resource.Name)
		svc.Spec = v1.ServiceSpec{
			Selector:        c.LabelsForCluster(),
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

		c.AddOwnerRefToObject(svc, c.AsOwner())
		err = action.Create(svc)
		if err != nil {
			return fmt.Errorf("failed to create driver service: %v", err)
		}
	}

	return nil
}

func (c *RethinkDBCluster) LabelsForCluster() map[string]string {
	return map[string]string{
		"app":     "rethinkdb",
		"cluster": c.Resource.Name,
	}
}

func (c *RethinkDBCluster) UpdateStatus() error {
	var podNames []string
	podList := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
	}

	labelSelector := apilabels.SelectorFromSet(c.Resource.ObjectMeta.Labels).String()
	listOps := &metav1.ListOptions{LabelSelector: labelSelector}

	err := query.List(c.Resource.ObjectMeta.Namespace, podList, query.WithListOptions(listOps))
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}

	if !reflect.DeepEqual(podNames, c.Resource.Status.Pods) {
		c.Resource.Status.Pods = podNames
		err := action.Update(c.Resource)
		if err != nil {
			return fmt.Errorf("failed to update rethinkdb status: %v", err)
		}
	}

	return nil
}
