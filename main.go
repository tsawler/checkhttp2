// A simple nagios plugin to test for HTTP/2 status and SSL expiration
package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tsawler/checkhttp2/certificateutils"
	"github.com/tsawler/checkhttp2/messages"
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

// main expects 1-5 flags: -host <somehost.com> [-protocol http|https] [-port 80|443] [-cert]
func main() {

	hostPtr := flag.String("host", "", "A valid internet site e.g. www.example.com")
	protocolPtr := flag.String("protocol", "https", "Protocol - https or http")
	portPtr := flag.Int("port", 443, "Port number")
	certPtr := flag.Bool("cert", false, "If set, perform scan SSL cert only")
	pagePtr := flag.String("page", "", "Specific page to scan")

	flag.Parse()

	if strings.Compare(*hostPtr, "") == 0 {
		fmt.Println("Usage: checkhttp2 -host example.com [-protocol http|https] [-port 80|443] [-cert]")
		os.Exit(0)
	}

	if strings.Compare(*protocolPtr, "http") == 0 && *portPtr == 443 {
		msg := "protocol http specified, but port 443 chosen as default - ensure that -port is set"
		err := errors.New(msg)
		messages.Critical(err)
	}

	if *certPtr == false {
		// checking http/http2 connectivity and TTFB
		url := fmt.Sprintf("%s://%s:%d%s", *protocolPtr, *hostPtr, *portPtr, *pagePtr)

		// call url & measure TTFB
		start := time.Now()
		resp, err := http.Get(url)
		if err != nil {
			messages.Critical(err)
		}
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
