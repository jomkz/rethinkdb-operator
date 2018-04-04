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
	"os"

	"github.com/coreos/operator-sdk/pkg/sdk/action"
	"github.com/coreos/operator-sdk/pkg/sdk/handler"
	"github.com/coreos/operator-sdk/pkg/sdk/types"
	v1alpha1 "github.com/jmckind/rethinkdb-operator/pkg/apis/operator/v1alpha1"
	"github.com/jmckind/rethinkdb-operator/pkg/util/k8sutil"
	"github.com/sirupsen/logrus"
	apps_v1beta2 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultConfig = `bind=all
directory=/var/lib/rethinkdb/default
`
)

func NewRethinkDBHandler() handler.Handler {
	return &RethinkDBHandler{
		namespace: os.Getenv("MY_POD_NAMESPACE"),
		kubecli:	 k8sutil.MustNewKubeClient(),
	}
}

type RethinkDBHandler struct {
	namespace string
	kubecli		kubernetes.Interface
}

func (h *RethinkDBHandler) Handle(ctx types.Context, event types.Event) []types.Action {
	var actions []types.Action

	switch o := event.Object.(type) {
	case *v1alpha1.RethinkDB:
		logrus.Infof("Received RethinkDB: %v", o.Name)
		o.SetDefaults()

		if event.Deleted {
			actions = h.DestroyRethinkDB(o)
		} else {
			actions = h.CreateRethinkDB(o)
		}
	}

	return actions
}

func (h *RethinkDBHandler) CreateRethinkDB(r *v1alpha1.RethinkDB) []types.Action {
	var actions []types.Action
	labels := map[string]string{
		"app":  "rethinkdb",
		"cluster": r.Name,
	}

	logrus.Infof("Creating RethinkDB: %v", r.Name)

	actions = append(actions, h.CreateClusterService(r, labels))
	actions = append(actions, h.CreateDriverService(r, labels))
	actions = append(actions, h.CreateStatefulSet(r, labels))

	return actions
}

func (h *RethinkDBHandler) CreateStatefulSet(r *v1alpha1.RethinkDB, labels map[string]string) types.Action {
	name := r.Name
	replicas := r.Spec.Nodes
	cluster := name + "-cluster"
	terminationSeconds := int64(5)

	logrus.Infof("Creating StatefulSet: %v", name)

	ss := &apps_v1beta2.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1beta2",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.ObjectMeta.Namespace,
			Labels: labels,
		},
		Spec: apps_v1beta2.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: cluster,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1.PodSpec{
					Containers: h.AddContainers(r),
					InitContainers: h.AddInitContainers(r),
					TerminationGracePeriodSeconds: &terminationSeconds,
					Volumes: h.AddVolumes(r),
				},
			},
			VolumeClaimTemplates: h.AddPVCs(r),
		},
	}

	return types.Action{
		Object: ss,
		Func:   action.KubeApplyFunc,
	}
}

func (h *RethinkDBHandler) AddContainers(r *v1alpha1.RethinkDB) []v1.Container {
	return []v1.Container{{
		Command: []string{
			"/usr/bin/rethinkdb",
			"--no-update-check",
			"--config-file",
			"/etc/rethinkdb/rethinkdb.conf",
		},
		Image: fmt.Sprintf("%s:%s", r.Spec.BaseImage, r.Spec.Version),
		Name: "rethinkdb",
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
		Resources: r.Spec.Pod.Resources,
		Stdin: true,
		TTY: true,
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

func (h *RethinkDBHandler) AddInitContainers(r *v1alpha1.RethinkDB) []v1.Container {
	name := r.Name
	cluster := name + "-cluster"

	return []v1.Container{{
		Command: []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf("echo '%s' > /etc/rethinkdb/rethinkdb.conf; if nslookup %s; then echo join=%s-0.%s:29015 >> /etc/rethinkdb/rethinkdb.conf; fi;", defaultConfig, cluster, name, cluster),
		},
		Image: "busybox:latest",
		Name: "cluster-init",
		VolumeMounts: []v1.VolumeMount{{
			Name:      "rethinkdb-etc",
			MountPath: "/etc/rethinkdb",
		}},
	}}
}

func (h *RethinkDBHandler) AddVolumes(r *v1alpha1.RethinkDB) []v1.Volume {
	var volumes []v1.Volume

	volumes = append(volumes, h.AddEmptyDirVolume("rethinkdb-etc"))

	if !r.IsPVEnabled() {
		volumes = append(volumes, h.AddEmptyDirVolume("rethinkdb-data"))
	}

	return volumes
}

func (h *RethinkDBHandler) AddEmptyDirVolume(name string) v1.Volume {
	return v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
}

func (h *RethinkDBHandler) AddPVCs(r *v1alpha1.RethinkDB) []v1.PersistentVolumeClaim {
	var pvcs []v1.PersistentVolumeClaim

	if r.IsPVEnabled() {
		pvcs = append(pvcs, v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rethinkdb-data",
				Namespace: r.ObjectMeta.Namespace,
				Labels: r.ObjectMeta.Labels,
			},
			Spec: *r.Spec.Pod.PersistentVolumeClaimSpec,
		})
	}

	return pvcs
}

func (h *RethinkDBHandler) CreateClusterService(r *v1alpha1.RethinkDB, labels map[string]string) types.Action {
	name := r.Name + "-cluster"

	logrus.Infof("Creating Cluster Service: %v", name)

	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: r.ObjectMeta.Namespace,
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Selector: labels,
			Ports: []v1.ServicePort{{
				Port: 29015,
				Name: "cluster",
			}},
		},
	}

	return types.Action{
		Object: svc,
		Func:   action.KubeApplyFunc,
	}
}

func (h *RethinkDBHandler) CreateDriverService(r *v1alpha1.RethinkDB, labels map[string]string) types.Action {
	logrus.Infof("Creating Driver Service: %v", r.Name)

	svc := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.ObjectMeta.Namespace,
			Labels: labels,
		},
		Spec: v1.ServiceSpec{
			Selector: labels,
			SessionAffinity: "ClientIP",
			Type: "NodePort",
			Ports: []v1.ServicePort{{
				Port: 8080,
				Name: "http",
			},
			{
				Port: 28015,
				Name: "driver",
			}},
		},
	}

	return types.Action{
		Object: svc,
		Func:   action.KubeApplyFunc,
	}
}

// TODO: remove this function when CRD GC is enabled.
func (h *RethinkDBHandler) DestroyRethinkDB(r *v1alpha1.RethinkDB) []types.Action {
	var actions []types.Action

	logrus.Infof("Destroying RethinkDB: %v", r.Name)

	action := h.DestroyRethinkDBDeployment(r)
	if action != nil {
			actions = append(actions, *action)
	}

	action = h.DestroyServices(r)
	if action != nil {
			actions = append(actions, *action)
	}

	return actions
}

func (h *RethinkDBHandler) DestroyRethinkDBDeployment(r *v1alpha1.RethinkDB) *types.Action {
	name := r.Name
	namespace := r.ObjectMeta.Namespace

	deployment, err := h.kubecli.AppsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		logrus.Warnf("Deployment not found: %v", name)
		return nil
	} else if err != nil {
		logrus.Errorf("Error deleting Deployment: %v", err)
		return nil
	}

	logrus.Infof("Destroying Deployment: %v", name)

	return &types.Action{
		Object: deployment,
		Func:   action.KubeDeleteFunc,
	}
}

func (h *RethinkDBHandler) DestroyServices(r *v1alpha1.RethinkDB) *types.Action {
	name := r.Name
	namespace := r.ObjectMeta.Namespace

	svc, err := h.kubecli.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		logrus.Warnf("Service not found: %v", name)
		return nil
	} else if err != nil {
		logrus.Errorf("Error deleting Service: %v", err)
		return nil
	}

	logrus.Infof("Destroying Service: %v", name)

	return &types.Action{
		Object: svc,
		Func:   action.KubeDeleteFunc,
	}
}
