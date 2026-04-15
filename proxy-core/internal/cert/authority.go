package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Authority struct {
	mu         sync.RWMutex
	storageDir string
	certPath   string
	keyPath    string
	rootCert   *x509.Certificate
	rootKey    *rsa.PrivateKey
	cache      map[string]*tlsMaterial
}

type tlsMaterial struct {
	certificate tlsCertificate
	leaf        *x509.Certificate
}

type tlsCertificate struct {
	Certificate [][]byte
	PrivateKey  *rsa.PrivateKey
	Leaf        *x509.Certificate
}

type Info struct {
	StorageDir      string `json:"storageDir"`
	CertificatePath string `json:"certificatePath"`
	PrivateKeyPath  string `json:"privateKeyPath"`
	CommonName      string `json:"commonName"`
	Trusted         bool   `json:"trusted"`
}

func NewAuthority(storageDir string) (*Authority, error) {
	if storageDir == "" {
		return nil, fmt.Errorf("storage dir is required")
	}
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return nil, err
	}

	a := &Authority{
		storageDir: storageDir,
		certPath:   filepath.Join(storageDir, "netclaw-root-ca.pem"),
		keyPath:    filepath.Join(storageDir, "netclaw-root-ca.key"),
		cache:      make(map[string]*tlsMaterial),
	}

	if err := a.ensureRootCA(); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Authority) ensureRootCA() error {
	if fileExists(a.certPath) && fileExists(a.keyPath) {
		return a.loadRootCA()
	}
	return a.generateRootCA()
}

func (a *Authority) Info() Info {
	a.mu.RLock()
	defer a.mu.RUnlock()

	commonName := ""
	if a.rootCert != nil {
		commonName = a.rootCert.Subject.CommonName
	}

	return Info{
		StorageDir:      a.storageDir,
		CertificatePath: a.certPath,
		PrivateKeyPath:  a.keyPath,
		CommonName:      commonName,
		Trusted:         false,
	}
}

func (a *Authority) TLSMaterialForHost(host string) (*tlsCertificate, error) {
	host = normalizeHost(host)
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	a.mu.RLock()
	cached, ok := a.cache[host]
	a.mu.RUnlock()
	if ok {
		return &cached.certificate, nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if cached, ok := a.cache[host]; ok {
		return &cached.certificate, nil
	}

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	serialNumber, err := randSerialNumber()
	if err != nil {
		return nil, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   host,
			Organization: []string{"NetClaw Local Intercept"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(7 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	if ip := net.ParseIP(host); ip != nil {
		tmpl.IPAddresses = []net.IP{ip}
	} else {
		tmpl.DNSNames = []string{host}
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, a.rootCert, &leafKey.PublicKey, a.rootKey)
	if err != nil {
		return nil, err
	}

	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, err
	}

	material := &tlsMaterial{
		certificate: tlsCertificate{
			Certificate: [][]byte{der, a.rootCert.Raw},
			PrivateKey:  leafKey,
			Leaf:        leaf,
		},
		leaf: leaf,
	}
	a.cache[host] = material
	return &material.certificate, nil
}

func (a *Authority) loadRootCA() error {
	certPEM, err := os.ReadFile(a.certPath)
	if err != nil {
		return err
	}
	keyPEM, err := os.ReadFile(a.keyPath)
	if err != nil {
		return err
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("failed to decode root certificate PEM")
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return fmt.Errorf("failed to decode root key PEM")
	}

	certObj, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return err
	}
	keyObj, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.rootCert = certObj
	a.rootKey = keyObj
	return nil
}

func (a *Authority) generateRootCA() error {
	rootKey, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		return err
	}

	serialNumber, err := randSerialNumber()
	if err != nil {
		return err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "NetClaw Root CA",
			Organization: []string{"NetClaw"},
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().AddDate(5, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLenZero:        true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &rootKey.PublicKey, rootKey)
	if err != nil {
		return err
	}

	rootCert, err := x509.ParseCertificate(der)
	if err != nil {
		return err
	}

	certFile, err := os.OpenFile(a.certPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		return err
	}

	keyFile, err := os.OpenFile(a.keyPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rootKey)}); err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.rootCert = rootCert
	a.rootKey = rootKey
	return nil
}

func randSerialNumber() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, limit)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if strings.HasPrefix(host, "[") && strings.Contains(host, "]") {
		host = strings.TrimPrefix(host, "[")
		host = strings.TrimSuffix(host, "]")
	}
	if strings.Contains(host, ":") {
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			return parsedHost
		}
	}
	return host
}
