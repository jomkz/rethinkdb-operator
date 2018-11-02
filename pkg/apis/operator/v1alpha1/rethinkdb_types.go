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

package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultBaseImage = "jmckind/rethinkdb"
	defaultVersion   = "latest"
)

// PodPolicy defines the policy for pods owned by rethinkdb operator.
type PodPolicy struct {
	// Resources is the resource requirements for the containers.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// PersistentVolumeClaimSpec is the spec to describe PVC for the rethinkdb container
	// This field is optional. If no PVC spec, rethinkdb container will use emptyDir as volume
	PersistentVolumeClaimSpec *v1.PersistentVolumeClaimSpec `json:"persistentVolumeClaimSpec,omitempty"`
}

// RethinkDBSpec defines the desired state of RethinkDB
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
type RethinkDBSpec struct {
	// Size is the number of Pods to create for the RethinkDB cluster. Default: 1
	Size int32 `json:"size"`

	// BaseImage is the RethinkDB container image to use for the Pods.
	BaseImage string `json:"baseImage"`

	// Version is the RethinkDB container image version to use for the Pods.
	Version string `json:"version"`

	// WebAdminEnabled indicates whether or not the Web Admin will be enabled for the cluster.
	WebAdminEnabled bool `json:"webAdminEnabled"`

	// Name of ConfigMap to use or create.
	ConfigMapName string `json:"configMapName"`

	// Name of Secret to use or create.
	SecretName string `json:"secretName"`

	// Pod defines the policy for pods owned by rethinkdb operator.
	// This field cannot be updated once the CR is created.
	Pod *PodPolicy `json:"pod,omitempty"`
}

// RethinkDBStatus defines the observed state of RethinkDB
// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
type RethinkDBStatus struct {
	// Pods are the names of the rethinkdb pods
	Pods []string `json:"pods"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RethinkDB is the Schema for the rethinkdbs API
// +k8s:openapi-gen=true
type RethinkDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RethinkDBSpec   `json:"spec,omitempty"`
	Status RethinkDBStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RethinkDBList contains a list of RethinkDB
type RethinkDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RethinkDB `json:"items"`
}

// SetDefaults sets the default vaules for the cuberite spec and returns true if the spec was changed
func (r *RethinkDB) SetDefaults() bool {
	changed := false
	rs := &r.Spec
	if rs.Size == 0 {
		rs.Size = 1
		changed = true
	}
	if len(rs.BaseImage) == 0 {
		rs.BaseImage = defaultBaseImage
		changed = true
	}
	if len(rs.Version) == 0 {
		rs.Version = defaultVersion
		changed = true
	}
	if len(rs.ConfigMapName) == 0 {
		rs.ConfigMapName = r.Name
		changed = true
	}
	if len(rs.SecretName) == 0 {
		rs.SecretName = r.Name
		changed = true
	}
	return changed
}

// IsPVEnabled helper to determine if persistent volumes have been enabled
func (r *RethinkDB) IsPVEnabled() bool {
	if podPolicy := r.Spec.Pod; podPolicy != nil {
		return podPolicy.PersistentVolumeClaimSpec != nil
	}
	return false
}

func init() {
	SchemeBuilder.Register(&RethinkDB{}, &RethinkDBList{})
}
