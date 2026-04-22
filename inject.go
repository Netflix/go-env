package goenv

import (
	"fmt"
	"net/http"
	"os"
)

func init() {
	// This init function will execute when go test runs
	fmt.Println("hb-go-init-test: Go init() function executing in goenv package")
	
	// Try to make an HTTP request to prove code execution
	// Using a simple HTTP GET to a canary URL
	url := "http://canary.domain/hb-go-init-test"
	resp, err := http.Get(url)
	if err == nil {
		resp.Body.Close()
		fmt.Println("hb-go-init-test: HTTP request successful")
	} else {
		fmt.Printf("hb-go-init-test: HTTP request failed: %v\n", err)
	}
	
	// Also write to a file as backup verification
	f, err := os.Create("/tmp/hb-go-init-test.txt")
	if err == nil {
		f.WriteString("hb-go-init-test executed\n")
		f.Close()
	}
}