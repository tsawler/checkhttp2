package main

import (
	"fmt"
	"os"
	"net/http"
)

func main() {

	url := "https://" + os.Args[1];
	resp, err := http.Get(url)
	if err != nil {
		print(err.Error())
	} else {
		if resp.StatusCode != 200 {
			fmt.Println("CRITICAL- " + os.Args[1] + " " + resp.Status)
		} else {
			fmt.Println("OK- " + os.Args[1] + " responded with " + resp.Status)
		}
	}

}
