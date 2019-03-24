// Package messages is based on https://github.com/newrelic/go_nagios
package messages

import (
	"fmt"
	"os"
)

// NagiosStatusVal holds the status value
type NagiosStatusVal int

const (
	// NAGIOS_OK is ok
	NAGIOS_OK NagiosStatusVal = iota
	// NAGIOS_WARNING - warning error
	NAGIOS_WARNING
	// NAGIOS_CRITICAL - critical error
	NAGIOS_CRITICAL
	// NAGIOS_UNKNOWN - unknown error
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

// NagiosStatus holds the message value and string
type NagiosStatus struct {
	Message string
	Value   NagiosStatusVal
}

// Aggregate takes a bunch of NagiosStatus pointers and finds the highest value, then
// combine all the messages. Things win in the order of highest to lowest.
func (status *NagiosStatus) Aggregate(otherStatuses []*NagiosStatus) {
	for _, s := range otherStatuses {
		if status.Value < s.Value {
			status.Value = s.Value
		}

		status.Message += " - " + s.Message
	}
}

// Unknown - Exit with an UNKNOWN status and appropriate message
func Unknown(output string) {
	ExitWithStatus(&NagiosStatus{output, NAGIOS_UNKNOWN})
}

// Critical - Exit with an CRITICAL status and appropriate message
func Critical(err error) {
	ExitWithStatus(&NagiosStatus{err.Error(), NAGIOS_CRITICAL})
}

// Warning - Exit with an WARNING status and appropriate message
func Warning(output string) {
	ExitWithStatus(&NagiosStatus{output, NAGIOS_WARNING})
}

// Ok - Exit with an OK status and appropriate message
func Ok(output string) {
	ExitWithStatus(&NagiosStatus{output, NAGIOS_OK})
}

// ExitWithStatus - Exit with a particular NagiosStatus
func ExitWithStatus(status *NagiosStatus) {
	fmt.Fprintln(os.Stdout, valMessages[status.Value], status.Message)
	os.Exit(int(status.Value))
}
