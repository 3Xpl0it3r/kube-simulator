package cert

import (
	"crypto/rand"
	"crypto/rsa"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/keyutil"

	k8certutil "k8s.io/client-go/util/cert"
)

const (
	CertificateBlockType = "CERTIFICATE"
	PublicKeyBlockType   = "PUBLIC KEY"
	defaultRsaKeySize    = 2048
)

// CertKeyPair represent certpair
type CertKeyPair struct {
	Name     string
	KeyFile  string
	CertFile string
}

func CreateCACertFiles(caCert CertKeyPair, config k8certutil.Config) error {
	if _, _, err := TryLoadCertAndKeyFromFile(caCert.KeyFile, caCert.CertFile); err == nil {
		return nil
	}

	key, err := rsa.GenerateKey(rand.Reader, defaultRsaKeySize)
	if err != nil {
		return errors.Wrapf(err, "failed generate ras key for %s", caCert.KeyFile)
	}

	certificate, err := k8certutil.NewSelfSignedCACert(config, key)
	if err != nil {
		return errors.Wrapf(err, "failed gen self-signed cert for %s", caCert.CertFile)
	}

	if err = TryWriteCertAndKeyToFile(certificate, key, caCert.KeyFile, caCert.CertFile); err != nil {
		return errors.Wrapf(err, "try write key/cert to file failed")
	}

	return nil
}

func CreateGenericCertFiles(serverCert, caCert CertKeyPair, config k8certutil.Config) error {
	// try load server key/certs from files first, if load success then return directlly
	if _, _, err := TryLoadCertAndKeyFromFile(serverCert.KeyFile, serverCert.CertFile); err == nil {
		return nil
	}
	// load ca key and ca certs from file, if load failed, then return error
	caKeyData, caCertData, err := TryLoadCertAndKeyFromFile(caCert.KeyFile, caCert.CertFile)
	if err != nil {
		return errors.Wrapf(err, "failed load ca certificated from %s/%s", caCert.KeyFile, caCert.CertFile)
	}

	keyData, certData, err := NewCertAndKey(caKeyData, caCertData, config)
	if err != nil {
		return errors.Wrapf(err, "[%s] failed to generate newCertAndKey.", serverCert.KeyFile)
	}

	if err = TryWriteCertAndKeyToFile(certData, keyData, serverCert.KeyFile, serverCert.CertFile); err != nil {
		return errors.Wrapf(err, "[%s] cannot write certs to file", serverCert.CertFile)
	}

	return nil
}

func CreateServiceAccountKeyAndPublicKeyFiles(privKeyFile, pubKeyFile string) error {
	if _, err := keyutil.PrivateKeyFromFile(privKeyFile); err == nil {
		return nil
	}
	key, err := rsa.GenerateKey(rand.Reader, defaultRsaKeySize)
	if err != nil {
		return errors.Wrap(err, "cannot genetate new rsa_key")
	}
	keyBytes, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return errors.Wrap(err, "faild marshal privKey to bytes")
	}
	if err := keyutil.WriteKey(privKeyFile, keyBytes); err != nil {
		return errors.Wrapf(err, "can't write key to %s", privKeyFile)
	}

	pkeyBytes, err := encodePublicKey(key.Public())
	if err != nil {
		return errors.Wrap(err, "failed encode publicKey")
	}

	if err := keyutil.WriteKey(pubKeyFile, pkeyBytes); err != nil {
		return errors.Wrapf(err, "can't wrtie public key to %s", pubKeyFile)
	}

	return nil
}
