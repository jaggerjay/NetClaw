package cert

import "crypto/tls"

func (a *Authority) TLSCertificateForHost(host string) (tls.Certificate, error) {
	material, err := a.TLSMaterialForHost(host)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: material.Certificate,
		PrivateKey:  material.PrivateKey,
		Leaf:        material.Leaf,
	}, nil
}
