package certificateutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func genSelfSignedCert(notBeforeDate time.Time, expirationDays int) (tls.Certificate, error) {

	// https://golang.org/src/crypto/tls/generate_cert.go
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("Unable to generate serial number for SSL certificate. Error: %v", err)
	}

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("Unable to generate private key for SSL certificate.  Error: %v", err)
	}

	if notBeforeDate == (time.Time{}) {
		notBeforeDate = time.Now()
	}
	notAfterDate := notBeforeDate.Add(time.Duration(expirationDays))

	certificateTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		NotBefore:    notBeforeDate,
		NotAfter:     notAfterDate,
		Subject: pkix.Name{
			Organization: []string{"Expiration Test Cert"},
		},
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	derBytes, _ := x509.CreateCertificate(
		rand.Reader,
		&certificateTemplate,
		&certificateTemplate,
		&rsaKey.PublicKey,
		rsaKey,
	)

	hosts := []string{"127.0.0.1", "::1", "localhost"}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			certificateTemplate.IPAddresses = append(certificateTemplate.IPAddresses, ip)
		} else {
			certificateTemplate.DNSNames = append(certificateTemplate.DNSNames, h)
		}
	}

	publicCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	privateKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(rsaKey)})

	tlsCert, err := tls.X509KeyPair(publicCert, privateKey)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("Problem generating x509 Key Pair. Error: %v", err)
	}

	return tlsCert, nil
}

func setupHTTPS(certificate tls.Certificate) (string, error) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.TLS = &tls.Config{Certificates: []tls.Certificate{certificate}}
	server.StartTLS()

	_, port, err := net.SplitHostPort(server.Listener.Addr().String())
	if err != nil {
		return "", fmt.Errorf("Unable to get port for httptest server.  Error: %v", err)
	}

	return port, nil
}

func TestExpiringCertificate(t *testing.T) {
	daysUntilExpirationThreshold := 60
	daysCertificateValid := 10

	cd := setupCertificatesAndHTTPS(time.Time{}, daysCertificateValid, t)

	CheckExpirationStatus(&cd, daysUntilExpirationThreshold)

	if !cd.ExpiringSoon {
		t.Errorf("Expected: true, got %v", cd.ExpiringSoon)
	}
}

func TestExpiredCertificate(t *testing.T) {
	daysCertificateValid := -2
	daysUntilExpirationThreshold := 60
	notBeforeDate, err := time.Parse(time.UnixDate, "Sat Mar 7 00:00:00 PST 2015")
	if err != nil {
		t.Fatalf("Error setting notBeforeDate: %v", err)
	}

	cd := setupCertificatesAndHTTPS(notBeforeDate, daysCertificateValid, t)

	CheckExpirationStatus(&cd, daysUntilExpirationThreshold)

	if !cd.Expired {
		t.Errorf("Expected: true, got: %v", cd.Expired)
	}
}

func setupCertificatesAndHTTPS(notBeforeDate time.Time, daysCertificateValid int, t *testing.T) CertificateDetails {
	cert, err := genSelfSignedCert(notBeforeDate, daysCertificateValid)
	if err != nil {
		t.Fatalf("Error generating certificate: %v", err)
	}

	port, err := setupHTTPS(cert)
	if err != nil {
		t.Fatalf("Error setting up HTTPS server: %v", err)
	}

	hostname := fmt.Sprintf("localhost:%s", port)

	cd, err := GetCertificateDetails(hostname, 5)
	if err != nil {
		t.Fatalf("Error getting certificate details from host adddress: %s", hostname)
	}

	return cd
}
