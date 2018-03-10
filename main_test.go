package main

import (
	"testing"

	"checkhttp2/certificateutils"
)

func TestScanhost(t *testing.T) {
	hostname := "www.google.com"
	res, _ := certificateutils.GetCertificateDetails(hostname, 10)

	if res.Hostname != "www.google.com:443" {
		t.Error("Expected www.google.com:443, got ", res.Hostname)
	}

}
