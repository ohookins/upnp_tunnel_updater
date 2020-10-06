package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const tunnelUpdateFmt = "https://ipv4.tunnelbroker.net/nic/update?myip=%s&username=%s&password=%s&hostname=%s"

func tunnelBrokerUpdate(address string) error {
	requestURI := fmt.Sprintf(tunnelUpdateFmt, address, *userID, *password, *tunnelID)
	log.Printf("Requesting update to tunnel %s config with IP: %s", *tunnelID, address)

	res, err := http.Get(requestURI)
	if err != nil {
		return fmt.Errorf("error requesting tunnel update: %s", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("tunnel config update failed: %d %s", res.StatusCode, body)
	}

	// Check for errors in the body
	if strings.Contains(string(body), "ERROR") {
		return fmt.Errorf(string(body))
	}

	log.Printf("successfully updated tunnel config")
	return nil
}
