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

package stub

import (
	"fmt"

	v1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func addContainers(r *v1alpha1.RethinkDB) []v1.Container {
	return []v1.Container{{
		Command: []string{
			"/usr/bin/rethinkdb",
			"--no-update-check",
			"--config-file",
			"/etc/rethinkdb/rethinkdb.conf",
		},
		Image: fmt.Sprintf("%s:%s", r.Spec.BaseImage, r.Spec.Version),
		Name:  "rethinkdb",
		Ports: []v1.ContainerPort{{
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
		Resources: addContainerResources(r),
		Stdin:     true,
		TTY:       true,
		VolumeMounts: []v1.VolumeMount{{
			Name:      "rethinkdb-data",
			MountPath: "/var/lib/rethinkdb/default",
		},
			{
				Name:      "rethinkdb-etc",
				MountPath: "/etc/rethinkdb",
			}},
	}}
}

func addContainerResources(r *v1alpha1.RethinkDB) v1.ResourceRequirements {
	resources := v1.ResourceRequirements{}
	if r.Spec.Pod != nil {
		resources = r.Spec.Pod.Resources
	}
	return resources
}

func addEmptyDirVolume(name string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
}

func addInitContainers(r *v1alpha1.RethinkDB) []v1.Container {
	name := r.Name
	cluster := name + "-cluster"

	return []v1.Container{{
		Command: []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf("echo '%s' > /etc/rethinkdb/rethinkdb.conf; if nslookup %s; then echo join=%s-0.%s:29015 >> /etc/rethinkdb/rethinkdb.conf; fi;", defaultConfig, cluster, name, cluster),
		},
		Image: "busybox:latest",
		Name:  "cluster-init",
		VolumeMounts: []v1.VolumeMount{{
			Name:      "rethinkdb-etc",
			MountPath: "/etc/rethinkdb",
		}},
	}}
}

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

func addPVCs(r *v1alpha1.RethinkDB) []v1.PersistentVolumeClaim {
	var pvcs []v1.PersistentVolumeClaim

	if r.IsPVEnabled() {
		pvcs = append(pvcs, v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rethinkdb-data",
				Namespace: r.ObjectMeta.Namespace,
				Labels:    r.ObjectMeta.Labels,
			},
			Spec: *r.Spec.Pod.PersistentVolumeClaimSpec,
		})
	}

	return pvcs
}

func addVolumes(r *v1alpha1.RethinkDB) []v1.Volume {
	var volumes []v1.Volume

	volumes = append(volumes, addEmptyDirVolume("rethinkdb-etc"))

	if !r.IsPVEnabled() {
		volumes = append(volumes, addEmptyDirVolume("rethinkdb-data"))
	}

	return volumes
}

// asOwner returns an OwnerReference set as the rethinkdb CR
func asOwner(r *v1alpha1.RethinkDB) metav1.OwnerReference {
	controller := true
	return metav1.OwnerReference{
		APIVersion: r.APIVersion,
		Kind:       r.Kind,
		Name:       r.Name,
		UID:        r.UID,
		Controller: &controller,
	}
}
