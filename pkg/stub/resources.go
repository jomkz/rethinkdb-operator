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

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *RethinkDBCluster) AddContainers() []v1.Container {
	return []v1.Container{{
		Command: []string{
			"/usr/bin/rethinkdb",
			"--no-update-check",
			"--config-file",
			"/etc/rethinkdb/rethinkdb.conf",
		},
		Image: fmt.Sprintf("%s:%s", c.Resource.Spec.BaseImage, c.Resource.Spec.Version),
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
		Resources: c.AddContainerResources(),
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

func (c *RethinkDBCluster) AddContainerResources() v1.ResourceRequirements {
	resources := v1.ResourceRequirements{}
	if c.Resource.Spec.Pod != nil {
		resources = c.Resource.Spec.Pod.Resources
	}
	return resources
}

func (c *RethinkDBCluster) AddEmptyDirVolume(name string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
}

func (c *RethinkDBCluster) AddInitContainers() []v1.Container {
	name := c.Resource.Name
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
func (c *RethinkDBCluster) AddOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

func (c *RethinkDBCluster) AddPVCs() []v1.PersistentVolumeClaim {
	var pvcs []v1.PersistentVolumeClaim

	if c.Resource.IsPVEnabled() {
		pvcs = append(pvcs, v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rethinkdb-data",
				Namespace: c.Resource.ObjectMeta.Namespace,
				Labels:    c.Resource.ObjectMeta.Labels,
			},
			Spec: *c.Resource.Spec.Pod.PersistentVolumeClaimSpec,
		})
	}

	return pvcs
}

func (c *RethinkDBCluster) AddVolumes() []v1.Volume {
	var volumes []v1.Volume

	volumes = append(volumes, c.AddEmptyDirVolume("rethinkdb-etc"))

	if !c.Resource.IsPVEnabled() {
		volumes = append(volumes, c.AddEmptyDirVolume("rethinkdb-data"))
	}

	return volumes
}

// asOwner returns an OwnerReference set as the rethinkdb CR
func (c *RethinkDBCluster) AsOwner() metav1.OwnerReference {
	controller := true
	return metav1.OwnerReference{
		APIVersion: c.Resource.APIVersion,
		Kind:       c.Resource.Kind,
		Name:       c.Resource.Name,
		UID:        c.Resource.UID,
		Controller: &controller,
	}
}
