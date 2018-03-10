/*
A simple test for web server status. This package is intended for use with Nagios.
*/
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
)

// main expects 1 - 3 flags: -host <somehost.com> [-protocol http|https] [-port 80|443|xxx]
func main() {

	start := time.Now()

	hostPtr := flag.String("host", "unset", "A valid internet site without http:// or https://")
	protocolPtr := flag.String("protocol", "https", "Protocol - either https or http")
	portPtr := flag.String("port", "443", "Port number - default 443")

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

	// build url
	url := *protocolPtr + "://" + *hostPtr + ":" + *portPtr

	// call url
	resp, err := http.Get(url)

	// measure TTFB
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
			msg := *hostPtr + " responded with " + resp.Proto + " " + resp.Status + " with TTFB of " + strconv.FormatFloat(time.Since(start).Seconds(), 'f', 6, 64) + "s"
			messages.Ok(msg)
		}
	}

}
