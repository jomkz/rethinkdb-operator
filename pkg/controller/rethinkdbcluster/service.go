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
	"fmt"

	"github.com/jmckind/rethinkdb-operator/pkg/apis/rethinkdb/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newAdminService constructs a new admin Service object.
func newAdminService(cr *v1alpha1.RethinkDBCluster) *corev1.Service {
	svc := newService(cr)
	svc.ObjectMeta.Name = fmt.Sprintf("%s-%s", cr.Name, RethinkDBAdminKey)
	svc.Spec.Ports = []corev1.ServicePort{
		corev1.ServicePort{Port: RethinkDBHttpPort, Name: RethinkDBHttpKey},
	}
	return svc
}

// newDriverService constructs a new driver Service object.
func newDriverService(cr *v1alpha1.RethinkDBCluster) *corev1.Service {
	svc := newService(cr)
	svc.Spec.Ports = []corev1.ServicePort{
		corev1.ServicePort{Port: RethinkDBDriverPort, Name: RethinkDBDriverKey},
	}
	return svc
}

// newService constructs a new Service object.
func newService(cr *v1alpha1.RethinkDBCluster) *corev1.Service {
	labels := labelsForCluster(cr)
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.ObjectMeta.Name,
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector:        labels,
			SessionAffinity: "ClientIP",
		},
	}
}
