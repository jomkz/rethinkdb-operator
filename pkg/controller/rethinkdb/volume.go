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
	operatorv1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newEmptyDirVolume creates a new EmptyDir volume with the given name.
func newEmptyDirVolume(name string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

// newPVCs creates the PVCs used by the application.
func newPVCs(cr *operatorv1alpha1.RethinkDB) []corev1.PersistentVolumeClaim {
	var pvcs []corev1.PersistentVolumeClaim

	if cr.IsPVEnabled() {
		pvcs = append(pvcs, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rethinkdb-data",
				Namespace: cr.ObjectMeta.Namespace,
				Labels:    cr.ObjectMeta.Labels,
			},
			Spec: *cr.Spec.Pod.PersistentVolumeClaimSpec,
		})
	}

	return pvcs
}

// newVolumes creates the volumes used by the application.
func newVolumes(cr *operatorv1alpha1.RethinkDB) []corev1.Volume {
	var volumes []corev1.Volume

	volumes = append(volumes, newEmptyDirVolume("rethinkdb-etc"))

	if !cr.IsPVEnabled() {
		volumes = append(volumes, newEmptyDirVolume("rethinkdb-data"))
	}

	return volumes
}
