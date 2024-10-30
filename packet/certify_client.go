package packet

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
)

// ReceiveCert - client side
func ReceiveCert(conn net.Conn) (*tls.Config, error) {
	cert := make([]byte, 2048)
	n, err := conn.Read(cert)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}

	fmt.Println("Received cert length:", n)

	pemBlock, _ := pem.Decode(cert[:n])
	if pemBlock == nil || pemBlock.Type != "CERTIFICATE" {
		return nil, errors.New("failed to decode PEM block")
	}

	certParsed, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certParsed)

	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}

	fmt.Println("TLS configuration ready with trusted certificate")

	return tlsConfig, nil
}
