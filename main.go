// A simple nagios plugin to test for HTTP/2 status and SSL expiration
package main

import (
	"fmt"
	"net/http"
	"flag"
	"os"
	"strings"
	"github.com/pkg/errors"
	"time"
	"strconv"
	"checkhttp2/messages"
	"checkhttp2/certificateutils"
)

// scanHost gets cert details from an internet host
func scanHost(hostname string, certDetailsChannel chan certificateutils.CertificateDetails, errorsChannel chan error) {

	res, err := certificateutils.GetCertificateDetails(hostname, 10)
	if err != nil {
		errorsChannel <- err
	} else {
		certDetailsChannel <- res
	}
}

// main expects 1-4 flags: -host <somehost.com> [-protocol http|https] [-port 80|443|xxx] [-cert true|false (default)\
func main() {

	hostPtr := flag.String("host", "unset", "A valid internet site without http:// or https://")
	protocolPtr := flag.String("protocol", "https", "Protocol - either https or http")
	portPtr := flag.String("port", "443", "Port number - default 443")
	certPtr := flag.String("cert", "false", "Scan SSL Cert - default false")

	flag.Parse()

	if strings.Compare(*hostPtr, "unset") == 0 {
		fmt.Println("Usage: checkhttp2 -host somehost.com [ -protocol http|https] [-port 80|443|xxx]")
		os.Exit(0)
	}

	if strings.Compare(*protocolPtr, "http") == 0 && strings.Compare(*portPtr, "443") == 0 {
		msg := "Protocol http specified, but port 443 chosen as default. Did you forget -port 80?"
		err := errors.New(msg)
		messages.Critical(err)
	}

	scancert := false
	scancert, _ = strconv.ParseBool(*certPtr)

	if scancert == false {

		// build url
		url := *protocolPtr + "://" + *hostPtr + ":" + *portPtr

		// call url & measure TTFB
		start := time.Now()
		resp, err := http.Get(url)
		defer resp.Body.Close()

		oneByte := make([]byte, 1)
		_, err = resp.Body.Read(oneByte)

		if err != nil {
			messages.Critical(err)
		} else {
			if resp.StatusCode == 503 {
				msg := *hostPtr + " " + resp.Proto + " " + resp.Status
				messages.Warning(msg)
			}
			if resp.StatusCode != 200 {
				msg := *hostPtr + " " + resp.Proto + " " + resp.Status
				err = errors.New(msg)
				messages.Critical(err)
			} else {
				elapsed := strconv.FormatFloat(time.Since(start).Seconds(), 'f', 6, 64)
				msg := *hostPtr + " responded with " + resp.Proto + " " + resp.Status + " with TTFB of " + elapsed + "s"
				messages.Ok(msg)
			}
		}
	} else {
		// scanning ssl cert for expiry date
		var certDetailsChannel chan certificateutils.CertificateDetails
		var errorsChannel chan error
		certDetailsChannel = make(chan certificateutils.CertificateDetails, 1)
		errorsChannel = make(chan error, 1)

		scanHost(*hostPtr, certDetailsChannel, errorsChannel)

		for i, certDetailsInQueue := 0, len(certDetailsChannel); i < certDetailsInQueue; i++ {
			certDetails := <-certDetailsChannel
			certificateutils.CheckExpirationStatus(&certDetails, 30)

			if certDetails.ExpiringSoon {

				if certDetails.DaysUntilExpiration < 7 {
					msg := certDetails.Hostname + " expiring in " + strconv.Itoa(certDetails.DaysUntilExpiration) + " days"
					err := errors.New(msg)
					messages.Critical(err)
				} else {
					msg := certDetails.Hostname + " expiring in " + strconv.Itoa(certDetails.DaysUntilExpiration) + " days"
					messages.Warning(msg)
				}

			} else {
				msg := certDetails.Hostname + " expiring in " + strconv.Itoa(certDetails.DaysUntilExpiration) + " days"
				messages.Ok(msg)
			}

		}

		if len(errorsChannel) > 0 {
			fmt.Printf("There were %d error(s):\n", len(errorsChannel))
			for i, errorsInChannel := 0, len(errorsChannel); i < errorsInChannel; i++ {
				fmt.Printf("%s\n", <-errorsChannel)
			}
			fmt.Printf("\n")
		}
	}

}
