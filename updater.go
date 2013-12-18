package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func tunnelBrokerUpdate(address string) error {
	requestURI := fmt.Sprintf("https://ipv4.tunnelbroker.net/ipv4_end.php?ipv4b=%s&user_id=%s&pass=%s&tunnel_id=%s",
		address,
		*userID,
		*password,
		*tunnelID,
	)
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

	log.Printf("successfully updated tunnel config")
	log.Printf(string(body))
	return nil
}
