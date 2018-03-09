/*
A simple test for web server status. This package is intended for use with Nagios.
*/
package main

import (
	"fmt"
	"os"
	"net/http"
)

// main expects one parameter on the command line: a valid website name.
// This host is called using https, and returns OK and the status if the status is 200, or
// Critical and the status if it's anything else.
func main() {

	url := "https://" + os.Args[1];
	resp, err := http.Get(url)

	if err != nil {
		fmt.Println("CRITICAL- site is unreachable!")
	} else {
		if resp.StatusCode != 200 {
			fmt.Println("CRITICAL- " + os.Args[1] + " " + resp.Status)
		} else {
			fmt.Println("OK- " + os.Args[1] + " responded with " + resp.Status)
		}
	}

}
