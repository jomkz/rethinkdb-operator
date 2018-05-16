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

const (
	defaultConfig = `bind=all
directory=/var/lib/rethinkdb/default
# driver-tls-cert=/etc/rethinkdb/driver-tls-cert.pem
# driver-tls-key=/etc/rethinkdb/driver-tls-key.pem
# http-tls-key=/etc/rethinkdb/http-tls-key.pem
# http-tls-cert=/etc/rethinkdb/http-tls-cert.pem
`
)

func (c *RethinkDBCluster) AddOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

// asOwner returns an OwnerReference set as the rethinkdb CR
func (c *RethinkDBCluster) AsOwner() metav1.OwnerReference {
	controller := true
	r := c.Resource
	return metav1.OwnerReference{
		APIVersion: r.APIVersion,
		Kind:       r.Kind,
		Name:       r.Name,
		UID:        r.UID,
		Controller: &controller,
	}
}

func (c *RethinkDBCluster) ConstructConfiguration() string {
	config := defaultConfig

	if !c.Resource.Spec.WebAdminEnabled {
		config += "no-http-admin\n"
	}

	return config
}

func (c *RethinkDBCluster) ConstructContainers() []v1.Container {
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
		Resources: c.ConstructContainerResources(),
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

func (c *RethinkDBCluster) ConstructContainerResources() v1.ResourceRequirements {
	resources := v1.ResourceRequirements{}
	if c.Resource.Spec.Pod != nil {
		resources = c.Resource.Spec.Pod.Resources
	}
	return resources
}

func (c *RethinkDBCluster) ConstructDriverServicePorts() []v1.ServicePort {
	var ports []v1.ServicePort

	ports = append(ports, v1.ServicePort{Port: 28015, Name: "driver"})
	if c.Resource.Spec.WebAdminEnabled {
		ports = append(ports, v1.ServicePort{Port: 8080, Name: "http"})
	}

	return ports
}

func (c *RethinkDBCluster) ConstructEmptyDirVolume(name string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
}

func (c *RethinkDBCluster) ConstructInitContainers() []v1.Container {
	name := c.Resource.Name
	cluster := name + "-cluster"
	config := c.ConstructConfiguration()

	return []v1.Container{{
		Command: []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf("echo '%s' > /etc/rethinkdb/rethinkdb.conf; if nslookup %s; then echo join=%s-0.%s:29015 >> /etc/rethinkdb/rethinkdb.conf; fi;", config, cluster, name, cluster),
		},
		Image: "busybox:latest",
		Name:  "cluster-init",
		VolumeMounts: []v1.VolumeMount{{
			Name:      "rethinkdb-etc",
			MountPath: "/etc/rethinkdb",
		}},
	}}
}

func (c *RethinkDBCluster) ConstructPVCs() []v1.PersistentVolumeClaim {
	var pvcs []v1.PersistentVolumeClaim
	r := c.Resource

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

func (c *RethinkDBCluster) ConstructVolumes() []v1.Volume {
	var volumes []v1.Volume

	volumes = append(volumes, c.ConstructEmptyDirVolume("rethinkdb-etc"))

	if !c.Resource.IsPVEnabled() {
		volumes = append(volumes, c.ConstructEmptyDirVolume("rethinkdb-data"))
	}

	return volumes
}
