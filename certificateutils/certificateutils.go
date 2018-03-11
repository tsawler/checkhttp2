package certificateutils

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

var (
	hostnameEmptyError = errors.New("hostname empty")
)

type ResultError struct {
	Res CertificateDetails
	Err error
}

type CertificateDetails struct {
	DaysUntilExpiration int
	IssuerName          string
	SubjectName         string
	SerialNumber        string
	ExpiringSoon        bool
	Expired             bool
	Hostname            string
	TimeTaken           time.Duration
	ExpirationDate      string
	Thumbprint          string
}

func (cd CertificateDetails) String() string {
	return fmt.Sprintf(
		"Subject Name: %s\nIssuer: %s\nExpiration date: %s\nDays Until Expiration: %d\nSerial #: %s\nRequest Time: %v\n",
		cd.SubjectName,
		cd.IssuerName,
		cd.ExpirationDate,
		cd.DaysUntilExpiration,
		cd.SerialNumber,
		cd.TimeTaken,
	)
}

func insertNth(s string, n int) string {
	var buffer bytes.Buffer
	var n_1 = n - 1
	var l_1 = len(s) - 1
	for i, r := range s {
		buffer.WriteRune(r)
		if i%n == n_1 && i != l_1 {
			buffer.WriteRune('-')
		}
	}
	return buffer.String()
}

func ReadCertificateDetailsFromFile(publicCertFile, privateCertFile string) ([]CertificateDetails, error) {
	currentTime := time.Now()
	var certDetails []CertificateDetails
	var blocks []byte

	rest, err := ioutil.ReadFile(publicCertFile)
	if err != nil {
		return certDetails, err
	}

	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)

		if block == nil {
			return certDetails, errors.New("Certificate doesn't have a valid PEM block??")
		}

		blocks = append(blocks, block.Bytes...)
		if len(rest) == 0 {
			break
		}
	}

	certs, err := x509.ParseCertificates(blocks)
	if err != nil {
		return certDetails, err
	}

	for _, cert := range certs {
		daysUntilExpiration := int(cert.NotAfter.Sub(currentTime).Hours() / 24)
		subjectName := cert.Subject.Names[len(cert.Subject.Names)-1].Value.(string)
		issuerName := cert.Issuer.Names[len(cert.Issuer.Names)-1].Value.(string)
		serialNumber := cert.SerialNumber.Text(16)
		elapsed := time.Since(currentTime)

		certDetails = append(certDetails, CertificateDetails{
			DaysUntilExpiration: daysUntilExpiration,
			SubjectName:         subjectName,
			IssuerName:          issuerName,
			SerialNumber:        strings.ToUpper(insertNth(serialNumber, 2)),
			TimeTaken:           elapsed,
			ExpirationDate:      cert.NotAfter.Format(time.UnixDate),
		})

	}

	return certDetails, nil
}

func GetCertificateDetails(hostname string, connectionTimeout int) (CertificateDetails, error) {
	currentTime := time.Now()
	var certDetails CertificateDetails

	if hostname == "" {
		return CertificateDetails{}, hostnameEmptyError
	}

	if !strings.Contains(hostname, ":") {
		hostname = fmt.Sprintf("%s:443", hostname)
	}

	// Establish a new TCP connection to hostname
	// Ignore invalid certificates, so we can scan via IP addresses or hostnames
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: time.Second * time.Duration(connectionTimeout)},
		"tcp",
		hostname,
		&tls.Config{InsecureSkipVerify: true})

	if err != nil {
		return CertificateDetails{}, fmt.Errorf("Connection error: %v", err)
	}

	if handshakeCompleted := conn.ConnectionState().HandshakeComplete; !handshakeCompleted {
		return CertificateDetails{}, fmt.Errorf("TLS Handshake failed to hostname %s.", hostname)
	}

	defer conn.Close()

	// Loop through each certificate peer and determine certificate details for non-CA certificate
	for _, cert := range conn.ConnectionState().PeerCertificates {

		if cert.IsCA {
			continue
		}

		daysUntilExpiration := int(cert.NotAfter.Sub(currentTime).Hours() / 24)
		subjectName := cert.Subject.Names[len(cert.Subject.Names)-1].Value.(string)
		issuerName := cert.Issuer.Names[len(cert.Issuer.Names)-1].Value.(string)
		serialNumber := cert.SerialNumber.Text(16)
		elapsed := time.Since(currentTime)

		certDetails = CertificateDetails{
			DaysUntilExpiration: daysUntilExpiration,
			SubjectName:         subjectName,
			IssuerName:          issuerName,
			SerialNumber:        strings.ToUpper(insertNth(serialNumber, 2)),
			Hostname:            hostname,
			TimeTaken:           elapsed,
			ExpirationDate:      cert.NotAfter.Format(time.UnixDate),
		}
		break
	}

	return certDetails, nil
}

func CheckExpirationStatus(cd *CertificateDetails, expirationDaysThreshold int) {
	if cd.DaysUntilExpiration < 0 {
		cd.Expired = true
	} else if cd.DaysUntilExpiration < expirationDaysThreshold {
		cd.ExpiringSoon = true
	}
}
