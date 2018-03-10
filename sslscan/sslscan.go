package sslscan

import (
	//"bytes"
	//"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"checkhttp2/certificateutils"
	//"sync"
	//"time"
)

type flags struct {
	notificationThreshold int
	connectionTimeout     int
	remoteSite            string
	remoteSiteFile        string
	publicCertificate     string
}

func ReadFile(fileName string) []byte {
	hostnamesFileBytes, err := ioutil.ReadFile(fileName)

	if err != nil {
		log.Fatalf("Error reading file: %s, err: %v", fileName, err)
	}

	return hostnamesFileBytes
}


func ReadCertificateFile(flags flags) {
	certsDetails, err := certificateutils.ReadCertificateDetailsFromFile(flags.publicCertificate, "")
	if err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Printf("Found %d certificate(s) inside file: %s:\n\n", len(certsDetails), flags.publicCertificate)
	for _, certDetails := range certsDetails {
		fmt.Printf("%v\n", certDetails)
	}

	os.Exit(0)
}

func ScanHost(hostname string, certDetailsChannel chan certificateutils.CertificateDetails, errorsChannel chan error) {

	res, err := certificateutils.GetCertificateDetails(hostname, 10)
	if err != nil {
		errorsChannel <- err
	} else {
		certDetailsChannel <- res
	}
}

func UpdateSitesAndCounts(count map[string]int, sites map[string]bool, certDetails certificateutils.CertificateDetails) {
	if _, ok := count[certDetails.SubjectName]; !ok {
		count[certDetails.SubjectName] = 1
	} else {
		count[certDetails.SubjectName]++
	}

	if !sites[certDetails.Hostname] {
		sites[certDetails.Hostname] = true
	}
}

func PrintCertificateStats(count map[string]int, sites map[string]bool) {
	for cert, instanceCount := range count {
		fmt.Printf("Subject name: %s -- Instances found: %d\n", cert, instanceCount)
	}

	for hostname := range sites {
		fmt.Printf("> %s\n", hostname)
	}
	fmt.Println("")
}
