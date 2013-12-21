package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"
)

var (
	timeout   = flag.Duration("timeout", 10*time.Second, "Discovery timeout duration")
	userID    = flag.String("user-id", "", "Tunnelbroker User ID")
	password  = flag.String("password", "", "Tunnelbroker password")
	tunnelID  = flag.String("tunnel-id", "", "Tunnelbroker tunnel ID")
	noopMode  = flag.Bool("noop", false, "Do not actually update the Tunnelbroker config")
	cacheFile = flag.String("cache-file", ".upnp_tunnel_updater.cache", "Cache file that stores the previously updated IP.")
)

func main() {
	// Check all necessary flags
	flag.Parse()
	if !*noopMode && (*userID == "" || *password == "" || *tunnelID == "") {
		log.Fatalf("Require user-id, password and tunnel-id to update Tunnelbroker config")
	}

	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		log.Fatalf("error creating UDP listener: %s", err.Error())
	}
	udp := conn.(*net.UDPConn)
	log.Printf("listening on UDP %s", udp.LocalAddr())

	decodeChan := make(chan []byte)
	ssdpChan := make(chan map[string]string)
	go msgDecoder(decodeChan, ssdpChan)
	go UPNPListener(udp, decodeChan)

	ssdp, _ := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	query := ssdpQueryMessage(*timeout)
	_, err = udp.WriteToUDP(query.Bytes(), ssdp)
	if err != nil {
		log.Fatalf("error during UPNP discovery message: %s", err.Error())
	}

	log.Printf("sent SSDP discovery message (timeout: %s)", *timeout)

	// Only wait for the first message or timeout
	var msg map[string]string
	select {
	case msg = <-ssdpChan:
		close(ssdpChan)
	case <-time.After(*timeout):
		log.Fatalf("timed out after %s waiting for discovery response", *timeout)
	}

	// Find the control point from the discovery response
	controlPoint, err := getWANControlPoint(msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("received control point: %s", controlPoint)

	// Hit the control point
	wanIP := retrieveWANIP(controlPoint)
	log.Printf("Current WAN IP is: %s", wanIP)

	// Check the cache
	if !hasCurrentIPChanged(wanIP, *cacheFile) {
		log.Printf("WAN IP has not changed since last run, exiting.")
		return
	}

	// Update the Tunnel config
	if *noopMode {
		return
	}
	err = tunnelBrokerUpdate(wanIP)
	if err != nil {
		log.Printf("%s", err)
	}

	// Save the new IP address since we successfully updated the tunnel config
	log.Printf("Saving new WAN IP to cache file %s", *cacheFile)
	saveNewWANIP(wanIP, *cacheFile)
}

func UPNPListener(conn *net.UDPConn, decodeChan chan []byte) {
	for {
		buf := make([]byte, 512)
		_, err := conn.Read(buf)
		if err != nil {
			log.Printf("error receiving UDP packet: %s", err.Error())
			break
		}

		decodeChan <- buf
	}
}

// Generate a discovery message that restricts the search target to
// only WANIPConnection:1 services.
func ssdpQueryMessage(timeoutDuration time.Duration) *bytes.Buffer {
	msg := &bytes.Buffer{}
	msg.WriteString("M-SEARCH * HTTP/1.1\r\n")
	msg.WriteString("Host: 239.255.255.250:1900\r\n")
	msg.WriteString("MAN: \"ssdp:discover\"\r\n")
	msg.WriteString(fmt.Sprintf("MX: %d\r\n", int(timeoutDuration.Seconds())))
	msg.WriteString("ST: urn:schemas-upnp-org:service:WANIPConnection:1\r\n")
	msg.WriteString("\r\n")
	return msg
}

func msgDecoder(decodeChan chan []byte, ssdpChan chan map[string]string) {
	for packet := range decodeChan {
		ssdpResponse := make(map[string]string)
		// Response lines are separated by CR+LF as in requests
		lines := strings.Split(string(packet), "\r\n")

		for i := range lines {
			// Throw away empty lines
			if lines[i] == "" {
				continue
			}

			// Capture status separately
			if len(lines[i]) > 4 && lines[i][:4] == "HTTP" {
				ssdpResponse["Status"] = lines[i]
			} else {
				// Split header:value from line
				fields := strings.SplitN(lines[i], ":", 2)

				// Guard against headers with no content
				if len(fields) < 2 {
					continue
				}

				ssdpResponse[fields[0]] = strings.TrimSpace(fields[1])
			}
		}

		ssdpChan <- ssdpResponse
	}
}

func hasCurrentIPChanged(wanIP, filename string) bool {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return true
	}

	return !(string(contents) == wanIP)
}

func saveNewWANIP(wanIP, filename string) {
	_ = ioutil.WriteFile(filename, []byte(wanIP), 0600)
}
