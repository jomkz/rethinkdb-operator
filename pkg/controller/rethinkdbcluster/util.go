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

const (
	// RethinkDBAdminKey is the key for the RethinkDB admin assets.
	RethinkDBAdminKey = "admin"

	// RethinkDBApp is the default RethinkDB application name.
	RethinkDBApp = "rethinkdb"

	// RethinkDBCAKey is the key for the RethinkDB CA TLS assets.
	RethinkDBCAKey = "ca"

	// RethinkDBClientKey is the key for the RethinkDB client TLS assets.
	RethinkDBClientKey = "client"

	// RethinkDBClusterKey is the key for the RethinkDB cluster TLS assets.
	RethinkDBClusterKey = "cluster"

	// RethinkDBClusterPort is the default RethinkDB cluster port.
	RethinkDBClusterPort = 29015

	// RethinkDBDataKey is the key for the RethinkDB data volume.
	RethinkDBDataKey = "rethinkdb-data"

	// RethinkDBDataPath is the default path for RethinkDB data.
	RethinkDBDataPath = "/data"

	// RethinkDBDriverKey is the key for the RethinkDB driver TLS assets.
	RethinkDBDriverKey = "driver"

	// RethinkDBDriverPort is the default RethinkDB driver port.
	RethinkDBDriverPort = 28015

	// RethinkDBHttpKey is the key for the RethinkDB http TLS assets.
	RethinkDBHttpKey = "http"

	// RethinkDBHttpPort is the default RethinkDB http (web-admin) port.
	RethinkDBHttpPort = 8080

	// RethinkDBImage is the default RethinkDB container image to run.
	RethinkDBImage = "rethinkdb"

	// RethinkDBImageTag is the default RethinkDB container image tag to run.
	RethinkDBImageTag = "latest"

	// RethinkDBTLSPath is the default path for RethinkDB TLS assets.
	RethinkDBTLSPath = "/etc/rethinkdb/tls"

	// RethinkDBTLSSecretsKey is the key for the RethinkDB TLS secrets volume.
	RethinkDBTLSSecretsKey = "tls-secrets"
)

// defaultLabels returns the default set of labels for the cluster.
func defaultLabels(cr *v1alpha1.RethinkDBCluster) map[string]string {
	return map[string]string{
		"app":     RethinkDBApp,
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
	spec := &cr.Spec
	if spec.Size <= 0 {
		spec.Size = 1
		changed = true
	}
	if strings.TrimSpace(spec.Version) == "" {
		spec.Version = RethinkDBImageTag
		changed = true
	}
	return changed
}
