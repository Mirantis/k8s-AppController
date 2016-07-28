package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: check-status POD-NAME")
	}

	podName := os.Args[1]

	url := fmt.Sprintf("http://127.0.0.1:8080/api/v1/namespaces/default/pods/%s", podName)
	log.Printf("Requesting %s", url)
	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Printf("Response: %s", body)

	if err != nil {
		log.Fatal(err)
	}
}
