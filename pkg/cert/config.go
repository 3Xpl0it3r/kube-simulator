package cert

import (
	"crypto/x509"
	"time"

	k8certutil "k8s.io/client-go/util/cert"
)

func NewCACertificateConfig(commonName string, organizations ...string) k8certutil.Config {
	return k8certutil.Config{
		NotBefore:    time.Now().Add(-10 * time.Second),
		CommonName:   commonName,
		Organization: organizations,
	}
}

func NewServerCerfiticateConfig(commondName string, altNames k8certutil.AltNames, organizations ...string) k8certutil.Config {
	return k8certutil.Config{
		NotBefore:    time.Now(),
		CommonName:   commondName,
		Organization: organizations,
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		AltNames:     altNames,
	}
}

func NewClientCertificateConfig(commonName string, organizations ...string) k8certutil.Config {
	return k8certutil.Config{
		CommonName:   commonName,
		Organization: organizations,
		NotBefore:    time.Now(),
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
}
