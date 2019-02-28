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
	"github.com/jmckind/rethinkdb-operator/pkg/apis/rethinkdb/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newServicePorts constructs the ServicePort objects for the Service.
func newServicePorts(cr *v1alpha1.RethinkDBCluster) []corev1.ServicePort {
	var ports []corev1.ServicePort

	ports = append(ports, corev1.ServicePort{Port: 28015, Name: "driver"})
	if cr.Spec.WebAdminEnabled {
		ports = append(ports, corev1.ServicePort{Port: 8080, Name: "http"})
	}

	return ports
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
			Name:      cr.Name,
			Namespace: cr.ObjectMeta.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector:        labels,
			SessionAffinity: "ClientIP",
			Ports:           newServicePorts(cr),
		},
	}
}
