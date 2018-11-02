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
	"fmt"
	"log"

	operatorv1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultConfig = `bind=all
directory=/var/lib/rethinkdb/default
# driver-tls-cert=/etc/rethinkdb/driver-tls-cert.pem
# driver-tls-key=/etc/rethinkdb/driver-tls-key.pem
# http-tls-key=/etc/rethinkdb/http-tls-key.pem
# http-tls-cert=/etc/rethinkdb/http-tls-cert.pem
`
)

// generateConfiguration will return the text for the RethinkDB configuration file.
func generateConfiguration(r *operatorv1alpha1.RethinkDB) string {
	config := defaultConfig

	if !r.Spec.WebAdminEnabled {
		config += "no-http-admin\n"
	}

	return config
}

// newContainers will create the Containers for the RethinkDB Pod.
func newContainers(cr *operatorv1alpha1.RethinkDB) []corev1.Container {
	return []corev1.Container{{
		Command: []string{
			"/usr/bin/rethinkdb",
			"--no-update-check",
			"--config-file",
			"/etc/rethinkdb/rethinkdb.conf",
		},
		Image: fmt.Sprintf("%s:%s", cr.Spec.BaseImage, cr.Spec.Version),
		Name:  "rethinkdb",
		Ports: []corev1.ContainerPort{{
			ContainerPort: 8080,
			Name:          "http",
		},
			{
				ContainerPort: 28015,
				Name:          "driver",
			},
			{
				ContainerPort: 29015,
				Name:          "cluster",
			}},
		Resources: newContainerResources(cr),
		Stdin:     true,
		TTY:       true,
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "rethinkdb-data",
			MountPath: "/var/lib/rethinkdb/default",
		},
			{
				Name:      "rethinkdb-etc",
				MountPath: "/etc/rethinkdb",
			}},
	}}
}

// newContainerResources will create the container Resources for the RethinkDB Pod.
func newContainerResources(cr *operatorv1alpha1.RethinkDB) corev1.ResourceRequirements {
	resources := corev1.ResourceRequirements{}
	if cr.Spec.Pod != nil {
		resources = cr.Spec.Pod.Resources
	}
	return resources
}

// newInitContainers will create the Init Containers for the RethinkDB Pod.
func newInitContainers(cr *operatorv1alpha1.RethinkDB, members []corev1.Pod) []corev1.Container {
	config := generateConfiguration(cr)

	// Add other existing Pod IPs to join cluster
	for _, pod := range members {
		log.Printf("POD: %v", pod)
		config = fmt.Sprintf("%s\njoin=%s:29015", config, pod.Status.PodIP)
	}

	return []corev1.Container{{
		Command: []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf("echo '%s' > /etc/rethinkdb/rethinkdb.conf", config),
		},
		Image: "busybox:latest",
		Name:  "cluster-init",
		VolumeMounts: []corev1.VolumeMount{{
			Name:      "rethinkdb-etc",
			MountPath: "/etc/rethinkdb",
		}},
	}}
}

// newPod returns a new Pod with the same namespace and name prefix as the cr
func newPod(cr *operatorv1alpha1.RethinkDB, members []corev1.Pod) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", cr.Name),
			Namespace:    cr.Namespace,
			Labels:       labelsForCluster(cr),
		},
		Spec: corev1.PodSpec{
			Containers:     newContainers(cr),
			InitContainers: newInitContainers(cr, members),
			Volumes:        newVolumes(cr),
		},
	}
}
