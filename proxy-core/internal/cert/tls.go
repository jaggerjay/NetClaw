package cert

import "crypto/tls"

func ConvertTLSCertificate(input *tlsCertificate) (tls.Certificate, error) {
	return tls.Certificate{
		Certificate: input.Certificate,
		PrivateKey:  input.PrivateKey,
		Leaf:        input.Leaf,
	}, nil
}
