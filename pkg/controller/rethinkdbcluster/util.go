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
	"strings"

	"github.com/jmckind/rethinkdb-operator/pkg/apis/rethinkdb/v1alpha1"
)

// defaultLabels returns the default set of labels for the cluster.
func defaultLabels(cr *v1alpha1.RethinkDBCluster) map[string]string {
	return map[string]string{
		"app":     "rethinkdb",
		"cluster": cr.Name,
	}
}

// labelsForCluster returns the labels for all cluster resources.
func labelsForCluster(cr *v1alpha1.RethinkDBCluster) map[string]string {
	labels := defaultLabels(cr)
	for key, val := range cr.ObjectMeta.Labels {
		labels[key] = val
	}
	return labels
}

// setDefaults sets the default vaules for the spec and returns true if the spec was changed.
func setDefaults(cr *v1alpha1.RethinkDBCluster) bool {
	changed := false
	rs := &cr.Spec
	if rs.Size <= 0 {
		rs.Size = 1
		changed = true
	}
	if strings.TrimSpace(rs.Version) == "" {
		rs.Version = defaultVersion
		changed = true
	}
	return changed
}
