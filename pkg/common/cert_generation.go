package common

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"net/url"
)

func NewCSRTemplate(spiffeID string) ([]byte, crypto.PublicKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	csr, err := NewCSRTemplateWithKey(spiffeID, key)
	if err != nil {
		return nil, nil, err
	}
	return csr, key.Public(), nil
}

func NewCSRTemplateWithKey(spiffeID string, key crypto.Signer) ([]byte, error) {
	uriSAN, err := url.Parse(spiffeID)
	if err != nil {
		return nil, err
	}
	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"SPIFFE_FOG"},
		},
		URIs: []*url.URL{uriSAN},
	}
	return x509.CreateCertificateRequest(rand.Reader, template, key)
}
