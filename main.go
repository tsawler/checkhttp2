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
)

type NagiosStatusVal int

const (
	NAGIOS_OK NagiosStatusVal = iota
	NAGIOS_WARNING
	NAGIOS_CRITICAL
	NAGIOS_UNKNOWN
)

var (
	valMessages = []string{
		"OK:",
		"WARNING:",
		"CRITICAL:",
		"UNKNOWN:",
	}
)

// Take a bunch of NagiosStatus pointers and find the highest value, then
// combine all the messages. Things win in the order of highest to lowest.
type NagiosStatus struct {
	Message string
	Value   NagiosStatusVal
}


func (status *NagiosStatus) Aggregate(otherStatuses []*NagiosStatus) {
	for _, s := range otherStatuses {
		if status.Value < s.Value {
			status.Value = s.Value
		}

		status.Message += " - " + s.Message
	}
}

// Exit with an UNKNOWN status and appropriate message
func Unknown(output string) {
	ExitWithStatus(&NagiosStatus{output, NAGIOS_UNKNOWN})
}

// Exit with an CRITICAL status and appropriate message
func Critical(err error) {
	ExitWithStatus(&NagiosStatus{err.Error(), NAGIOS_CRITICAL})
}

// Exit with an WARNING status and appropriate message
func Warning(output string) {
	ExitWithStatus(&NagiosStatus{output, NAGIOS_WARNING})
}

// Exit with an OK status and appropriate message
func Ok(output string) {
	ExitWithStatus(&NagiosStatus{output, NAGIOS_OK})
}

// Exit with a particular NagiosStatus
func ExitWithStatus(status *NagiosStatus) {
	fmt.Fprintln(os.Stdout, valMessages[status.Value], status.Message)
	os.Exit(int(status.Value))
}


// main expects one parameter on the command line: a valid website name.
// This host is called using https, and returns OK and the status if the status is 200, or
// Critical and the status if it's anything else.
func main() {

	hostPtr := flag.String("host", "unset", "A valid internet site without http:// or https://")
	protocolPtr := flag.String("protocol", "https", "Protocol - either https or http")

	flag.Parse()

	if strings.Compare(*hostPtr, "unset") == 0 {
		fmt.Println("Usage: checkHttp2 -host somehost.com [ -protocol http|https]")
		os.Exit(0)
	}

	url := *protocolPtr + "://" +  *hostPtr

	resp, err := http.Get(url)

	if err != nil {
		Critical(err)

	} else {
		if resp.StatusCode == 503 {
			msg := *hostPtr + " " + resp.Proto + " " + resp.Status
			Warning(msg)
		}
		if resp.StatusCode != 200 {
			msg := *hostPtr + " " + resp.Proto + " " + resp.Status
			err = errors.New(msg)
			Critical(err)
		} else {
			msg := *hostPtr + " responded with " + resp.Proto + " " + resp.Status
			Ok(msg)
		}
	}

}
