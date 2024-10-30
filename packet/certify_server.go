package packet

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

// Server side
func EstablishServerTLS(addr string) (net.Conn, error) {
	// Генеруємо RSA ключі
	privateKey, publicKey, err := GenerateKeys()
	if err != nil {
		return nil, err
	}

	// Слухаємо на порту
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer listener.Close()

	// Приймаємо з'єднання
	conn, err := listener.Accept()
	if err != nil {
		return nil, err
	}
	remoteAddr := conn.RemoteAddr()

	// Завантажуємо або генеруємо TLS-сертифікат
	tlsCert := LoadTLSCert()

	// Шифруємо сертифікат
	encryptedCert, err := Encrypt(tlsCert.Certificate, publicKey)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	// Передаємо зашифрований сертифікат клієнту
	err = SendCert(conn, encryptedCert)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	_ = conn.Close() // Закриваємо початкове з'єднання

	// Слухаємо TLS-з'єднання на тому ж порту
	tlsListener, err := tls.Listen("tcp", addr, &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})
	if err != nil {
		return nil, err
	}
	defer tlsListener.Close()

	// Приймаємо TLS-з'єднання
	tlsConn, err := tlsListener.Accept()
	if err != nil {
		return nil, err
	}

	// Перевіряємо, що віддалена адреса збігається
	if tlsConn.RemoteAddr().String() != remoteAddr.String() {
		_ = tlsConn.Close()
		return nil, errors.New("TLS connection remote address mismatch")
	}

	// Проводимо handshake
	if tlsConnection, ok := tlsConn.(*tls.Conn); ok {
		err = tlsConnection.Handshake()
		if err != nil {
			_ = tlsConn.Close()
			return nil, fmt.Errorf("TLS handshake failed: %w", err)
		}
		fmt.Println("Secure TLS connection established")
		return tlsConnection, nil
	}

	_ = tlsConn.Close()
	return nil, errors.New("failed to establish secure TLS connection")
}

func SendCert(conn net.Conn) error {
	cert, err := os.ReadFile("server.crt")
	if err != nil {
		return err
	}

	n, err := conn.Write(cert)
	if err != nil {
		return err
	}

	if n != len(cert) {
		return errors.New("short write")
	}
	fmt.Println("cert send over network with bytes: ", n)
	return nil
}

func LoadTLSCert() tls.Certificate {
	var (
		err  error
		cert tls.Certificate
	)
	cert, err = tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		fmt.Println("cert files not found, generate new one")
		cert, err = generateCert()
		if err != nil {
			panic(err)
		}
	}

	return cert
}

func generateCert() (tls.Certificate, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"EternalStorage"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certOut, err := os.Create("server.crt")
	if err != nil {
		return tls.Certificate{}, err
	}
	defer certOut.Close()
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if err != nil {
		return tls.Certificate{}, err
	}
	fmt.Println("server cert written to: ", certOut.Name())

	keyOut, err := os.Create("server.key")
	if err != nil {
		return tls.Certificate{}, err
	}
	defer keyOut.Close()
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		return tls.Certificate{}, err
	}
	fmt.Println("server key written to: ", keyOut.Name())

	return tls.LoadX509KeyPair("server.crt", "server.key")
}
