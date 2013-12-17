package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
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
	query := ssdpQueryMessage(10)
	_, err = udp.WriteToUDP(query.Bytes(), ssdp)
	if err != nil {
		log.Fatalf("error during UPNP discovery message: %s", err.Error())
	}

	// TODO: Retry after timeout period if no response received.
	log.Println("sent SSDP discovery message")

	// Only interested in the first message
	msg := <-ssdpChan
	close(ssdpChan)
	controlPoint, err := getWANControlPoint(msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("received control point: %s", controlPoint)

	// Hit the control point
	log.Printf("Current WAN IP is: %s", retrieveWANIP(controlPoint))
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
func ssdpQueryMessage(timeout int) *bytes.Buffer {
	msg := &bytes.Buffer{}
	msg.WriteString("M-SEARCH * HTTP/1.1\r\n")
	msg.WriteString("Host: 239.255.255.250:1900\r\n")
	msg.WriteString("MAN: \"ssdp:discover\"\r\n")
	msg.WriteString(fmt.Sprintf("MX: %d\r\n", timeout))
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
