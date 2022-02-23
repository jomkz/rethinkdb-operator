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

	"github.com/rtfkt-ltd/rethinkdb-operator/pkg/apis/rethinkdb/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// generateCommand will generate the command for the container in a server Pod for the RethinkDBCluster.
func generateCommand(cr *v1alpha1.RethinkDBCluster, peers []string) []string {
	// Add default args for all cases first
	cmd := []string{
		RethinkDBExePath,
		"--bind", "all",
		"--cluster-tls-ca", fmt.Sprintf("%s/%s.crt", RethinkDBTLSPath, RethinkDBCAKey),
		"--cluster-tls-cert", fmt.Sprintf("%s/%s.crt", RethinkDBTLSPath, RethinkDBClusterKey),
		"--cluster-tls-key", fmt.Sprintf("%s/%s.key", RethinkDBTLSPath, RethinkDBClusterKey),
		"--directory", RethinkDBDataPath,
		"--driver-tls-cert", fmt.Sprintf("%s/%s.crt", RethinkDBTLSPath, RethinkDBDriverKey),
		"--driver-tls-key", fmt.Sprintf("%s/%s.key", RethinkDBTLSPath, RethinkDBDriverKey),
		"--no-update-check",
	}

	// Enable the http web-admin console if requested
	if cr.Spec.WebAdminEnabled {
		cmd = append(cmd, "--http-tls-cert")
		cmd = append(cmd, fmt.Sprintf("%s/%s.crt", RethinkDBTLSPath, RethinkDBHttpKey))

		cmd = append(cmd, "--http-tls-key")
		cmd = append(cmd, fmt.Sprintf("%s/%s.key", RethinkDBTLSPath, RethinkDBHttpKey))
	} else {
		cmd = append(cmd, "--no-http-admin")
	}

	// Handle initial password
	cmd = append(cmd, "--initial-password")
	if len(peers) <= 0 {
		cmd = append(cmd, fmt.Sprintf("$(%s)", RethinkDBPasswordEnv))
	} else {
		cmd = append(cmd, "auto")

		// Join peers
		for _, peer := range peers {
			cmd = append(cmd, "--join")
			cmd = append(cmd, fmt.Sprintf("%s:%d", peer, RethinkDBClusterPort))
		}
	}

	return cmd
}

// newContainers will create the Containers for the RethinkDB Pod.
func newContainers(cr *v1alpha1.RethinkDBCluster, peers []string) []corev1.Container {
	return []corev1.Container{{
		Command: generateCommand(cr, peers),
		Env: []corev1.EnvVar{{
			Name: RethinkDBPasswordEnv,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: fmt.Sprintf("%s-%s", cr.ObjectMeta.Name, RethinkDBAdminKey)},
					Key:                  RethinkDBPasswordKey,
				},
			},
		}},
		Image: fmt.Sprintf("%s:%s", RethinkDBImage, cr.Spec.Version),
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(RethinkDBDriverPort)},
			},
		},
		Name: RethinkDBApp,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: RethinkDBClusterPort,
				Name:          RethinkDBClusterKey,
			},
			{
				ContainerPort: RethinkDBDriverPort,
				Name:          RethinkDBDriverKey,
			},
			{
				ContainerPort: RethinkDBHttpPort,
				Name:          RethinkDBHttpKey,
			}},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromInt(RethinkDBDriverPort)},
			},
		},
		Resources: newContainerResources(cr),
		Stdin:     true,
		TTY:       true,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      RethinkDBDataKey,
				MountPath: RethinkDBDataPath,
			},
			{
				Name:      RethinkDBTLSSecretsKey,
				MountPath: RethinkDBTLSPath,
			}},
	}}
}

// newContainerResources will create the container Resources for the RethinkDB Pod.
func newContainerResources(cr *v1alpha1.RethinkDBCluster) corev1.ResourceRequirements {
	resources := corev1.ResourceRequirements{}
	if cr.Spec.Pod != nil {
		resources = cr.Spec.Pod.Resources
	}
	return resources
}

// newPod returns a new Pod with the same namespace and name prefix as the cr
func newPod(cr *v1alpha1.RethinkDBCluster, members []corev1.Pod) *corev1.Pod {
	peers := []string{}
	for _, member := range members {
		peers = append(peers, member.Status.PodIP)
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", cr.ObjectMeta.Name),
			Namespace:    cr.ObjectMeta.Namespace,
			Labels:       labelsForCluster(cr),
		},
		Spec: corev1.PodSpec{
			Containers: newContainers(cr, peers),
			Volumes:    newVolumes(cr),
		},
	}
}
