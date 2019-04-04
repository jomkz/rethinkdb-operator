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

// newCAConfigMap creates a new ConfigMap for the given RethinkDBCluster and CA certificate.
func newCAConfigMap(cr *v1alpha1.RethinkDBCluster, caSecret *corev1.Secret) (*corev1.ConfigMap, error) {
	cm := newConfigMapWithSuffix(cr, "ca")
	cm.Data = map[string]string{
		TLSCACertKey: string(caSecret.Data[TLSCertKey]),
	}
	return cm, nil
}

// newConfigMap creates a new ConfigMap for the given RethinkDBCluster.
func newConfigMap(cr *v1alpha1.RethinkDBCluster) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labelsForCluster(cr),
		},
	}
}

// newConfigMapWithName creates a new ConfigMap with the given name for the given RethinkDBCluster.
func newConfigMapWithName(cr *v1alpha1.RethinkDBCluster, name string) *corev1.ConfigMap {
	cm := newConfigMap(cr)
	cm.ObjectMeta.Name = name
	return cm
}

// newConfigMapWithName creates a new ConfigMap with the given suffix appended to the name.
// The name for the CongifMap is based on the name of the given RethinkDBCluster.
func newConfigMapWithSuffix(cr *v1alpha1.RethinkDBCluster, suffix string) *corev1.ConfigMap {
	return newConfigMapWithName(cr, fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, suffix))
}
