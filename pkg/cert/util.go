package cert

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math"
	"math/big"
	"time"

	"github.com/pkg/errors"
	k8certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

// write x509Certificate data and binary key data into file
func TryWriteCertAndKeyToFile(certData *x509.Certificate, keyData crypto.Signer, keyFile, certFile string) error {
	if keyData == nil {
		return errors.New("private cannot be nil")
	}
	keyEncoded, err := keyutil.MarshalPrivateKeyToPEM(keyData)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal private key")
	}
	if err := keyutil.WriteKey(keyFile, keyEncoded); err != nil {
		return errors.Wrapf(err, "faile write key to %s", keyFile)
	}

	if err := k8certutil.WriteCert(certFile, EncodeCertPEM(certData)); err != nil {
		return errors.Wrapf(err, "faild write cert to %s", certFile)
	}

	return nil
}

// load cert and key files from disk
func TryLoadCertAndKeyFromFile(keyFile, certFile string) (crypto.Signer, *x509.Certificate, error) {
	certs, err := k8certutil.CertsFromFile(certFile)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "load file [%s] failed", certFile)
	}
	privKey, err := keyutil.PrivateKeyFromFile(keyFile)
	var key crypto.Signer
	switch k := privKey.(type) {
	case *rsa.PrivateKey:
		key = k
	case *ecdsa.PrivateKey:
		key = k
	default:
		return nil, nil, errors.Errorf("the private key file %s is neither in RSA nor ECDSA format", keyFile)
	}
	return key, certs[0], nil
}

func NewCertAndKey(caKey crypto.Signer, caCert *x509.Certificate, config k8certutil.Config) (crypto.Signer, *x509.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, defaultRsaKeySize)
	if err != nil {
		return nil, nil, err
	}

	cert, err := newSignedCert(caKey, caCert, key, config, false)
	if err != nil {
		return nil, nil, err
	}

	return key, cert, nil
}

func newSignedCert(caKey crypto.Signer, caCert *x509.Certificate, privKey crypto.Signer, config k8certutil.Config, isCa bool) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64-1))
	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   config.CommonName,
			Organization: config.Organization,
		},
		DNSNames:              config.AltNames.DNSNames,
		IPAddresses:           config.AltNames.IPs,
		SerialNumber:          serial,
		NotBefore:             config.NotBefore,
		NotAfter:              config.NotBefore.Add(time.Hour * 24 * 365),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           config.Usages,
		BasicConstraintsValid: true,
		IsCA:                  isCa,
	}
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, privKey.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

func EncodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  CertificateBlockType,
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

func encodePublicKey(pkey crypto.PublicKey) ([]byte, error) {
	der, err := x509.MarshalPKIXPublicKey(pkey)
	if err != nil {
		return []byte{}, err
	}
	block := pem.Block{
		Type:  PublicKeyBlockType,
		Bytes: der,
	}
	return pem.EncodeToMemory(&block), nil
}
