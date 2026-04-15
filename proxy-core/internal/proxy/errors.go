package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

func statusCodeForUpstreamError(err error) int {
	if err == nil {
		return http.StatusBadGateway
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return http.StatusGatewayTimeout
	}

	var unknownAuthorityErr x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthorityErr) {
		return http.StatusBadGateway
	}

	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		return http.StatusBadGateway
	}

	var invalidCertErr x509.CertificateInvalidError
	if errors.As(err, &invalidCertErr) {
		return http.StatusBadGateway
	}

	var recordHeaderErr tls.RecordHeaderError
	if errors.As(err, &recordHeaderErr) {
		return http.StatusBadGateway
	}

	return http.StatusBadGateway
}

func describeUpstreamError(err error) string {
	if err == nil {
		return ""
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Sprintf("upstream timeout: %v", err)
	}

	var unknownAuthorityErr x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthorityErr) {
		return fmt.Sprintf("upstream certificate not trusted: %v", err)
	}

	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		return fmt.Sprintf("upstream hostname validation failed: %v", err)
	}

	var invalidCertErr x509.CertificateInvalidError
	if errors.As(err, &invalidCertErr) {
		return fmt.Sprintf("upstream certificate invalid: %v", err)
	}

	var recordHeaderErr tls.RecordHeaderError
	if errors.As(err, &recordHeaderErr) {
		return fmt.Sprintf("upstream TLS protocol error: %v", err)
	}

	message := err.Error()
	lower := strings.ToLower(message)
	if strings.Contains(lower, "tls") || strings.Contains(lower, "x509") {
		return fmt.Sprintf("upstream TLS error: %v", err)
	}
	if strings.Contains(lower, "connection refused") {
		return fmt.Sprintf("upstream connection refused: %v", err)
	}
	return fmt.Sprintf("upstream request failed: %v", err)
}
