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
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/jmckind/rethinkdb-operator/pkg/apis/rethinkdb/v1alpha1"
	tlsutil "github.com/operator-framework/operator-sdk/pkg/tls"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newCASecret creates a new CA secret for the given RethinkDBCluster.
func newCASecret(cr *v1alpha1.RethinkDBCluster, name string) (*corev1.Secret, error) {
	secret := newTLSSecret(cr, name)

	key, err := newPrivateKey()
	if err != nil {
		return nil, err
	}

	cert, err := newSelfSignedCACertificate(key)
	if err != nil {
		return nil, err
	}

	secret.Data = map[string][]byte{
		corev1.TLSCertKey:       encodeCertificatePEM(cert),
		corev1.TLSPrivateKeyKey: encodePrivateKeyPEM(key),
	}

	return secret, nil
}

// newCertificateSecret creates a new secret for a TLS certificate.
func newCertificateSecret(cr *v1alpha1.RethinkDBCluster, name string, caCert *x509.Certificate, caKey *rsa.PrivateKey) (*corev1.Secret, error) {
	secret := newTLSSecret(cr, name)

	key, err := newPrivateKey()
	if err != nil {
		return nil, err
	}

	cfg := &tlsutil.CertConfig{
		CertName:     name,
		CertType:     tlsutil.ClientAndServingCert,
		CommonName:   name,
		Organization: []string{cr.ObjectMeta.Namespace},
	}

	dnsNames := []string{fmt.Sprintf("%s.%s.svc.cluster.local", cr.ObjectMeta.Name, cr.ObjectMeta.Namespace)}
	cert, err := newSignedCertificate(cfg, dnsNames, key, caCert, caKey)
	if err != nil {
		return nil, err
	}

	secret.Data = map[string][]byte{
		corev1.TLSCertKey:       encodeCertificatePEM(cert),
		corev1.TLSPrivateKeyKey: encodePrivateKeyPEM(key),
	}

	return secret, nil
}

// newSecret creates a new secret for the given RethinkDBCluster.
func newSecret(cr *v1alpha1.RethinkDBCluster) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labelsForCluster(cr),
		},
		Type: corev1.SecretTypeOpaque,
	}
}

// newTLSSecret creates a new TLS secret with the given name for the given RethinkDBCluster.
func newTLSSecret(cr *v1alpha1.RethinkDBCluster, name string) *corev1.Secret {
	secret := newSecret(cr)
	secret.ObjectMeta.Name = name
	secret.Type = corev1.SecretTypeTLS
	return secret
}
