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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IMPORTANT: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// RethinkDBPodPolicy defines the policy for pods owned by rethinkdb operator.
// +k8s:openapi-gen=true
type RethinkDBPodPolicy struct {
	// Resources is the resource requirements for the containers.
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// PersistentVolumeClaimSpec is the spec to describe PVC for the rethinkdb container
	// This field is optional. If no PVC spec, rethinkdb container will use emptyDir as volume
	PersistentVolumeClaimSpec *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaimSpec,omitempty"`
}

// RethinkDBClusterSpec defines the desired state of RethinkDBCluster
// +k8s:openapi-gen=true
type RethinkDBClusterSpec struct {
	// Size is the number of Pods to create for the RethinkDB cluster. Default: 1
	Size int32 `json:"size"`

	// Version is the RethinkDB version to use for the cluster.
	Version string `json:"version,omitempty"`

	// WebAdminEnabled indicates whether or not the Web Admin will be enabled for the cluster.
	WebAdminEnabled bool `json:"webAdminEnabled,omitempty"`

	// Pod defines the policy for pods owned by rethinkdb operator.
	// This field cannot be updated once the CR is created.
	Pod *RethinkDBPodPolicy `json:"pod,omitempty"`
}

// RethinkDBClusterStatus defines the observed state of RethinkDBCluster
// +k8s:openapi-gen=true
type RethinkDBClusterStatus struct {
	// Servers is a list of the names of the rethinkdb server Pods in the cluster.
	Servers []string `json:"servers,omitempty"`

	// ServiceName is the name of the Service for accessing the RethinkDB cluster.
	ServiceName string `json:"serviceName,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RethinkDBCluster is the Schema for the rethinkdbclusters API
// +k8s:openapi-gen=true
type RethinkDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RethinkDBClusterSpec   `json:"spec,omitempty"`
	Status RethinkDBClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RethinkDBClusterList contains a list of RethinkDBCluster
type RethinkDBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RethinkDBCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RethinkDBCluster{}, &RethinkDBClusterList{})
}
