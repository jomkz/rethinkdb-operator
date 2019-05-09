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

// isPVEnabled helper to determine if persistent volumes have been enabled.
func isPVEnabled(cr *v1alpha1.RethinkDBCluster) bool {
	if podPolicy := cr.Spec.Pod; podPolicy != nil {
		return podPolicy.PersistentVolumeClaimSpec != nil
	}
	return false
}

// newEmptyDirVolume creates a new EmptyDir volume with the given name.
func newEmptyDirVolume(name string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}
}

// newProjectedVolume creates a new Projected volume with the given name.
func newProjectedVolume(cr *v1alpha1.RethinkDBCluster, name string) corev1.Volume {
	return corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{
					corev1.VolumeProjection{
						Secret: &corev1.SecretProjection{
							LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, RethinkDBCAKey)},
							Items: []corev1.KeyToPath{
								corev1.KeyToPath{
									Key:  corev1.TLSCertKey,
									Path: fmt.Sprintf("%s.crt", RethinkDBCAKey),
								},
							},
						},
					},
					newTLSVolumeProjection(cr, RethinkDBClusterKey),
					newTLSVolumeProjection(cr, RethinkDBDriverKey),
					newTLSVolumeProjection(cr, RethinkDBHttpKey),
				},
			},
		},
	}
}

// newTLSVolumeProjection will retun a TLS certificate and key volume projection for the given name.
func newTLSVolumeProjection(cr *v1alpha1.RethinkDBCluster, name string) corev1.VolumeProjection {
	return corev1.VolumeProjection{
		Secret: &corev1.SecretProjection{
			LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, name)},
			Items: []corev1.KeyToPath{
				corev1.KeyToPath{
					Key:  corev1.TLSCertKey,
					Path: fmt.Sprintf("%s.crt", name),
				},
				corev1.KeyToPath{
					Key:  corev1.TLSPrivateKeyKey,
					Path: fmt.Sprintf("%s.key", name),
				},
			},
		},
	}
}

// newPVCs creates the PVCs used by the application.
func newPVCs(cr *v1alpha1.RethinkDBCluster) []corev1.PersistentVolumeClaim {
	var pvcs []corev1.PersistentVolumeClaim

	if isPVEnabled(cr) {
		pvcs = append(pvcs, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      RethinkDBDataKey,
				Namespace: cr.ObjectMeta.Namespace,
				Labels:    cr.ObjectMeta.Labels,
			},
			Spec: *cr.Spec.Pod.PersistentVolumeClaimSpec,
		})
	}

	return pvcs
}

// newVolumes creates the volumes used by the application.
func newVolumes(cr *v1alpha1.RethinkDBCluster) []corev1.Volume {
	volumes := []corev1.Volume{
		newProjectedVolume(cr, RethinkDBTLSSecretsKey),
	}

	if !isPVEnabled(cr) {
		volumes = append(volumes, newEmptyDirVolume(RethinkDBDataKey))
	}

	return volumes
}
